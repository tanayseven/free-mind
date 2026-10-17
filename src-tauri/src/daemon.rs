//! Privileged helper: the app binary started with `--daemon` by launchd,
//! systemd or the Windows service manager. It owns all hosts file writes so the
//! app never needs to ask for a password after the one-time install.

use crate::hosts;
use crate::ipc::{self, Request, Response};
use interprocess::local_socket::{prelude::*, Listener, ListenerOptions, Stream};
use std::io::{self, BufReader};
use std::path::Path;
use std::sync::Mutex;
use std::thread;

/// Serializes hosts file edits across concurrent connections.
static HOSTS_LOCK: Mutex<()> = Mutex::new(());

/// Applies one request to the hosts file at `hosts_path`. The daemon is the
/// trust boundary, so domains are re-validated here.
pub fn handle(request: Request, hosts_path: &Path) -> Response {
    let _guard = HOSTS_LOCK.lock().unwrap_or_else(|e| e.into_inner());
    let result = match request {
        Request::Ping => return Response::ok(),
        Request::Block { domains } => {
            if let Some(bad) = domains.iter().find(|d| !hosts::is_valid_domain(d)) {
                return Response::err(format!("invalid domain: {bad:?}"));
            }
            hosts::block(hosts_path, &domains)
        }
        Request::Unblock => hosts::unblock(hosts_path),
    };
    match result {
        Ok(()) => Response::ok(),
        Err(e) => Response::err(format!("failed to update hosts file: {e}")),
    }
}

/// Entry point for `--daemon`. `args` are the process arguments.
pub fn run(args: &[String]) -> io::Result<()> {
    #[cfg(windows)]
    {
        let _ = args;
        windows_service_main::run()
    }
    #[cfg(unix)]
    {
        let owner_uid = args
            .iter()
            .position(|a| a == "--owner-uid")
            .and_then(|i| args.get(i + 1))
            .and_then(|uid| uid.parse().ok())
            .ok_or_else(|| io::Error::other("--owner-uid <uid> is required"))?;
        serve(listen(owner_uid)?)
    }
}

/// Creates the socket, accessible only to root and the installing user.
#[cfg(unix)]
fn listen(owner_uid: u32) -> io::Result<Listener> {
    use std::os::unix::fs::PermissionsExt;

    // interprocess' `mode()` is unsupported on macOS, so create the socket
    // owner-only via the umask (no window where others can connect), then
    // hand it to the installing user.
    // SAFETY: umask has no preconditions; nothing else runs concurrently yet.
    unsafe { libc::umask(0o077) };
    let listener = ListenerOptions::new()
        .name(ipc::socket_name()?)
        .try_overwrite(true)
        .create_sync()?;
    std::os::unix::fs::chown(ipc::SOCKET_PATH, Some(owner_uid), None)?;
    std::fs::set_permissions(ipc::SOCKET_PATH, std::fs::Permissions::from_mode(0o600))?;
    Ok(listener)
}

/// Creates the pipe, readable and writable by any signed-in user.
#[cfg(windows)]
fn listen() -> io::Result<Listener> {
    use interprocess::os::windows::{
        local_socket::ListenerOptionsExt, security_descriptor::SecurityDescriptor,
    };
    use widestring::U16CString;

    let sddl = U16CString::from_str("D:(A;;GA;;;SY)(A;;GA;;;BA)(A;;GRGW;;;AU)")
        .map_err(io::Error::other)?;
    ListenerOptions::new()
        .name(ipc::socket_name()?)
        .security_descriptor(SecurityDescriptor::deserialize(&sddl)?)
        .create_sync()
}

fn serve(listener: Listener) -> io::Result<()> {
    eprintln!("free-mind daemon listening");
    for conn in listener.incoming() {
        match conn {
            Ok(conn) => {
                thread::spawn(move || {
                    if let Err(e) = handle_connection(conn) {
                        eprintln!("connection error: {e}");
                    }
                });
            }
            Err(e) => eprintln!("incoming connection failed: {e}"),
        }
    }
    Ok(())
}

fn handle_connection(conn: Stream) -> io::Result<()> {
    let mut conn = BufReader::new(conn);
    loop {
        let request: Request = match ipc::read_message(&mut conn) {
            Ok(request) => request,
            Err(e) if e.kind() == io::ErrorKind::UnexpectedEof => return Ok(()),
            Err(e) => {
                ipc::write_message(conn.get_mut(), &Response::err(e.to_string()))?;
                return Ok(());
            }
        };
        let changes_hosts = !matches!(request, Request::Ping);
        let response = handle(request, Path::new(hosts::HOSTS_PATH));
        if response.ok && changes_hosts {
            hosts::flush_dns();
        }
        ipc::write_message(conn.get_mut(), &response)?;
    }
}

/// Windows service glue: reports to the service manager and serves the pipe
/// on a background thread until asked to stop.
#[cfg(windows)]
mod windows_service_main {
    use super::{listen, serve};
    use std::ffi::OsString;
    use std::io;
    use std::sync::mpsc;
    use std::time::Duration;
    use windows_service::service::{
        ServiceControl, ServiceControlAccept, ServiceExitCode, ServiceState, ServiceStatus,
        ServiceType,
    };
    use windows_service::service_control_handler::{self, ServiceControlHandlerResult};
    use windows_service::{define_windows_service, service_dispatcher};

    pub const SERVICE_NAME: &str = "FreeMindDaemon";

    define_windows_service!(ffi_service_main, service_main);

    pub fn run() -> io::Result<()> {
        // Fails when started from a console rather than by the service
        // manager; serve directly then, which is handy for debugging.
        if service_dispatcher::start(SERVICE_NAME, ffi_service_main).is_err() {
            return serve(listen()?);
        }
        Ok(())
    }

    fn service_main(_args: Vec<OsString>) {
        if let Err(e) = run_service() {
            eprintln!("service error: {e}");
        }
    }

    fn status(state: ServiceState, accept: ServiceControlAccept) -> ServiceStatus {
        ServiceStatus {
            service_type: ServiceType::OWN_PROCESS,
            current_state: state,
            controls_accepted: accept,
            exit_code: ServiceExitCode::Win32(0),
            checkpoint: 0,
            wait_hint: Duration::default(),
            process_id: None,
        }
    }

    fn run_service() -> Result<(), Box<dyn std::error::Error>> {
        let (stop_tx, stop_rx) = mpsc::channel();
        let handle = service_control_handler::register(SERVICE_NAME, move |event| match event {
            ServiceControl::Stop | ServiceControl::Shutdown => {
                let _ = stop_tx.send(());
                ServiceControlHandlerResult::NoError
            }
            ServiceControl::Interrogate => ServiceControlHandlerResult::NoError,
            _ => ServiceControlHandlerResult::NotImplemented,
        })?;

        let listener = listen()?;
        std::thread::spawn(move || serve(listener));
        handle.set_service_status(status(
            ServiceState::Running,
            ServiceControlAccept::STOP | ServiceControlAccept::SHUTDOWN,
        ))?;

        let _ = stop_rx.recv();
        handle.set_service_status(status(ServiceState::Stopped, ServiceControlAccept::empty()))?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    fn temp_hosts(name: &str) -> std::path::PathBuf {
        let path =
            std::env::temp_dir().join(format!("free-mind-test-{name}-{}", std::process::id()));
        fs::write(&path, "127.0.0.1 localhost\n").unwrap();
        path
    }

    #[test]
    fn block_then_unblock_round_trips_hosts_file() {
        let path = temp_hosts("roundtrip");
        let block = Request::Block {
            domains: vec!["x.com".into()],
        };
        assert_eq!(handle(block, &path), Response::ok());
        assert!(fs::read_to_string(&path).unwrap().contains("0.0.0.0 x.com"));

        assert_eq!(handle(Request::Unblock, &path), Response::ok());
        assert_eq!(fs::read_to_string(&path).unwrap(), "127.0.0.1 localhost\n");
        fs::remove_file(path).unwrap();
    }

    #[test]
    fn rejects_invalid_domains_without_touching_hosts_file() {
        let path = temp_hosts("invalid");
        let block = Request::Block {
            domains: vec!["ok.com".into(), "evil.com\n6.6.6.6 bank.com".into()],
        };
        let response = handle(block, &path);
        assert!(!response.ok);
        assert!(response.error.unwrap().contains("invalid domain"));
        assert_eq!(fs::read_to_string(&path).unwrap(), "127.0.0.1 localhost\n");
        fs::remove_file(path).unwrap();
    }
}

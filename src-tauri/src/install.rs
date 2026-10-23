//! One-time daemon install: copies the running executable to an admin-only
//! location and registers it as a boot-time service, behind a single
//! elevation prompt. A root service must never run a user-writable binary,
//! hence the copy.

use std::fs;
use std::io;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};
use std::time::{Duration, Instant};

use crate::ipc::{self, Request};

#[cfg(target_os = "macos")]
const INSTALL_PATH: &str = "/Library/PrivilegedHelperTools/tech.tanay.free-mind.daemon";
#[cfg(target_os = "macos")]
const LAUNCHD_LABEL: &str = "tech.tanay.free-mind.daemon";

#[cfg(all(unix, not(target_os = "macos")))]
const INSTALL_PATH: &str = "/usr/local/libexec/free-mind/free-mind-daemon";
/// Service name for systemd, OpenRC, runit and Slackware's rc scripts.
#[cfg(target_os = "linux")]
const SERVICE: &str = "free-mind-daemon";
/// Written by package-managed installs (the NixOS module) that run the daemon
/// themselves, so the app never tries to install its own copy.
#[cfg(target_os = "linux")]
const MANAGED_MARKER: &str = "/etc/free-mind/managed-daemon";

#[cfg(target_os = "freebsd")]
const RC_NAME: &str = "free_mind_daemon";

#[cfg(windows)]
const INSTALL_PATH: &str = r"C:\Program Files\Free Mind\free-mind-daemon.exe";
#[cfg(windows)]
const SERVICE_NAME: &str = "FreeMindDaemon";

fn is_managed() -> bool {
    #[cfg(target_os = "linux")]
    return Path::new(MANAGED_MARKER).exists();
    #[cfg(not(target_os = "linux"))]
    false
}

pub fn is_installed() -> bool {
    is_managed() || Path::new(INSTALL_PATH).exists()
}

/// The file the daemon is copied from. Inside an AppImage that's the AppImage
/// itself: the inner binary sits on a FUSE mount root can't read, and needs
/// the libraries bundled next to it.
fn daemon_source() -> io::Result<PathBuf> {
    #[cfg(target_os = "linux")]
    if let Some(appimage) = std::env::var_os("APPIMAGE") {
        return Ok(PathBuf::from(appimage));
    }
    std::env::current_exe()
}

/// Whether the installed daemon is byte-identical to the running app, so app
/// updates (and dev rebuilds) trigger a reinstall.
pub fn is_current() -> bool {
    if is_managed() {
        return true;
    }
    let Ok(exe) = daemon_source() else {
        return false;
    };
    match (fs::metadata(&exe), fs::metadata(INSTALL_PATH)) {
        (Ok(a), Ok(b)) if a.len() == b.len() => {
            matches!((fs::read(&exe), fs::read(INSTALL_PATH)), (Ok(a), Ok(b)) if a == b)
        }
        _ => false,
    }
}

/// Installs (or reinstalls) and starts the daemon, then waits for it to answer.
pub fn install() -> io::Result<()> {
    let exe = daemon_source()?;
    let dir = private_temp_dir()?;
    let result = run_elevated_install(&exe, &dir);
    let _ = fs::remove_dir_all(&dir);
    result?;

    let deadline = Instant::now() + Duration::from_secs(10);
    loop {
        match ipc::send(&Request::Ping) {
            Ok(()) => return Ok(()),
            Err(e) if Instant::now() >= deadline => {
                return Err(io::Error::other(format!(
                    "daemon installed but not responding: {e}"
                )))
            }
            Err(_) => std::thread::sleep(Duration::from_millis(200)),
        }
    }
}

/// A fresh directory only the current user can write, for the files the
/// elevated script reads (so no other user can swap them mid-install).
fn private_temp_dir() -> io::Result<PathBuf> {
    let dir = std::env::temp_dir().join(format!("free-mind-install-{}", std::process::id()));
    let _ = fs::remove_dir_all(&dir);
    #[cfg_attr(not(unix), allow(unused_mut))]
    let mut builder = fs::DirBuilder::new();
    #[cfg(unix)]
    std::os::unix::fs::DirBuilderExt::mode(&mut builder, 0o700);
    builder.create(&dir)?;
    Ok(dir)
}

#[cfg(unix)]
fn sh_quote(s: &str) -> String {
    format!("'{}'", s.replace('\'', r"'\''"))
}

#[cfg(unix)]
fn current_uid() -> u32 {
    // SAFETY: getuid has no preconditions and cannot fail.
    unsafe { libc::getuid() }
}

#[cfg(target_os = "macos")]
fn run_elevated_install(exe: &Path, dir: &Path) -> io::Result<()> {
    let plist_path = format!("/Library/LaunchDaemons/{LAUNCHD_LABEL}.plist");
    // Keep in sync with packaging/macos-pkg/postinstall.
    let plist = format!(
        r#"<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key><string>{LAUNCHD_LABEL}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{INSTALL_PATH}</string>
        <string>--daemon</string>
        <string>--owner-uid</string>
        <string>{uid}</string>
    </array>
    <key>RunAtLoad</key><true/>
    <key>KeepAlive</key><true/>
    <key>StandardErrorPath</key><string>/var/log/free-mind-daemon.log</string>
</dict>
</plist>
"#,
        uid = current_uid()
    );
    let plist_tmp = dir.join("daemon.plist");
    fs::write(&plist_tmp, plist)?;

    let script = format!(
        "set -e
launchctl bootout system/{LAUNCHD_LABEL} 2>/dev/null || true
mkdir -p /Library/PrivilegedHelperTools
cp {exe} {dest}.new
chown root:wheel {dest}.new
chmod 755 {dest}.new
mv -f {dest}.new {dest}
cp {plist_tmp} {plist}
chown root:wheel {plist}
chmod 644 {plist}
# bootstrap can fail briefly while the old job is still being torn down.
n=0
until launchctl bootstrap system {plist}; do
    n=$((n+1)); [ $n -ge 5 ] && exit 1; sleep 1
done
",
        exe = sh_quote(&exe.to_string_lossy()),
        dest = INSTALL_PATH,
        plist_tmp = sh_quote(&plist_tmp.to_string_lossy()),
        plist = plist_path,
    );
    let script_path = dir.join("install.sh");
    fs::write(&script_path, script)?;

    // AppleScript string literal: escape backslashes and double quotes.
    let command = format!("/bin/sh {}", sh_quote(&script_path.to_string_lossy()))
        .replace('\\', "\\\\")
        .replace('"', "\\\"");
    let apple_script = format!(
        "do shell script \"{command}\" with prompt \"Free Mind needs to install a helper to block websites.\" with administrator privileges"
    );
    check(
        Command::new("osascript")
            .args(["-e", &apple_script])
            .output()?,
    )
}

// Every Linux service definition sets APPIMAGE_EXTRACT_AND_RUN so an AppImage
// daemon runs as root without FUSE; plain binaries ignore it.

#[cfg(target_os = "linux")]
fn systemd_unit(uid: u32) -> String {
    format!(
        "[Unit]
Description=Free Mind website blocking helper

[Service]
ExecStart={INSTALL_PATH} --daemon --owner-uid {uid}
Environment=APPIMAGE_EXTRACT_AND_RUN=1
Restart=always

[Install]
WantedBy=multi-user.target
"
    )
}

#[cfg(target_os = "linux")]
fn openrc_script(uid: u32) -> String {
    format!(
        "#!/sbin/openrc-run
description=\"Free Mind website blocking helper\"
supervisor=supervise-daemon
command={INSTALL_PATH}
command_args=\"--daemon --owner-uid {uid}\"
export APPIMAGE_EXTRACT_AND_RUN=1

depend() {{
    after localmount
}}
"
    )
}

#[cfg(target_os = "linux")]
fn runit_run(uid: u32) -> String {
    format!(
        "#!/bin/sh
export APPIMAGE_EXTRACT_AND_RUN=1
exec {INSTALL_PATH} --daemon --owner-uid {uid} 2>&1
"
    )
}

/// Slackware has no supervisor, so this is a plain start/stop script.
#[cfg(target_os = "linux")]
fn slackware_rc(uid: u32) -> String {
    format!(
        "#!/bin/sh
# Free Mind website blocking helper
PIDFILE=/var/run/{SERVICE}.pid
case \"$1\" in
start)
    APPIMAGE_EXTRACT_AND_RUN=1 nohup {INSTALL_PATH} --daemon --owner-uid {uid} \\
        </dev/null >>/var/log/{SERVICE}.log 2>&1 &
    echo $! >$PIDFILE
    ;;
stop)
    [ -f $PIDFILE ] && kill \"$(cat $PIDFILE)\" 2>/dev/null
    rm -f $PIDFILE
    ;;
restart)
    \"$0\" stop
    sleep 1
    \"$0\" start
    ;;
*)
    echo \"usage: $0 start|stop|restart\"
    ;;
esac
"
    )
}

/// Installs the daemon under whichever init system is running: systemd,
/// OpenRC (Alpine, Gentoo), runit (Void) or Slackware's rc scripts.
#[cfg(target_os = "linux")]
fn run_elevated_install(exe: &Path, dir: &Path) -> io::Result<()> {
    let uid = current_uid();
    let files = [
        ("systemd", systemd_unit(uid)),
        ("openrc", openrc_script(uid)),
        ("runit", runit_run(uid)),
        ("slackware", slackware_rc(uid)),
    ];
    for (name, content) in &files {
        fs::write(dir.join(name), content)?;
    }
    run_pkexec(dir, &linux_install_script(exe, dir))
}

/// The elevated script: installs the binary, then the service file for the
/// running init system from the ones written to `dir`.
#[cfg(target_os = "linux")]
fn linux_install_script(exe: &Path, dir: &Path) -> String {
    let tmp = |name: &str| sh_quote(&dir.join(name).to_string_lossy());
    let install_dir = Path::new(INSTALL_PATH).parent().unwrap().display();
    format!(
        "set -e
mkdir -p {install_dir}
install -m 755 -o root -g root {exe} {dest}.new
mv -f {dest}.new {dest}
if [ -d /run/systemd/system ]; then
    install -m 644 -o root -g root {systemd} /etc/systemd/system/{SERVICE}.service
    systemctl daemon-reload
    systemctl enable {SERVICE}.service
    systemctl restart {SERVICE}.service
elif command -v openrc-run >/dev/null; then
    install -m 755 -o root -g root {openrc} /etc/init.d/{SERVICE}
    rc-update add {SERVICE} default
    rc-service {SERVICE} restart
elif [ -d /etc/sv ] && [ -d /var/service ]; then
    mkdir -p /etc/sv/{SERVICE}
    install -m 755 -o root -g root {runit} /etc/sv/{SERVICE}/run
    if [ -e /var/service/{SERVICE} ]; then
        sv restart {SERVICE}
    else
        ln -s /etc/sv/{SERVICE} /var/service/
    fi
elif [ -f /etc/rc.d/rc.M ]; then
    install -m 755 -o root -g root {slackware} /etc/rc.d/rc.{SERVICE}
    grep -q 'rc.{SERVICE}' /etc/rc.d/rc.local 2>/dev/null ||
        echo '[ -x /etc/rc.d/rc.{SERVICE} ] && /etc/rc.d/rc.{SERVICE} start' >>/etc/rc.d/rc.local
    /etc/rc.d/rc.{SERVICE} restart
else
    echo 'Free Mind needs systemd, OpenRC, runit or Slackware init to run its helper' >&2
    exit 1
fi
",
        exe = sh_quote(&exe.to_string_lossy()),
        dest = INSTALL_PATH,
        systemd = tmp("systemd"),
        openrc = tmp("openrc"),
        runit = tmp("runit"),
        slackware = tmp("slackware"),
    )
}

#[cfg(target_os = "freebsd")]
fn rc_script(uid: u32) -> String {
    format!(
        "#!/bin/sh
# PROVIDE: {RC_NAME}
# REQUIRE: LOGIN
# KEYWORD: shutdown

. /etc/rc.subr

name={RC_NAME}
rcvar={RC_NAME}_enable
pidfile=/var/run/{RC_NAME}.pid
command=/usr/sbin/daemon
command_args=\"-r -f -P $pidfile {INSTALL_PATH} --daemon --owner-uid {uid}\"

load_rc_config $name
run_rc_command \"$1\"
"
    )
}

#[cfg(target_os = "freebsd")]
fn run_elevated_install(exe: &Path, dir: &Path) -> io::Result<()> {
    let rc_tmp = dir.join(RC_NAME);
    fs::write(&rc_tmp, rc_script(current_uid()))?;

    let install_dir = Path::new(INSTALL_PATH).parent().unwrap().display();
    let script = format!(
        "set -e
mkdir -p {install_dir}
install -m 755 -o root -g wheel {exe} {dest}.new
mv -f {dest}.new {dest}
install -m 555 -o root -g wheel {rc_tmp} /usr/local/etc/rc.d/{RC_NAME}
sysrc {RC_NAME}_enable=YES
service {RC_NAME} stop 2>/dev/null || true
service {RC_NAME} start
",
        exe = sh_quote(&exe.to_string_lossy()),
        dest = INSTALL_PATH,
        rc_tmp = sh_quote(&rc_tmp.to_string_lossy()),
    );
    run_pkexec(dir, &script)
}

#[cfg(all(unix, not(target_os = "macos")))]
fn run_pkexec(dir: &Path, script: &str) -> io::Result<()> {
    let script_path = dir.join("install.sh");
    fs::write(&script_path, script)?;
    check(
        Command::new("pkexec")
            .arg("/bin/sh")
            .arg(&script_path)
            .output()?,
    )
}

#[cfg(windows)]
fn run_elevated_install(exe: &Path, dir: &Path) -> io::Result<()> {
    fn ps_quote(s: &str) -> String {
        format!("'{}'", s.replace('\'', "''"))
    }

    let install_dir = Path::new(INSTALL_PATH).parent().unwrap();
    let script = format!(
        "$ErrorActionPreference = 'Stop'
$svc = Get-Service {SERVICE_NAME} -ErrorAction SilentlyContinue
if ($svc -and $svc.Status -ne 'Stopped') {{
    Stop-Service {SERVICE_NAME} -Force
    $svc.WaitForStatus('Stopped', '00:00:15')
}}
New-Item -ItemType Directory -Force {dir} | Out-Null
Copy-Item -Force {exe} {dest}
if (-not $svc) {{
    New-Service -Name {SERVICE_NAME} -DisplayName 'Free Mind Daemon' -StartupType Automatic `
        -BinaryPathName ('\"' + {dest} + '\" --daemon') | Out-Null
}}
Start-Service {SERVICE_NAME}
",
        dir = ps_quote(&install_dir.to_string_lossy()),
        exe = ps_quote(&exe.to_string_lossy()),
        dest = ps_quote(INSTALL_PATH),
    );
    let script_path = dir.join("install.ps1");
    fs::write(&script_path, script)?;

    // Start-Process -Verb RunAs shows the UAC prompt; relay the exit code.
    let launcher = format!(
        "$p = Start-Process powershell -Verb RunAs -Wait -PassThru -WindowStyle Hidden \
         -ArgumentList '-NoProfile','-ExecutionPolicy','Bypass','-File',{}; exit $p.ExitCode",
        ps_quote(&format!("\"{}\"", script_path.display()))
    );
    check(
        Command::new("powershell")
            .args(["-NoProfile", "-Command", &launcher])
            .output()?,
    )
}

fn check(output: Output) -> io::Result<()> {
    if output.status.success() {
        return Ok(());
    }
    let stderr = String::from_utf8_lossy(&output.stderr).trim().to_string();
    Err(io::Error::other(if stderr.is_empty() {
        format!("installer exited with {}", output.status)
    } else {
        stderr
    }))
}

#[cfg(all(test, unix))]
mod tests {
    use super::*;

    #[test]
    fn sh_quote_handles_single_quotes() {
        assert_eq!(sh_quote("/Users/o'brien/app"), r"'/Users/o'\''brien/app'");
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn linux_services_run_the_daemon_for_the_installing_user() {
        for service in [
            systemd_unit(1000),
            openrc_script(1000),
            runit_run(1000),
            slackware_rc(1000),
        ] {
            assert!(service.contains(INSTALL_PATH), "{service}");
            assert!(service.contains("--daemon --owner-uid 1000"), "{service}");
            assert!(service.contains("APPIMAGE_EXTRACT_AND_RUN=1"), "{service}");
        }
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn linux_install_script_is_valid_sh() {
        let script =
            linux_install_script(Path::new("/home/o'brien/free-mind"), Path::new("/tmp/a b"));
        let status = Command::new("sh")
            .args(["-n", "-c", &script])
            .status()
            .unwrap();
        assert!(status.success(), "{script}");
    }

    #[cfg(target_os = "linux")]
    #[test]
    fn init_scripts_have_the_right_interpreter() {
        assert!(openrc_script(1000).starts_with("#!/sbin/openrc-run\n"));
        assert!(runit_run(1000).starts_with("#!/bin/sh\n"));
        assert!(slackware_rc(1000).starts_with("#!/bin/sh\n"));
    }

    #[cfg(target_os = "freebsd")]
    #[test]
    fn rc_script_supervises_the_daemon() {
        let script = rc_script(1000);
        assert!(script.contains(&format!(
            "-r -f -P $pidfile {INSTALL_PATH} --daemon --owner-uid 1000"
        )));
        assert!(script.contains("rcvar=free_mind_daemon_enable"));
    }
}

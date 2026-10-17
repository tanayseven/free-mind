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
#[cfg(all(unix, not(target_os = "macos")))]
const SYSTEMD_UNIT: &str = "free-mind-daemon.service";

#[cfg(windows)]
const INSTALL_PATH: &str = r"C:\Program Files\Free Mind\free-mind-daemon.exe";
#[cfg(windows)]
const SERVICE_NAME: &str = "FreeMindDaemon";

pub fn is_installed() -> bool {
    Path::new(INSTALL_PATH).exists()
}

/// Whether the installed daemon is byte-identical to the running app, so app
/// updates (and dev rebuilds) trigger a reinstall.
pub fn is_current() -> bool {
    let Ok(exe) = std::env::current_exe() else {
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
    let exe = std::env::current_exe()?;
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

#[cfg(all(unix, not(target_os = "macos")))]
fn run_elevated_install(exe: &Path, dir: &Path) -> io::Result<()> {
    let unit = format!(
        "[Unit]
Description=Free Mind website blocking helper

[Service]
ExecStart={INSTALL_PATH} --daemon --owner-uid {uid}
Restart=always

[Install]
WantedBy=multi-user.target
",
        uid = current_uid()
    );
    let unit_tmp = dir.join(SYSTEMD_UNIT);
    fs::write(&unit_tmp, unit)?;

    let install_dir = Path::new(INSTALL_PATH).parent().unwrap().display();
    let script = format!(
        "set -e
command -v systemctl >/dev/null || {{ echo 'Free Mind needs systemd to run its helper' >&2; exit 1; }}
systemctl stop {SYSTEMD_UNIT} 2>/dev/null || true
mkdir -p {install_dir}
install -m 755 -o root -g root {exe} {dest}.new
mv -f {dest}.new {dest}
install -m 644 -o root -g root {unit_tmp} /etc/systemd/system/{SYSTEMD_UNIT}
systemctl daemon-reload
systemctl enable --now {SYSTEMD_UNIT}
",
        exe = sh_quote(&exe.to_string_lossy()),
        dest = INSTALL_PATH,
        unit_tmp = sh_quote(&unit_tmp.to_string_lossy()),
    );
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
}

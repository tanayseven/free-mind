//! Website blocking via the system hosts file.
//!
//! Blocked domains are written between marker lines so they can be detected and
//! removed without touching the rest of the file. Writing the hosts file needs
//! elevated privileges, obtained per write through the OS's native prompt.

use std::fs;
use std::io;

const BEGIN_MARKER: &str = "# BEGIN Free Mind block list";
const END_MARKER: &str = "# END Free Mind block list";

#[cfg(windows)]
const HOSTS_PATH: &str = r"C:\Windows\System32\drivers\etc\hosts";
#[cfg(not(windows))]
const HOSTS_PATH: &str = "/etc/hosts";

/// Accepts plain hostnames only, so user input can't inject extra hosts entries.
pub fn is_valid_domain(domain: &str) -> bool {
    !domain.is_empty()
        && domain.len() <= 253
        && domain.contains('.')
        && domain
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || c == '.' || c == '-')
        && !domain.starts_with(['.', '-'])
}

/// Returns `content` with any Free Mind block removed.
fn strip_block(content: &str) -> String {
    let mut out = String::with_capacity(content.len());
    let mut inside = false;
    for line in content.lines() {
        match line.trim() {
            BEGIN_MARKER => inside = true,
            END_MARKER => inside = false,
            _ if !inside => {
                out.push_str(line);
                out.push('\n');
            }
            _ => {}
        }
    }
    // Drop the blank separator line `with_block` adds, so toggling is stable.
    match out.trim_end() {
        "" => String::new(),
        trimmed => format!("{trimmed}\n"),
    }
}

/// Returns `content` with a Free Mind block for `domains` appended.
fn with_block(content: &str, domains: &[String]) -> String {
    let mut out = strip_block(content);
    if !out.is_empty() {
        out.push('\n');
    }
    out.push_str(BEGIN_MARKER);
    out.push('\n');
    for domain in domains {
        out.push_str(&format!("0.0.0.0 {domain}\n:: {domain}\n"));
    }
    out.push_str(END_MARKER);
    out.push('\n');
    out
}

fn has_block(content: &str) -> bool {
    content.lines().any(|line| line.trim() == BEGIN_MARKER)
}

pub fn is_blocking() -> bool {
    fs::read_to_string(HOSTS_PATH).is_ok_and(|c| has_block(&c))
}

pub fn block(domains: &[String]) -> io::Result<()> {
    let current = fs::read_to_string(HOSTS_PATH)?;
    write_hosts(&with_block(&current, domains))
}

pub fn unblock() -> io::Result<()> {
    let current = fs::read_to_string(HOSTS_PATH)?;
    if !has_block(&current) {
        return Ok(());
    }
    write_hosts(&strip_block(&current))
}

/// Copies a per-user temp file over the hosts file via an admin prompt, then
/// flushes the DNS cache.
#[cfg(target_os = "macos")]
fn write_hosts(content: &str) -> io::Result<()> {
    use std::process::Command;

    // $TMPDIR on macOS is per-user, so other users can't swap the file.
    let tmp = std::env::temp_dir().join("free-mind-hosts");
    fs::write(&tmp, content)?;
    let script = format!(
        "do shell script \"cat '{}' > {HOSTS_PATH} && dscacheutil -flushcache; killall -HUP mDNSResponder\" with administrator privileges",
        tmp.display()
    );
    let output = Command::new("osascript").args(["-e", &script]).output();
    let _ = fs::remove_file(&tmp);
    check_output(output?)
}

/// Pipes the new content to `tee` through polkit's admin prompt.
#[cfg(all(unix, not(target_os = "macos")))]
fn write_hosts(content: &str) -> io::Result<()> {
    use std::io::Write;
    use std::process::{Command, Stdio};

    let mut child = Command::new("pkexec")
        .args(["tee", HOSTS_PATH])
        .stdin(Stdio::piped())
        .stdout(Stdio::null())
        .stderr(Stdio::piped())
        .spawn()?;
    child
        .stdin
        .take()
        .expect("stdin is piped")
        .write_all(content.as_bytes())?;
    check_output(child.wait_with_output()?)?;
    // Best effort: only systemd-resolved caches by default.
    let _ = Command::new("resolvectl").arg("flush-caches").output();
    Ok(())
}

/// Writes directly; requires the app to be running as Administrator.
#[cfg(windows)]
fn write_hosts(content: &str) -> io::Result<()> {
    fs::write(HOSTS_PATH, content)?;
    let _ = std::process::Command::new("ipconfig")
        .arg("/flushdns")
        .output();
    Ok(())
}

#[cfg(unix)]
fn check_output(output: std::process::Output) -> io::Result<()> {
    if output.status.success() {
        Ok(())
    } else {
        Err(io::Error::other(
            String::from_utf8_lossy(&output.stderr).trim().to_string(),
        ))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const ORIGINAL: &str = "127.0.0.1 localhost\n::1 localhost\n";

    #[test]
    fn adds_block_after_existing_entries() {
        let result = with_block(ORIGINAL, &["x.com".to_string()]);
        assert_eq!(
            result,
            "127.0.0.1 localhost\n::1 localhost\n\n# BEGIN Free Mind block list\n0.0.0.0 x.com\n:: x.com\n# END Free Mind block list\n"
        );
        assert!(has_block(&result));
    }

    #[test]
    fn replaces_existing_block_instead_of_duplicating() {
        let once = with_block(ORIGINAL, &["x.com".to_string()]);
        let twice = with_block(&once, &["reddit.com".to_string()]);
        assert_eq!(twice.matches(BEGIN_MARKER).count(), 1);
        assert!(!twice.contains("x.com"));
        assert!(twice.contains("0.0.0.0 reddit.com"));
    }

    #[test]
    fn strip_restores_original() {
        let blocked = with_block(ORIGINAL, &["x.com".to_string()]);
        assert_eq!(strip_block(&blocked), ORIGINAL);
        assert!(!has_block(&strip_block(&blocked)));
    }

    #[test]
    fn rejects_domains_that_could_inject_entries() {
        assert!(is_valid_domain("www.youtube.com"));
        assert!(!is_valid_domain("evil.com\n1.2.3.4 bank.com"));
        assert!(!is_valid_domain("a.com b.com"));
        assert!(!is_valid_domain("localhost"));
        assert!(!is_valid_domain(""));
    }
}

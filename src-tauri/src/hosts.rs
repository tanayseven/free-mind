//! Website blocking via the system hosts file.
//!
//! Blocked domains are written between marker lines so they can be detected and
//! removed without touching the rest of the file. Writing the hosts file needs
//! elevated privileges, so writes happen in the daemon (see `daemon`).

use std::fs;
use std::io;
use std::path::Path;

const BEGIN_MARKER: &str = "# BEGIN Free Mind block list";
const END_MARKER: &str = "# END Free Mind block list";

#[cfg(windows)]
pub const HOSTS_PATH: &str = r"C:\Windows\System32\drivers\etc\hosts";
#[cfg(not(windows))]
pub const HOSTS_PATH: &str = "/etc/hosts";

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

/// Writes a block for `domains` into the hosts file at `path`. Needs root /
/// Administrator, so in the app this only runs inside the daemon.
pub fn block(path: &Path, domains: &[String]) -> io::Result<()> {
    let current = fs::read_to_string(path)?;
    fs::write(path, with_block(&current, domains))
}

/// Removes the block from the hosts file at `path`, if present.
pub fn unblock(path: &Path) -> io::Result<()> {
    let current = fs::read_to_string(path)?;
    if !has_block(&current) {
        return Ok(());
    }
    fs::write(path, strip_block(&current))
}

/// Best-effort DNS cache flush so hosts changes apply immediately.
pub fn flush_dns() {
    use std::process::Command;

    #[cfg(target_os = "macos")]
    {
        let _ = Command::new("dscacheutil").arg("-flushcache").output();
        let _ = Command::new("killall")
            .args(["-HUP", "mDNSResponder"])
            .output();
    }
    // Only systemd-resolved caches by default.
    #[cfg(all(unix, not(target_os = "macos")))]
    let _ = Command::new("resolvectl").arg("flush-caches").output();
    #[cfg(windows)]
    let _ = Command::new("ipconfig").arg("/flushdns").output();
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

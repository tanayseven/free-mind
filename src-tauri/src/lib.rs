pub mod daemon;
mod hosts;
mod install;
mod ipc;
mod tray;

use ipc::Request;
use serde::{Deserialize, Serialize};
use std::sync::Mutex;
use tauri::AppHandle;

// NOTE: Blocking goes through the privileged daemon (see `daemon`), which edits
// the hosts file. Persistence commands are still in-memory stubs.

// Domains from the last `send_block_list`, applied by `start_blocking`.
static BLOCK_LIST: Mutex<Vec<String>> = Mutex::new(Vec::new());

// In-memory website list (JSON); `None` until the frontend saves its first edit.
static WEBSITES: Mutex<Option<String>> = Mutex::new(None);

// In-memory unblock wait in seconds (stub until settings are persisted).
static UNBLOCK_WAITING: Mutex<u32> = Mutex::new(30);

const DEFAULT_WEBSITES: &[(&str, &[&str])] = &[
    (
        "Social",
        &[
            "facebook.com",
            "instagram.com",
            "x.com",
            "twitter.com",
            "tiktok.com",
            "reddit.com",
            "snapchat.com",
            "pinterest.com",
            "linkedin.com",
            "threads.net",
            "bsky.app",
            "tumblr.com",
        ],
    ),
    (
        "News",
        &[
            "news.google.com",
            "cnn.com",
            "bbc.com",
            "bbc.co.uk",
            "nytimes.com",
            "theguardian.com",
            "foxnews.com",
            "washingtonpost.com",
            "reuters.com",
            "news.ycombinator.com",
            "buzzfeed.com",
        ],
    ),
    (
        "Video",
        &[
            "youtube.com",
            "netflix.com",
            "twitch.tv",
            "primevideo.com",
            "disneyplus.com",
            "hulu.com",
            "max.com",
            "hotstar.com",
            "vimeo.com",
            "9gag.com",
        ],
    ),
];

#[derive(Serialize, Deserialize)]
struct WebsiteEntry {
    id: String,
    domain: String,
    category: String,
    enabled: bool,
}

/// The default list, one entry per site; `with_www` adds the `www.` variants
/// when blocking.
fn default_websites() -> Vec<WebsiteEntry> {
    let mut entries = Vec::new();
    for (category, domains) in DEFAULT_WEBSITES {
        for domain in *domains {
            entries.push(WebsiteEntry {
                id: (entries.len() + 1).to_string(),
                domain: domain.to_string(),
                category: category.to_string(),
                enabled: true,
            });
        }
    }
    entries
}

/// Adds a `www.` variant for every domain, since the hosts file matches exact
/// hostnames and the website list keeps one row per site. Keeps order and
/// drops duplicates.
fn with_www(domains: impl IntoIterator<Item = String>) -> Vec<String> {
    let mut out: Vec<String> = Vec::new();
    for domain in domains {
        let www = (!domain.starts_with("www.")).then(|| format!("www.{domain}"));
        for host in std::iter::once(domain).chain(www) {
            if hosts::is_valid_domain(&host) && !out.contains(&host) {
                out.push(host);
            }
        }
    }
    out
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct AppSettings {
    unblock_waiting: u32,
}

#[derive(Serialize)]
struct Environment {
    platform: String,
    arch: String,
}

// Commands that talk to the daemon or show an admin prompt are `async` so they
// run off the main thread and don't freeze the window.

/// Empty string means connected; anything containing "Error" makes the
/// frontend (re)install the daemon.
#[tauri::command(async)]
fn connect_to_daemon() -> String {
    if !install::is_current() {
        return "Error: daemon is not installed or is outdated".to_string();
    }
    match ipc::send(&Request::Ping) {
        Ok(()) => String::new(),
        Err(e) => format!("Error: {e}"),
    }
}

#[tauri::command]
fn check_blocking() -> bool {
    hosts::is_blocking()
}

#[tauri::command]
fn check_daemon_installed() -> bool {
    install::is_installed()
}

#[tauri::command(async)]
fn install_and_start_daemon() -> String {
    match install::install() {
        Ok(()) => "Daemon installed".to_string(),
        Err(e) => format!("Error installing daemon: {e}"),
    }
}

#[tauri::command]
fn send_block_list(list: String) -> bool {
    let domains: Vec<String> = list
        .split(',')
        .map(|d| d.trim().to_lowercase())
        .filter(|d| !d.is_empty())
        .collect();
    if !domains.iter().all(|d| hosts::is_valid_domain(d)) {
        return false;
    }
    *BLOCK_LIST.lock().unwrap() = with_www(domains);
    true
}

#[tauri::command(async)]
fn start_blocking() -> Result<bool, String> {
    let domains = BLOCK_LIST.lock().unwrap().clone();
    ipc::send(&Request::Block { domains })?;
    Ok(true)
}

#[tauri::command(async)]
fn stop_blocking() -> String {
    match ipc::send(&Request::Unblock) {
        Ok(()) => String::new(),
        Err(e) => format!("Error: {e}"),
    }
}

#[tauri::command]
fn load_blocked_websites() -> String {
    WEBSITES
        .lock()
        .unwrap()
        .clone()
        .unwrap_or_else(|| serde_json::to_string(&default_websites()).unwrap())
}

#[tauri::command]
fn save_blocked_websites(json: String) -> bool {
    *WEBSITES.lock().unwrap() = Some(json);
    true
}

#[tauri::command]
fn load_settings() -> AppSettings {
    AppSettings {
        unblock_waiting: *UNBLOCK_WAITING.lock().unwrap(),
    }
}

#[tauri::command]
fn save_settings(settings: AppSettings) -> bool {
    *UNBLOCK_WAITING.lock().unwrap() = settings.unblock_waiting;
    true
}

#[tauri::command]
fn set_status_icon(app: AppHandle, status: tray::Status) -> Result<(), String> {
    tray::set_status(&app, status).map_err(|e| e.to_string())
}

#[tauri::command]
fn environment() -> Environment {
    // Map Rust's OS name onto Go's runtime.GOOS values the frontend expects.
    let platform = match std::env::consts::OS {
        "macos" => "darwin".to_string(),
        other => other.to_string(),
    };
    Environment {
        platform,
        arch: std::env::consts::ARCH.to_string(),
    }
}

/// Enabled domains from the current website list, as the UI would send them.
fn enabled_domains() -> Vec<String> {
    with_www(
        serde_json::from_str::<Vec<WebsiteEntry>>(&load_blocked_websites())
            .unwrap_or_default()
            .into_iter()
            .filter(|w| w.enabled)
            .map(|w| w.domain),
    )
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .setup(|app| {
            tray::create(app.handle())?;
            Ok(())
        })
        // Closing the window hides it; the app keeps running in the tray.
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                api.prevent_close();
                let _ = window.hide();
            }
        })
        .invoke_handler(tauri::generate_handler![
            connect_to_daemon,
            check_blocking,
            check_daemon_installed,
            install_and_start_daemon,
            send_block_list,
            start_blocking,
            stop_blocking,
            load_blocked_websites,
            save_blocked_websites,
            load_settings,
            save_settings,
            set_status_icon,
            environment,
        ])
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|_app, _event| {
            // Clicking the Dock icon brings the hidden window back.
            #[cfg(target_os = "macos")]
            if let tauri::RunEvent::Reopen { .. } = _event {
                tray::show_main_window(_app);
            }
        });
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn with_www_adds_variants_once() {
        let domains = ["x.com", "www.y.com", "x.com", "bbc.co.uk"].map(String::from);
        assert_eq!(
            with_www(domains),
            [
                "x.com",
                "www.x.com",
                "www.y.com",
                "bbc.co.uk",
                "www.bbc.co.uk"
            ]
        );
    }

    #[test]
    fn default_websites_have_one_row_per_site() {
        assert!(default_websites()
            .iter()
            .all(|w| !w.domain.starts_with("www.")));
    }
}

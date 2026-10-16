mod hosts;

use serde::{Deserialize, Serialize};
use std::sync::Mutex;

// NOTE: Blocking edits the hosts file directly (see `hosts`). The daemon and
// persistence commands are still stubs returning sensible defaults.

// Domains from the last `send_block_list`, applied by `start_blocking`.
static BLOCK_LIST: Mutex<Vec<String>> = Mutex::new(Vec::new());

// In-memory website list (JSON); `None` until the frontend saves its first edit.
static WEBSITES: Mutex<Option<String>> = Mutex::new(None);

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

#[derive(Serialize)]
struct WebsiteEntry {
    id: String,
    domain: String,
    category: String,
    enabled: bool,
}

/// Builds the default list, adding a `www.` variant for apex domains since the
/// hosts file matches exact hostnames.
fn default_websites() -> Vec<WebsiteEntry> {
    let mut entries = Vec::new();
    for (category, domains) in DEFAULT_WEBSITES {
        for domain in *domains {
            let mut hosts = vec![domain.to_string()];
            if domain.matches('.').count() == 1 {
                hosts.push(format!("www.{domain}"));
            }
            for host in hosts {
                entries.push(WebsiteEntry {
                    id: (entries.len() + 1).to_string(),
                    domain: host,
                    category: category.to_string(),
                    enabled: true,
                });
            }
        }
    }
    entries
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

#[tauri::command]
fn connect_to_daemon() -> String {
    // Empty string signals "connected" to the frontend (no "Error" substring).
    String::new()
}

#[tauri::command]
fn check_blocking() -> bool {
    hosts::is_blocking()
}

#[tauri::command]
fn check_daemon_installed() -> bool {
    true
}

#[tauri::command]
fn install_and_start_daemon() -> String {
    "Daemon installed".to_string()
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
    *BLOCK_LIST.lock().unwrap() = domains;
    true
}

#[tauri::command]
fn start_blocking() -> Result<bool, String> {
    let domains = BLOCK_LIST.lock().unwrap().clone();
    hosts::block(&domains).map_err(|e| format!("Failed to update hosts file: {e}"))?;
    Ok(true)
}

#[tauri::command]
fn stop_blocking() -> String {
    match hosts::unblock() {
        Ok(()) => String::new(),
        Err(e) => format!("Error: failed to update hosts file: {e}"),
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
        unblock_waiting: 30,
    }
}

#[tauri::command]
fn save_settings(settings: AppSettings) -> bool {
    let _ = settings;
    true
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

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
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
            environment,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

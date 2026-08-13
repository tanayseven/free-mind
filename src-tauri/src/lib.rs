use serde::{Deserialize, Serialize};

// NOTE: These commands are stubs for the initial Tauri scaffold. The website
// blocking / root-daemon logic is not implemented yet — each command returns a
// sensible default so the existing frontend can run against the Tauri backend.

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
    false
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
    let _ = list;
    true
}

#[tauri::command]
fn start_blocking() -> bool {
    true
}

#[tauri::command]
fn stop_blocking() -> String {
    String::new()
}

#[tauri::command]
fn load_blocked_websites() -> String {
    "[]".to_string()
}

#[tauri::command]
fn save_blocked_websites(json: String) -> bool {
    let _ = json;
    true
}

#[tauri::command]
fn load_settings() -> AppSettings {
    AppSettings { unblock_waiting: 30 }
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

// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use log::{info};
use tauri::Manager;
use std::process::Command;
use tauri::{SystemTray, SystemTrayEvent, CustomMenuItem, SystemTrayMenu, WindowEvent};

#[tauri::command]
fn is_elevated() -> Result<bool, String> {
    info!("is_elevated function called");

    #[cfg(target_os = "windows")]
    {
        info!("Checking elevation on Windows");
        let output = Command::new("net session")
            .output()
            .map_err(|e| format!("Failed to run net session command: {}", e))?;
        let result = !str::from_utf8(&output.stdout).unwrap_or("").contains("Access is denied.");
        info!("Windows elevation result: {}", result);
        if !result {
            Ok(true)
        } else {
            Err("Access is denied".into())
        }
    }

    #[cfg(any(target_os = "linux", target_os = "macos"))]
    {
        info!("Checking elevation on Linux/macOS");
        let output = Command::new("id")
            .arg("-u")
            .output()
            .map_err(|e| format!("Failed to run id command: {}", e))?;

        let user_id = String::from_utf8_lossy(&output.stdout).trim().parse::<u32>().unwrap_or(0);
        Ok(user_id == 0)
    }

    #[cfg(not(any(target_os = "windows", target_os = "linux", target_os = "macos")))]
    {
        error!("Unsupported operating system");
        Err("Unsupported operating system".into())
    }
}

#[tauri::command]
fn read_file_contents(file_path: String) -> Result<String, String> {
    std::fs::read_to_string(file_path)
        .map_err(|err| format!("Failed to read file: {}", err))
}

#[tauri::command]
fn write_file_contents(file_path: String, contents: String) -> Result<(), String> {
    std::fs::write(file_path, contents)
        .map_err(|err| format!("Failed to write file: {}", err))
}

#[tauri::command]
fn restart_network() -> Result<(), String> {
    #[cfg(target_os = "macos")]
    {
        let output = Command::new("dscacheutil")
            .arg("-flushcache")
            .output()
            .map_err(|e| format!("Failed to command to restart network: {}", e))?;
        if output.status.success() {
            Ok(())
        } else {
            Err(format!("Failed to restart network: {}", String::from_utf8_lossy(&output.stderr)))
        }
    }

    #[cfg(target_os = "linux")]
    {
        let output = Command::new("/etc/init.d/networking")
            .arg("restart")
            .output()
            .map_err(|e| format!("Failed to command to restart network: {}", e))?;
        if output.status.success() {
            Ok(())
        } else {
            Err(format!("Failed to restart network: {}", String::from_utf8_lossy(&output.stderr)))
        }
//         let output = Command::new("/etc/init.d/nscd")
//             .arg("restart")
//             .output()
//             .map_err(|e| format!("Failed to command to restart network: {}", e))?;
//         if output.status.success() {
//             Ok(())
//         } else {
//             Err(format!("Failed to restart network: {}", String::from_utf8_lossy(&output.stderr)))
//         }
//         let output = Command::new("/etc/rc.d/nscd")
//             .arg("restart")
//             .output()
//             .map_err(|e| format!("Failed to command to restart network: {}", e))?;
//         if output.status.success() {
//             Ok(())
//         } else {
//             Err(format!("Failed to restart network: {}", String::from_utf8_lossy(&output.stderr)))
//         }
//         let output = Command::new("/etc/rc.d/init.d/nscd")
//             .arg("restart")
//             .output()
//             .map_err(|e| format!("Failed to command to restart network: {}", e))?;
    }

    #[cfg(target_os = "windows")]
    {
        let output = Command::new("ipconfig")
            .arg("/flushdns")
            .output()
            .map_err(|e| format!("Failed to command to restart network: {}", e))?;
        if output.status.success() {
            Ok(())
        } else {
            Err(format!("Failed to restart network: {}", String::from_utf8_lossy(&output.stderr)))
        }
    }
}

fn main() {
    let tray_menu = SystemTrayMenu::new()
        .add_item(CustomMenuItem::new("show", "Show"))
        .add_item(CustomMenuItem::new("quit", "Quit"));

    let system_tray = SystemTray::new().with_menu(tray_menu);

    tauri::Builder::default()
        .setup(|app| {
            info!("Tauri setup complete");
            #[cfg(debug_assertions)]
            {
                let window = app.get_window("main").unwrap();
                window.open_devtools();
            }
            Ok(())
        })
        .on_window_event(|event| {
            if let WindowEvent::CloseRequested { api, .. } = event.event() {
                api.prevent_close();
                let window = event.window();
                window.hide().unwrap();
            }
        })
        .system_tray(system_tray)
        .on_system_tray_event(|app, event| {
            match event {
                SystemTrayEvent::MenuItemClick { id, .. } => {
                    let main_window = app.get_window("main").unwrap();
                    match id.as_str() {
                        "show" => {
                            main_window.show().unwrap();
                        }
                        "quit" => {
                            std::process::exit(0);
                        }
                        _ => {}
                    }
                }
                _ => {}
            }
        })
        .invoke_handler(tauri::generate_handler![is_elevated, read_file_contents, write_file_contents, restart_network])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

//! Status bar / system tray icon. Keeps the app reachable while its window is
//! hidden (closing the window hides it; only "Quit" exits).

use std::thread;
use std::time::Duration;

use tauri::menu::{Menu, MenuItem, PredefinedMenuItem};
use tauri::tray::TrayIconBuilder;
use tauri::{AppHandle, Manager};

use crate::hosts;
use crate::ipc::{self, Request};

const MENU_TOGGLE: &str = "toggle";
const MENU_OPEN: &str = "open";
const MENU_QUIT: &str = "quit";

fn status_text(blocking: bool) -> &'static str {
    if blocking {
        "Blocking is on"
    } else {
        "Blocking is off"
    }
}

fn toggle_text(blocking: bool) -> &'static str {
    if blocking {
        // Opens the window, so the unblock countdown still applies.
        "Stop Blocking…"
    } else {
        "Start Blocking"
    }
}

pub fn create(app: &AppHandle) -> tauri::Result<()> {
    let blocking = hosts::is_blocking();
    let status = MenuItem::with_id(app, "status", status_text(blocking), false, None::<&str>)?;
    let toggle = MenuItem::with_id(app, MENU_TOGGLE, toggle_text(blocking), true, None::<&str>)?;
    let open = MenuItem::with_id(app, MENU_OPEN, "Open Free Mind", true, None::<&str>)?;
    let quit = MenuItem::with_id(app, MENU_QUIT, "Quit Free Mind", true, None::<&str>)?;
    let menu = Menu::with_items(
        app,
        &[
            &status,
            &toggle,
            &PredefinedMenuItem::separator(app)?,
            &open,
            &PredefinedMenuItem::separator(app)?,
            &quit,
        ],
    )?;

    let mut tray = TrayIconBuilder::with_id("main")
        .tooltip("Free Mind")
        .menu(&menu)
        .show_menu_on_left_click(true)
        .on_menu_event(|app, event| match event.id().as_ref() {
            MENU_TOGGLE => toggle_blocking(app),
            MENU_OPEN => show_main_window(app),
            MENU_QUIT => app.exit(0),
            _ => {}
        });
    if let Some(icon) = app.default_window_icon() {
        tray = tray.icon(icon.clone());
    }
    tray.build(app)?;

    // Keep the menu in sync with the hosts file, whoever changed it.
    thread::spawn(move || {
        let mut shown = blocking;
        loop {
            thread::sleep(Duration::from_secs(1));
            let now = hosts::is_blocking();
            if now != shown {
                let _ = status.set_text(status_text(now));
                let _ = toggle.set_text(toggle_text(now));
                shown = now;
            }
        }
    });
    Ok(())
}

/// Starts blocking directly. Stopping goes through the window instead, so the
/// unblock countdown can't be skipped from the tray.
fn toggle_blocking(app: &AppHandle) {
    if hosts::is_blocking() {
        show_main_window(app);
        return;
    }
    let app = app.clone();
    // Off the main thread: the daemon round trip shouldn't stall the menu.
    thread::spawn(move || {
        let request = Request::Block {
            domains: crate::enabled_domains(),
        };
        if let Err(e) = ipc::send(&request) {
            eprintln!("tray: failed to start blocking: {e}");
            // Most likely the daemon isn't installed; the window handles that.
            show_main_window(&app);
        }
    });
}

pub fn show_main_window(app: &AppHandle) {
    if let Some(window) = app.get_webview_window("main") {
        let _ = window.show();
        let _ = window.unminimize();
        let _ = window.set_focus();
    }
}

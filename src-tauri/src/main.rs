// Prevents an additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.iter().any(|a| a == "--daemon") {
        if let Err(e) = free_mind_lib::daemon::run(&args) {
            eprintln!("free-mind daemon: {e}");
            std::process::exit(1);
        }
        return;
    }
    free_mind_lib::run()
}

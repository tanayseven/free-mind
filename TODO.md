## Done

- [x] Migrate from Wails/Go to Tauri v2 + Rust
- [x] Manage toolchain with mise (Rust, Node, Tauri CLI)
- [x] Rewire frontend from Wails bindings to Tauri `invoke` (`frontend/src/lib/api.ts`)
- [x] Switch SvelteKit to `adapter-static` (SPA) for the Tauri webview
- [x] Scaffold `src-tauri` with stubbed backend commands

## Pending

- [ ] Implement website blocking in Rust
  - [ ] Modify the system hosts file with elevated privileges
  - [ ] DNS cache flush per platform
    - Linux: `nscd` / `systemd-resolve --flush-caches`
    - macOS: `dscacheutil -flushcache`
    - Windows: `ipconfig /flushdns`
  - [ ] Wire up `data/default-blocklist.json`
- [ ] Focus timer feature (Pomodoro, Countdown, Stopwatch, Schedule)
- [ ] Packaging via `tauri build`
  - [ ] .deb / .AppImage / .rpm (Linux)
  - [ ] .dmg (macOS, x64 + aarch64)
  - [ ] .msi / .exe (Windows)
- [ ] Cross-platform test execution (Linux, macOS, Windows)

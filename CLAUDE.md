# Free Mind — CLAUDE.md

## Project Overview

**Free Mind** is a cross-platform desktop app that blocks distracting websites and helps users stay focused.

**Key traits:** offline-only, zero data collection. Websites are blocked by editing the system hosts file through a privileged helper daemon, installed once with a single admin prompt. Website list and settings persistence are still in-memory stubs.

---

## Architecture

```
Frontend (SvelteKit / TypeScript / Tailwind)
    ↓ Tauri `invoke` (frontend/src/lib/api.ts)
src-tauri (Rust, user privileges)
    ↓ #[tauri::command] handlers (src-tauri/src/lib.rs)
    ↓ JSON lines over a local socket (src-tauri/src/ipc.rs)
Daemon (same binary with `--daemon`, root / LocalSystem)
    ↓ writes the hosts file (src-tauri/src/hosts.rs)
```

- **Frontend**: SvelteKit built as a static SPA (`@sveltejs/adapter-static`), served inside the native webview.
- **Rust backend** (`src-tauri/`): exposes commands via `#[tauri::command]`. Blocking commands forward to the daemon. Website list and settings are kept in memory only.
- **Daemon** (`src-tauri/src/daemon.rs`): the app binary started with `--daemon` by launchd (macOS), systemd (Linux) or the Windows service manager. It is the only thing that writes the hosts file, and it re-validates every domain.

### Daemon install

`install::install()` (`src-tauri/src/install.rs`) copies the running executable to an admin-only location and registers it as a boot-time service, behind one elevation prompt (`osascript` / `pkexec` / UAC). On startup, `connect_to_daemon` byte-compares the installed copy with the running exe. A mismatch (app update or dev rebuild) makes the frontend reinstall it, which shows the prompt again.

| OS | Binary | Service | Socket |
|----|--------|---------|--------|
| macOS | `/Library/PrivilegedHelperTools/tech.tanay.free-mind.daemon` | `/Library/LaunchDaemons/tech.tanay.free-mind.daemon.plist` | `/var/run/tech.tanay.free-mind.sock` (mode 0600, owned by installing user) |
| Linux | `/usr/local/libexec/free-mind/free-mind-daemon` | `free-mind-daemon.service` (systemd) | `/run/tech.tanay.free-mind.sock` (same) |
| Windows | `C:\Program Files\Free Mind\free-mind-daemon.exe` | `FreeMindDaemon` service | `\\.\pipe\tech.tanay.free-mind` (authenticated users) |

Uninstall:

```bash
# macOS
sudo launchctl bootout system/tech.tanay.free-mind.daemon
sudo rm /Library/LaunchDaemons/tech.tanay.free-mind.daemon.plist /Library/PrivilegedHelperTools/tech.tanay.free-mind.daemon
# Linux
sudo systemctl disable --now free-mind-daemon && sudo rm /etc/systemd/system/free-mind-daemon.service /usr/local/libexec/free-mind/free-mind-daemon
# Windows (admin PowerShell)
Stop-Service FreeMindDaemon; sc.exe delete FreeMindDaemon; Remove-Item "C:\Program Files\Free Mind" -Recurse
```

### Frontend ↔ Rust bindings

`frontend/src/lib/api.ts` wraps each Tauri command with a typed function (e.g. `CheckBlocking()` → `invoke('check_blocking')`). This replaced the old auto-generated Wails bindings. External links are opened with `@tauri-apps/plugin-opener` (`BrowserOpenURL`).

Rust commands live in `src-tauri/src/lib.rs`:

| Command | Behaviour |
|---------|-----------|
| `connect_to_daemon` | Pings the daemon; `""` = connected, `"Error: …"` = not installed / outdated / unreachable |
| `check_blocking` | Reads the hosts file for the Free Mind block (no IPC; polled every 100ms) |
| `check_daemon_installed` | Whether the daemon binary exists at its install path |
| `install_and_start_daemon` | One-time elevated install; waits for the daemon to answer |
| `send_block_list(list)` | Validates and stores comma-separated domains in memory |
| `start_blocking` | Sends the stored list to the daemon |
| `stop_blocking` | Asks the daemon to remove the block; `""` on success |
| `load_blocked_websites` | In-memory list, seeded with social / news / video defaults (stub) |
| `save_blocked_websites(json)` | Stores the list in memory (stub) |
| `load_settings` | `{ unblockWaiting: 30 }` (stub) |
| `save_settings(settings)` | `true` (stub) |
| `environment` | `{ platform, arch }` (platform mapped to Go's GOOS names) |

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Rust (edition 2021), Tauri v2 |
| Frontend | Svelte 5, SvelteKit 2, TypeScript 5 |
| Styling | Tailwind CSS 4 |
| Build tool | Vite 7 |
| Testing | Vitest, Playwright |
| Icons | Lucide Svelte |
| Toolchain | mise (Rust, Node, Tauri CLI) |

---

## Toolchain (mise)

All tools are managed by [mise](https://mise.jdx.dev) via `.mise.toml`:

```toml
[tools]
node = "25.6.1"
rust = "stable"
"npm:@tauri-apps/cli" = "latest"
```

```bash
mise install        # install Rust, Node, and the Tauri CLI
mise exec -- <cmd>  # run a command with the toolchain on PATH
```

Once mise is activated in your shell, `tauri`, `cargo`, and `npm` are available directly.

---

## Build & Run

```bash
make dev              # tauri dev — Vite hot reload + native window
make build            # tauri build — bundle the desktop app
make build-frontend   # build the Svelte static assets only
make icons SRC=x.png  # regenerate app icons from a 1024x1024 PNG
make lint             # frontend lint + format check
make test             # frontend tests
make help             # list all targets
```

Frontend commands (run from `frontend/`):

```bash
npm run dev       # Vite dev server
npm run build     # static build → frontend/build
npm run check     # svelte-check type checking
npm run lint      # ESLint + Prettier
npm run test      # Vitest tests
```

---

## Key Files

| File | Role |
|------|------|
| `src-tauri/src/lib.rs` | Tauri app setup + command handlers |
| `src-tauri/src/main.rs` | Binary entry point → `daemon::run()` with `--daemon`, else `free_mind_lib::run()` |
| `src-tauri/src/daemon.rs` | Privileged daemon: socket server + Windows service glue |
| `src-tauri/src/ipc.rs` | App ↔ daemon message types and client |
| `src-tauri/src/install.rs` | One-time elevated daemon install per OS |
| `src-tauri/src/hosts.rs` | Hosts file block editing + DNS flush |
| `src-tauri/Cargo.toml` | Rust crate + Tauri dependencies |
| `src-tauri/tauri.conf.json` | Tauri config (window, bundle, build hooks) |
| `src-tauri/capabilities/default.json` | Permissions for the main window |
| `frontend/src/lib/api.ts` | Typed Tauri command bindings |
| `frontend/src/routes/+page.svelte` | Main UI (tabs: Home / Modes / Websites / Settings) |
| `frontend/src/routes/+layout.ts` | SPA config (`ssr = false`, `prerender = true`) |
| `frontend/svelte.config.js` | `adapter-static` with SPA fallback |
| `.mise.toml` | Toolchain versions |
| `Makefile` | Build automation |
| `data/default-blocklist.json` | Default categorized block list (data, not yet wired up) |

---

## Outstanding TODOs

- Persist the website list and settings to disk
- Test the daemon install on Linux and Windows
- Packaging: DEB, AppImage, RPM (Linux), DMG (macOS), MSI (Windows) via `tauri build`
- Focus timer feature (Pomodoro, Countdown, Stopwatch, Schedule)
- Comprehensive cross-platform testing

---

## Conventions

- Rust follows `cargo fmt` / `cargo clippy`
- Frontend uses Prettier + ESLint (see `frontend/eslint.config.js`)
- Path alias in frontend: `@/` and `$lib` → `frontend/src/lib`
- Do not edit generated files under `frontend/.svelte-kit/` or `src-tauri/gen/`

# Free Mind — CLAUDE.md

## Project Overview

**Free Mind** is a cross-platform desktop app that blocks distracting websites and helps users stay focused.

**Key traits:** offline-only, zero data collection. Website blocking (which modifies the system hosts file and requires elevated privileges) is **not yet implemented** in the Tauri version — the current codebase is a scaffold with a working UI and stubbed backend commands.

---

## Architecture

```
Frontend (SvelteKit / TypeScript / Tailwind)
    ↓ Tauri `invoke` (frontend/src/lib/api.ts)
src-tauri (Rust)
    ↓ #[tauri::command] handlers (src-tauri/src/lib.rs)
```

The app is a single Tauri process:

- **Frontend** — SvelteKit built as a static SPA (`@sveltejs/adapter-static`), served inside the native webview.
- **Rust backend** (`src-tauri/`) — exposes commands via `#[tauri::command]`. These are currently **stubs** returning sensible defaults so the UI runs end-to-end.

### Frontend ↔ Rust bindings

`frontend/src/lib/api.ts` wraps each Tauri command with a typed function (e.g. `CheckBlocking()` → `invoke('check_blocking')`). This replaced the old auto-generated Wails bindings. External links are opened with `@tauri-apps/plugin-opener` (`BrowserOpenURL`).

Rust commands live in `src-tauri/src/lib.rs`:

| Command | Returns (stub) |
|---------|----------------|
| `connect_to_daemon` | `""` (empty = connected) |
| `check_blocking` | `false` |
| `check_daemon_installed` | `true` |
| `install_and_start_daemon` | `"Daemon installed"` |
| `send_block_list(list)` | `true` |
| `start_blocking` | `true` |
| `stop_blocking` | `""` |
| `load_blocked_websites` | `"[]"` |
| `save_blocked_websites(json)` | `true` |
| `load_settings` | `{ unblockWaiting: 30 }` |
| `save_settings(settings)` | `true` |
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
| `src-tauri/src/lib.rs` | Tauri app setup + command handlers (stubs) |
| `src-tauri/src/main.rs` | Binary entry point → `free_mind_lib::run()` |
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

- Implement website blocking in Rust (hosts file, DNS flush, elevated privileges)
- Packaging: DEB, AppImage, RPM (Linux), DMG (macOS), MSI (Windows) via `tauri build`
- Focus timer feature (Pomodoro, Countdown, Stopwatch, Schedule)
- Comprehensive cross-platform testing

---

## Conventions

- Rust follows `cargo fmt` / `cargo clippy`
- Frontend uses Prettier + ESLint (see `frontend/eslint.config.js`)
- Path alias in frontend: `@/` and `$lib` → `frontend/src/lib`
- Do not edit generated files under `frontend/.svelte-kit/` or `src-tauri/gen/`

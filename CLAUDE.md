# Free Mind — CLAUDE.md

## Project Overview

**Free Mind** is a cross-platform desktop app that blocks distracting websites and helps users stay focused. It uses a two-process architecture: a user-level GUI and a root daemon that modifies `/etc/hosts`.

**Key traits:** offline-only, zero data collection, requires root/admin for blocking.

---

## Architecture

```
Frontend (Svelte/TS)
    ↓ Wails bindings
app.go — Main process (user-level)
    ↓ IPC (Unix socket / Windows named pipe)
root-daemon/main.go — Daemon (root/admin)
    ↓
/etc/hosts (website blocking)
```

### Two-Process Design

1. **Main app** (`app.go`) — Wails desktop GUI, embeds daemon binaries, manages daemon lifecycle, sends IPC commands.
2. **Root daemon** (`root-daemon/main.go`) — Runs as root, accepts IPC connections, modifies `/etc/hosts`, flushes DNS cache.

### IPC Package (`ipc/`)

- `ipc.go` — Platform-independent interfaces
- `unix.go` — Unix domain socket: `/tmp/tech.tanay.free-mind.sock`
- `windows.go` — Windows named pipe: `\\.\pipe\tech.tanay.free-mind`
- JSON message protocol: `{ Action: "start"|"stop"|"update", Content: "..." }`

### Website Blocking

- Appends to `/etc/hosts` using comment markers to track Free Mind's additions
- Block list stored in `/etc/hosts-list-to-be-blocked`
- DNS flush: Linux (`nscd`), macOS (`dscacheutil -flushcache`), Windows (`ipconfig /flushdns`)

### Daemon Lifecycle

- Daemon binaries embedded in the main executable (via `//go:embed build/bin/*`)
- Extracted to temp, then installed to `/usr/bin/free-mind-daemon` with elevated privileges
- Elevation: `pkexec` (Linux), `osascript` (macOS), PowerShell (Windows)
- Socket readiness: 20 retries × 500ms delay

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23 |
| Desktop framework | Wails v2.11.0 |
| Frontend | Svelte 5, SvelteKit 2, TypeScript 5 |
| Styling | Tailwind CSS 4 |
| Build tool | Vite 7 |
| Testing | Vitest, Playwright |
| Icons | Lucide Svelte |
| Windows IPC | github.com/Microsoft/go-winio |

---

## Build & Run

```bash
# Development
wails dev

# Build daemon for all platforms
make build-daemon-all

# Build app for current platform
wails build

# Individual daemon builds
make build-daemon-linux
make build-daemon-macos
make build-daemon-windows

# See all targets
make help
```

Frontend commands (run from `frontend/`):
```bash
npm run dev       # Dev server
npm run build     # Production build
npm run check     # Type checking
npm run lint      # ESLint + Prettier
npm run test      # Vitest tests
```

---

## Key Files

| File | Role |
|------|------|
| `app.go` | Wails app, daemon install/start/stop, IPC client calls |
| `main.go` | Wails entry point, embeds daemon binaries |
| `root-daemon/main.go` | Root daemon: IPC server, hosts file manipulation |
| `ipc/ipc.go` | IPC interface definitions |
| `ipc/unix.go` | Unix socket client/server |
| `ipc/windows.go` | Windows named pipe client/server |
| `daemon-test-client/main.go` | CLI tool for testing IPC manually |
| `frontend/src/routes/+page.svelte` | Main UI (Start/Stop blocking, daemon status) |
| `wails.json` | Wails configuration |
| `Makefile` | Build automation |

---

## Current Branch: `move-to-better-ipc`

The current branch migrated from ZMQ to a custom IPC package (`ipc/`) with platform-specific implementations. The plans in `plans/` document this migration.

---

## Outstanding TODOs

- Packaging: RPM, DEB, AppImage (Linux), DMG (macOS), MSI (Windows)
- Verify DNS cache flush commands on all platforms
- Focus timer feature (Pomodoro, Countdown, Stopwatch, Schedule)
- Comprehensive cross-platform testing

---

## Conventions

- Go code follows standard formatting (`gofmt`)
- Frontend uses Prettier + ESLint (see `frontend/.eslintrc` / `frontend/package.json`)
- Path alias in frontend: `@/` → `src/lib/`
- Wails bindings are auto-generated in `frontend/wailsjs/go/main/`; do not edit manually

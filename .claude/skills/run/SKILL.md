---
name: run
description: "Use this skill to build or run the free-mind project. Covers: running the app in dev mode, building the Wails desktop app, and building the root daemon for one or all platforms."
compatibility: "Linux, macOS, Windows with Go 1.23+, Wails v2, Node.js, and nvm installed"
---

# Free Mind — Build & Run

## Available targets

| Command | What it does |
|---------|-------------|
| `make run-app` | Run the app in dev mode (hot reload via Wails + Vite) |
| `make build-wails-app` | Build the desktop app for the current OS/arch |
| `make build-daemon` | Build the root daemon for the current OS/arch |
| `make build-daemon-all` | Build the daemon for Linux, macOS, and Windows (amd64) |
| `make build-daemon-linux` | Build daemon → `build/bin/free-mind-daemon-linux` |
| `make build-daemon-macos` | Build daemon → `build/bin/free-mind-daemon-darwin` |
| `make build-daemon-windows` | Build daemon → `build/bin/free-mind-daemon-windows.exe` |
| `make test` | Run all tests (`./...`) |
| `make test-ipc` | Run IPC layer tests with verbose output |
| `make test-daemon` | Run root daemon logic tests with verbose output |
| `make test-app` | Run app-level IPC client tests with verbose output |
| `make help` | List all targets with descriptions |

---

## Running in development

```bash
# From the repo root
make run-app
```

This calls `nvm install` then `wails dev`, which:
1. Starts the Vite dev server for the Svelte frontend (hot reload)
2. Launches the Go Wails app pointing at the dev server
3. Watches Go files and rebuilds on change

> The daemon is **not** started automatically in dev mode. If you need to test
> blocking, build and install the daemon separately (see below).

---

## Building the desktop app

```bash
make build-wails-app
```

Produces a native executable for the current platform in `build/bin/`.
The daemon binaries (already in `build/bin/`) are embedded at compile time via
`//go:embed build/bin/*` in `main.go` — make sure they exist before building.

**Recommended order:**

```bash
make build-daemon-all   # build daemons for all platforms first
make build-wails-app    # then embed them into the main app
```

---

## Building the root daemon

The daemon runs as root and modifies `/etc/hosts`. It must be built separately
and embedded in (or installed alongside) the main app.

### Current platform only

```bash
make build-daemon
# Output: build/bin/free-mind-daemon
```

### All platforms (cross-compile)

```bash
make build-daemon-all
# Output:
#   build/bin/free-mind-daemon-linux       (GOOS=linux  GOARCH=amd64)
#   build/bin/free-mind-daemon-darwin      (GOOS=darwin GOARCH=amd64)
#   build/bin/free-mind-daemon-windows.exe (GOOS=windows GOARCH=amd64)
```

### Single platform

```bash
make build-daemon-linux    # Linux amd64
make build-daemon-macos    # macOS amd64
make build-daemon-windows  # Windows amd64
```

---

## First-time setup

```bash
# 1. Install Go dependencies
go mod tidy

# 2. Install frontend dependencies
cd frontend && npm install && cd ..

# 3. Install the Wails CLI (if not already installed)
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 4. Verify everything is in order
wails doctor
```

---

## Running tests

### All packages

```bash
make test
```

### Individual packages

```bash
make test-ipc      # IPC layer (Unix socket send/receive, server lifecycle)
make test-daemon   # Root daemon logic (hosts file manipulation, ProcessMessage routing)
make test-app      # App-level IPC client methods (ConnectToDaemon, SendBlockList, Start/StopBlocking)
```

### Notes

- Tests are Linux/macOS only (`//go:build linux || darwin`); Windows named-pipe tests are not yet written.
- The root-daemon tests mock `/etc/hosts` and `/etc/hosts-list-to-be-blocked` — they do **not** touch system files and require no elevated privileges.
- App tests redirect the Unix socket path via `ipc.OverrideSocketPath` so they never interfere with a running daemon at `/tmp/tech.tanay.free-mind.sock`.

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `wails: command not found` | Run `go install github.com/wailsapp/wails/v2/cmd/wails@latest` and ensure `$(go env GOPATH)/bin` is on `$PATH` |
| `nvm: command not found` | Source nvm in your shell: `source ~/.nvm/nvm.sh` |
| Embedded daemon binary missing at build time | Run `make build-daemon-all` before `make build-wails-app` |
| App starts but daemon won't connect | Check socket exists at `/tmp/tech.tanay.free-mind.sock`; run the daemon manually with `sudo /usr/bin/free-mind-daemon` |
| Cross-compile fails for Windows | Ensure CGO is disabled: `CGO_ENABLED=0 make build-daemon-windows` |

---
name: run
description: "Use this skill to build or run the free-mind project. Covers: running the Tauri app in dev mode and building the desktop bundle."
compatibility: "Linux, macOS, Windows with mise (Rust, Node, Tauri CLI v2)"
---

# Free Mind — Build & Run

Free Mind is a Tauri v2 desktop app: a SvelteKit frontend inside a native
webview, with a Rust backend in `src-tauri/`.

## First-time setup

All tooling is managed by [mise](https://mise.jdx.dev) from `.mise.toml`
(Rust, Node, and the Tauri CLI):

```bash
mise install                    # install Rust, Node, Tauri CLI
npm ci --prefix frontend        # install frontend dependencies
```

Once mise is activated in your shell, `tauri`, `cargo`, and `npm` are on PATH.
If mise is not activated, prefix commands with `mise exec --`.

## Available targets

| Command | What it does |
|---------|-------------|
| `make dev` | Run the app in dev mode (Tauri window + Vite hot reload) |
| `make build` | Build the desktop bundle for the current OS/arch |
| `make build-frontend` | Build only the Svelte static assets (`frontend/build`) |
| `make icons SRC=path.png` | Regenerate app icons from a 1024x1024 PNG |
| `make lint` | Lint + format-check the frontend |
| `make check` | Type-check the frontend (`svelte-check`) |
| `make test` | Run frontend tests (Vitest) |
| `make help` | List all targets |

## Running in development

```bash
make dev        # equivalent to: tauri dev
```

This:
1. Runs `beforeDevCommand` (`npm --prefix frontend run dev`) — Vite dev server on `:5173`
2. Compiles the Rust backend in `src-tauri/`
3. Launches the native window pointing at the dev server (hot reload)

The first run compiles all Tauri crates and takes a few minutes; subsequent runs
are incremental.

## Building the desktop app

```bash
make build      # equivalent to: tauri build
```

Runs `beforeBuildCommand` (static frontend build → `frontend/build`), compiles
the Rust backend in release mode, and produces platform bundles under
`src-tauri/target/release/bundle/`.

## Notes

- The Rust commands in `src-tauri/src/lib.rs` are **stubs** — website blocking /
  daemon logic is not implemented yet.
- Frontend ↔ Rust bindings live in `frontend/src/lib/api.ts` (Tauri `invoke`).
- macOS uses the system WKWebView (no extra deps). On Linux you need
  `webkit2gtk` and related packages (see `.github/workflows/ci.yml`).

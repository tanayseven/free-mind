# Free Mind

## About

Free Mind is a desktop application designed to help users maintain focus by blocking distracting websites and managing productive time intervals. It empowers users to concentrate on what matters most by eliminating digital distractions.

> **Status:** The app is being rebuilt on **Tauri v2 + Rust**. The current codebase is a scaffold with a working UI and stubbed backend commands — website blocking is not implemented yet.

## Features & Scope

### Platform Support
- **Cross-Platform Desktop Application**: Available for Windows, macOS, and Linux
- **Architecture Support**: Compatible with both x86_64 and ARM64 processors

### Core Functionality
- **Website Blocking**: Prevents access to user-defined distracting websites *(planned)*
- **Time Management**: Tracks and manages focus sessions with various timer options
- **Customizable Blocking**: Manage individual websites or entire categories
- **Permission Requirements**: Will require administrator/root privileges for website blocking

### Privacy & Security
- **100% Privacy-Focused**: Operates entirely locally without sending any data over the network
- **No Data Collection**: Your browsing habits and focus patterns remain on your device
- **No Internet Dependency**: Functions fully offline after installation

## Architecture

Free Mind is a single Tauri process:

```
Frontend (SvelteKit / TypeScript / Tailwind)
    ↓ Tauri `invoke` (frontend/src/lib/api.ts)
src-tauri (Rust) — #[tauri::command] handlers
```

- **Frontend** — SvelteKit built as a static SPA and served in the native webview.
- **Rust backend** (`src-tauri/`) — exposes commands via `#[tauri::command]`, currently stubbed.

### Focus Timer (planned)

| Timer Type | Description                                                               |
|------------|---------------------------------------------------------------------------|
| Pomodoro   | Alternates between focus sessions and breaks                              |
| Countdown  | Counts down from a user-defined focus period, then switches to break time |
| Stopwatch  | Tracks elapsed time since focus mode was activated                        |
| Schedule   | Automatically manages focus/break periods based on pre-defined schedules  |

### Technical Stack

- **Backend**: Rust + [Tauri v2](https://v2.tauri.app)
- **Frontend**: TypeScript + SvelteKit + Tailwind CSS
- **Toolchain**: [mise](https://mise.jdx.dev) (Rust, Node, Tauri CLI)

## Development Setup

### Prerequisites

Install [mise](https://mise.jdx.dev/getting-started.html). It manages the Rust,
Node, and Tauri CLI versions pinned in `.mise.toml`.

On **Linux** you also need the Tauri system dependencies (`webkit2gtk`, etc.) —
see the [Tauri prerequisites](https://v2.tauri.app/start/prerequisites/). macOS
and Windows use the system webview and need no extra packages.

### Setup

```shell
mise install                 # Rust, Node, Tauri CLI
npm ci --prefix frontend     # frontend dependencies
```

### Running the Application

```shell
make dev      # development mode (tauri dev)
make build    # build the desktop bundle (tauri build)
```

See `make help` for all targets, or the `run` skill under `.claude/skills/run`.

## License

(Decision pending)

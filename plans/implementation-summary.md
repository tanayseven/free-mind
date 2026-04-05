# IPC Implementation Summary

## Overview

This document summarizes the changes made to implement the IPC package and update the root daemon and app.go files to use the new IPC package instead of ZMQ.

## Changes Made

### 1. Created IPC Package

Created a new package called `ipc` with the following files:

- **ipc/ipc.go**: Defines the interfaces for Server, Connection, and Client, as well as the Message struct.
- **ipc/unix.go**: Implements the interfaces for Linux and Darwin using Unix domain sockets.
- **ipc/windows.go**: Implements the interfaces for Windows using named pipes.

### 2. Updated Root Daemon

Updated the root-daemon/main.go file to:

- Replace ZMQ router with IPC server
- Update message handling to use the new IPC connection
- Keep the same message structure and processing logic
- Update port file handling to socket/pipe path handling

### 3. Updated App

Updated the app.go file to:

- Replace ZMQ dealer with IPC client
- Update connection logic to use the new IPC client
- Keep the same message structure
- Update port file handling to socket/pipe path handling

### 4. Updated Daemon Test Client

Updated the daemon-test-client/main.go file to:

- Replace ZMQ dealer with IPC client
- Update message handling to use the new IPC connection
- Keep the same message structure

### 5. Updated Dependencies

Updated the go.mod file to:

- Add github.com/Microsoft/go-winio as a direct dependency for Windows named pipes

## Key Features

### Platform-Independent Interface

The IPC package provides a platform-independent interface for inter-process communication, with platform-specific implementations for:

- Linux/Darwin: Unix domain sockets
- Windows: Named pipes

### Socket/Pipe Paths

The socket/pipe paths are fixed for each platform:

- Linux/Darwin: `/run/tech.tanay.free-mind.sock`
- Windows: `\\.\pipe\tech.tanay.free-mind`

### Message Structure

The Message struct is used for communication between the daemon and the client:

```go
type Message struct {
    Action  string `json:"action"`
    Content string `json:"content"`
}
```

### JSON Serialization/Deserialization

Messages are serialized to JSON for transmission and deserialized from JSON upon reception.

## Benefits

1. **Platform Independence**: The IPC package provides a platform-independent interface for inter-process communication.
2. **Simplified Code**: The IPC package simplifies the code by abstracting away the platform-specific details.
3. **Improved Security**: Unix domain sockets and Windows named pipes are more secure than TCP sockets.
4. **Reduced Dependencies**: Removed the dependency on ZMQ, which is a complex library.
5. **Better Error Handling**: The IPC package provides better error handling and more meaningful error messages.

### 6. Written Unit and Integration Tests

- **ipc/unix_test.go**: Unit tests for Unix socket client/server
- **root-daemon/daemon_test.go**: Unit tests for daemon hosts-file manipulation and IPC handling
- **app_test.go**: Unit tests for app-level IPC and daemon lifecycle methods
- **test-socket-connection.sh**: Shell script for manual socket connection testing

### 7. Frontend Improvements

Significant UI work was done alongside the IPC migration:

- Extracted `Footer.svelte` from `+page.svelte` (with `Footer.svelte.spec.ts` unit tests)
- Header inlined directly into `+page.svelte` (a separate `Header.svelte` was created then removed)
- Added dark mode toggle (Switch component) with `applyTheme`/`detectInitialTheme` from `theme.ts`
- Added daemon connection status indicator in the header area
- Added `theme.ts` with dark/light mode logic and corresponding `theme.spec.ts` tests
- Added tabbed navigation (Focus / Block / Debug tabs) with `ui/tabs` components
- Added `ui/tooltip` components for icon labels
- Added `ui/switch` component
- Added `debug.svelte` tab for daemon connection status
- Renamed the app from "Free-Mind" to "Free Mind"
- Converted state variables to Svelte 5 `$state()` runes

### 8. Documentation and Project Setup

- Added `CLAUDE.md` with full architecture overview and build instructions
- Added `docs/architecture.md` with architecture diagrams
- Added `.claude/skills/run/SKILL.md` and `.claude/skills/renovate-merge/SKILL.md`
- Added `README.md`

### 9. Modes UI

Added a four-mode blocking system to the frontend:

- **⛓️‍💥 Free** — simple toggle switch (previously the only home-screen control)
- **⏲️ Timer** — duration presets (15 min / 30 min / 1 h / 2 h) with a start/stop button; backend pending
- **🗓️ Schedule** — day-of-week selector with a time-range picker; backend pending
- **⏰ Pomodoro** — configurable work/break/cycle counts with a start/stop button; backend pending

New files:

| File | Role |
|------|------|
| `frontend/src/lib/components/modes/StatusDot.svelte` | Shared LED blinker (red pulsing = blocking, green = not blocking) |
| `frontend/src/lib/components/modes/FreeMode.svelte` | Free-mode widget |
| `frontend/src/lib/components/modes/TimerMode.svelte` | Timer-mode widget |
| `frontend/src/lib/components/modes/ScheduleMode.svelte` | Schedule-mode widget |
| `frontend/src/lib/components/modes/PomodoroMode.svelte` | Pomodoro-mode widget |

Home tab changes:
- Shows the active mode label (e.g. "⛓️‍💥 Free") above the current mode's widget.
- Mode widget always includes the `StatusDot` LED blinker.

Modes tab changes:
- `ToggleGroup` (single-select) to switch between the four modes.
- Destructive `Alert` appears when blocking is active, and the `ToggleGroup` is disabled — mode browsing is read-only while blocking.

## Socket Path Note

The actual socket path used is `/tmp/tech.tanay.free-mind.sock` (not `/run/...`), chosen because `/tmp` is writable by all users without requiring root for socket creation.

## Testing Status

Unit tests are written and cover cases 1–5 from the testing plan. Cross-platform execution (Linux, macOS, Windows) is pending.

## Testing Plan

A testing plan has been created to test the implementation on different platforms:

- Linux (Ubuntu/Debian)
- macOS (Darwin)
- Windows

The testing plan includes test cases for:

- Daemon installation and startup
- Client connection
- Message sending and receiving
- Site blocking functionality
- Error handling
- Platform-specific tests
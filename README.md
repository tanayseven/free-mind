# Free Mind

## About

Free Mind is a desktop application designed to help users maintain focus by blocking distracting websites and managing productive time intervals. It empowers users to concentrate on what matters most by eliminating digital distractions.

## Features & Scope

### Platform Support
- **Cross-Platform Desktop Application**: Available for Windows, macOS, and Linux
- **Architecture Support**: Compatible with both x86_64 and ARM64 processors

### Core Functionality
- **Website Blocking**: Prevents access to user-defined distracting websites
- **Time Management**: Tracks and manages focus sessions with various timer options
- **Customizable Blocking**: Manage individual websites or entire categories
- **System Integration**: Runs in the background with status indicator in the system tray
- **Permission Requirements**: Requires administrator/root privileges for website blocking functionality

### Privacy & Security
- **100% Privacy-Focused**: Operates entirely locally without sending any data over the network
- **No Data Collection**: Your browsing habits and focus patterns remain on your device
- **No Internet Dependency**: Functions fully offline after installation

## Architecture

### Component Overview
The application consists of two primary components:

#### 1. Website Blocker
- Maintains a database of websites to block, organized by categories
- Provides granular control to enable/disable individual sites or entire categories
- Features an intuitive user interface with:
  - Tabular display of websites
  - Search and filter capabilities
  - Toggle controls for enabling/blocking
- Implements platform-specific blocking mechanisms for Windows, macOS, and Linux

#### 2. Focus Timer
The timer component offers multiple time management methods:

| Timer Type | Description                                                               |
|------------|---------------------------------------------------------------------------|
| Pomodoro   | Alternates between 25-minute focus sessions and 5-minute breaks           |
| Countdown  | Counts down from a user-defined focus period, then switches to break time |
| Stopwatch  | Tracks elapsed time since focus mode was activated                        |
| Schedule   | Automatically manages focus/break periods based on pre-defined schedules  |

### Technical Stack

- **Backend**: Go (for system-level interactions)
- **Frontend**: JavaScript/TypeScript with Svelte Kit framework
- **Desktop Integration**: Wails (provides API for JS/TS to communicate with Go backend)

## Development Setup

### Prerequisites
[Install Go v1.23](https://go.dev/dl/)
[Install Wails](https://wails.io/docs/gettingstarted/installation)
[Install NVM](https://github.com/nvm-sh/nvm)

### Frontend Setup
```shell
cd frontend/
nvm install
nvm use
npm i
```

### Running the Application
```shell
# Development mode
wails dev

# Build for all supported platforms
wails build
```

## Troubleshooting Desktop Issues

If the application works in browser but not on desktop, here are common issues and solutions:

1. **Daemon Installation Failure**: The application requires a daemon process to manage website blocking. This needs elevated privileges.
   - On Linux/macOS: Ensure `pkexec` is available
   - On Windows: Make sure PowerShell execution policies allow script execution

2. **Permission Issues**:
   - The daemon binary must be installed in system directories with proper permissions
   - Check that the application has sufficient privileges to modify `/etc/hosts`

3. **ZMQ Communication Problems**:
   - Ensure ZeroMQ libraries are properly linked
   - Verify that port files are created correctly

4. **Debugging Steps**:
   - Run `wails dev` and check browser console for errors
   - Check system logs for permission denials or process failures
   - Use the debug page in the application to get detailed error information

## Building for Desktop

To build a desktop version that works properly:

```shell
# Build daemon binaries for all platforms
make build-daemon-all

# Then build the main application
wails build
```

# License

(Decision pending)

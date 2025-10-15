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

| Timer Type | Description |
|------------|-------------|
| Pomodoro | Alternates between 25-minute focus sessions and 5-minute breaks |
| Countdown | Counts down from a user-defined focus period, then switches to break time |
| Stopwatch | Tracks elapsed time since focus mode was activated |
| Schedule | Automatically manages focus/break periods based on pre-defined schedules |

### Technical Stack

- **Backend**: Go (for system-level interactions)
- **Frontend**: JavaScript/TypeScript with Svelte Kit framework
- **Desktop Integration**: Wails (provides API for JS/TS to communicate with Go backend)

## Development Setup

### Prerequisites
1. [Install Wails](https://wails.io/docs/gettingstarted/installation)

### Frontend Setup
```shell
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

# License

(Decision pending)

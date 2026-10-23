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
- **Tray** (`src-tauri/src/tray.rs`): status bar / system tray icon with a menu (blocking status, Start Blocking, Stop Blocking…, Open, Quit). Closing the window only hides it; the app keeps running until Quit. "Stop Blocking…" opens the window instead of stopping, so the unblock countdown can't be bypassed.
- **Daemon** (`src-tauri/src/daemon.rs`): the app binary started with `--daemon` by launchd (macOS), systemd (Linux) or the Windows service manager. It is the only thing that writes the hosts file, and it re-validates every domain.

### Daemon install

`install::install()` (`src-tauri/src/install.rs`) copies the running executable to an admin-only location and registers it as a boot-time service, behind one elevation prompt (`osascript` / `pkexec` / UAC). On startup, `connect_to_daemon` byte-compares the installed copy with the running exe. A mismatch (app update or dev rebuild) makes the frontend reinstall it, which shows the prompt again.

- **AppImage:** the daemon is a copy of the AppImage itself (`$APPIMAGE`), not the inner binary, which needs the bundled libraries. Every service definition sets `APPIMAGE_EXTRACT_AND_RUN=1` so it runs as root without FUSE.
- **Linux init system:** the elevated script detects what is running: systemd (`/run/systemd/system`), else OpenRC (`openrc-run`), else runit (`/etc/sv` + `/var/service`), else Slackware rc (`/etc/rc.d/rc.M`).
- **Package-managed daemon:** if `/etc/free-mind/managed-daemon` exists (the NixOS module writes it), the app treats the daemon as installed and current, and only pings it.
- **macOS `.pkg`:** its `postinstall` installs the daemon at install time, so the app skips its own prompt.

| OS | Binary | Service | Socket |
|----|--------|---------|--------|
| macOS | `/Library/PrivilegedHelperTools/tech.tanay.free-mind.daemon` | `/Library/LaunchDaemons/tech.tanay.free-mind.daemon.plist` | `/var/run/tech.tanay.free-mind.sock` (mode 0600, owned by installing user) |
| Linux | `/usr/local/libexec/free-mind/free-mind-daemon` | `free-mind-daemon` (systemd unit, OpenRC `/etc/init.d`, runit `/etc/sv`, or Slackware `/etc/rc.d/rc.free-mind-daemon`) | `/run/tech.tanay.free-mind.sock` (same) |
| FreeBSD | `/usr/local/libexec/free-mind/free-mind-daemon` | `/usr/local/etc/rc.d/free_mind_daemon` (`daemon(8)`) | `/var/run/tech.tanay.free-mind.sock` (same) |
| Windows | `C:\Program Files\Free Mind\free-mind-daemon.exe` | `FreeMindDaemon` service | `\\.\pipe\tech.tanay.free-mind` (authenticated users) |

Uninstall:

```bash
# macOS
sudo launchctl bootout system/tech.tanay.free-mind.daemon
sudo rm /Library/LaunchDaemons/tech.tanay.free-mind.daemon.plist /Library/PrivilegedHelperTools/tech.tanay.free-mind.daemon
# Linux (systemd)
sudo systemctl disable --now free-mind-daemon && sudo rm /etc/systemd/system/free-mind-daemon.service /usr/local/libexec/free-mind/free-mind-daemon
# Linux (OpenRC)
sudo rc-service free-mind-daemon stop; sudo rc-update del free-mind-daemon; sudo rm /etc/init.d/free-mind-daemon /usr/local/libexec/free-mind/free-mind-daemon
# Linux (runit)
sudo rm /var/service/free-mind-daemon; sudo rm -r /etc/sv/free-mind-daemon /usr/local/libexec/free-mind/free-mind-daemon
# Linux (Slackware): also delete the rc.free-mind-daemon line from /etc/rc.d/rc.local
sudo /etc/rc.d/rc.free-mind-daemon stop; sudo rm /etc/rc.d/rc.free-mind-daemon /usr/local/libexec/free-mind/free-mind-daemon
# FreeBSD
sudo service free_mind_daemon stop; sudo sysrc -x free_mind_daemon_enable; sudo rm /usr/local/etc/rc.d/free_mind_daemon /usr/local/libexec/free-mind/free-mind-daemon
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
| `send_block_list(list)` | Validates and stores comma-separated domains in memory, adding `www.` variants |
| `start_blocking` | Sends the stored list to the daemon |
| `stop_blocking` | Asks the daemon to remove the block; `""` on success |
| `load_blocked_websites` | In-memory list, seeded with social / news / video defaults, one row per site (stub) |
| `save_blocked_websites(json)` | Stores the list in memory (stub) |
| `load_settings` | `{ unblockWaiting }`, default 30, kept in memory (stub) |
| `save_settings(settings)` | Stores settings in memory (stub) |
| `set_status_icon(status)` | Swaps tray + window icon: `inactive` / `blocked` / `unblocking` / `unblocked` |
| `environment` | `{ platform, arch }` (platform mapped to Go's GOOS names) |

### Icons

Source artwork lives in `frontend/static/` (800x600 SVGs): `frame` (locked brain), `broken-chain-1` (chains breaking), `broken-chain-2` (chains broken), each with a `-grayscale` variant. `make icons` regenerates everything from them:

| Use | Artwork |
|-----|---------|
| App bundle icon, favicon | `frame` |
| Tray/window: blocked | `frame` |
| Tray/window: unblocking (Free Mode countdown) | `broken-chain-1` |
| Tray/window: unblocked | `broken-chain-2` |
| Tray/window: inactive (daemon not connected) | `frame-grayscale` |

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
make icons            # regenerate app, tray and favicon icons from frontend/static/*.svg
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

## Releasing

`.github/workflows/release.yml` builds installers for every platform. It runs the Rust tests on each platform first.

| Platform | Formats | Built by |
|----------|---------|----------|
| macOS (universal) | `.dmg`, `.app.zip` | `build` job (`tauri-action`, `--target universal-apple-darwin`) |
| macOS (universal) | `.pkg` (installs the daemon too) | `macos-pkg` job: `pkgbuild` + `productbuild` with `packaging/macos-pkg/` |
| Linux (x64 + ARM64) | `.deb`, `.rpm`, `.AppImage`, `.tar.gz` | `build` job |
| Linux (x64 + ARM64) | `.pkg.tar.zst` (Arch) | `arch` job: repackages the `.deb` with `packaging/arch/PKGBUILD` in an `archlinux` container |
| Linux (x64 + ARM64) | `.snap` (classic) | `snap` job: repackages the `.deb` with `packaging/snap/snapcraft.yaml` |
| Linux (x64 + ARM64) | `.apk` (Alpine, musl) | `alpine` job: source build with `packaging/alpine/APKBUILD` in Docker |
| Linux (x64 + ARM64) | `.txz` (Slackware), `.xbps` (Void) | `linux-repack` job: repackages the `.deb` (`packaging/slackware/`, `packaging/void/`) |
| FreeBSD (x64) | `.pkg` | `freebsd` job: source build in a VM (`packaging/freebsd/package.sh`). Unofficial Tauri platform, so `continue-on-error`. |
| Windows (x64 + ARM64) | `.msi`, NSIS `.exe`, portable `.zip` | `build` job |
| Windows (x64 + ARM64) | `.msix` | `msix` job: packs `free-mind.exe` with `packaging/msix/AppxManifest.xml` using `makeappx` |
| Windows | `.nupkg` (Chocolatey) | `chocolatey` job: `choco pack` on the filled-in manifest |
| NixOS / Nix | flake (`flake.nix`, `packaging/nix/`) | `nix` job: only checks that it builds |
| Package managers | `package-manifests.zip` | `manifests` job: `scripts/generate-manifests.sh` fills in `packaging/` templates (Homebrew cask, winget, Scoop, Chocolatey, MacPorts, Solus, AUR `free-mind` + `free-mind-bin` with `.SRCINFO`, Gentoo overlay) |

The jobs after `build` use `gh release upload` to add their files to the same draft release. The package-manager manifests are submitted by hand, as are the `.snap` (Snap Store) and `.msix` (Microsoft Store).

`packaging/nix/package.nix` pins the npm dependencies' hash. When `frontend/package-lock.json` changes, the `nix` job fails and prints the new hash to paste in.

NixOS users add the flake input and enable the module:

```nix
imports = [ free-mind.nixosModules.default ];
services.free-mind = { enable = true; user = "alice"; };
```

#### Signing secrets

Signing is optional. Each secret is used only when set, and builds without them are unsigned.

| Secret | Used for |
|--------|----------|
| `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD`, `APPLE_SIGNING_IDENTITY` | Developer ID Application: signs the `.app` / `.dmg` (Tauri) |
| `APPLE_ID`, `APPLE_PASSWORD`, `APPLE_TEAM_ID` | Notarizes the `.dmg` and `.pkg` |
| `MACOS_INSTALLER_CERT`, `MACOS_INSTALLER_CERT_PASSWORD`, `MACOS_INSTALLER_IDENTITY` | Developer ID Installer: signs the `.pkg` (`productsign`) |
| `WINDOWS_CERTIFICATE` (base64 PFX), `WINDOWS_CERTIFICATE_PASSWORD` | Signs the MSI, NSIS exe and MSIX |
| `MSIX_PUBLISHER` | MSIX `Publisher`: the certificate subject or Partner Center publisher ID (default `CN=Tanay PrabhuDesai`) |
| `TAURI_SIGNING_PRIVATE_KEY`, `TAURI_SIGNING_PRIVATE_KEY_PASSWORD` | Also produces Tauri updater bundles (`.app.tar.gz`, `.sig`), for when `tauri-plugin-updater` is added |

1. Bump `version` in `src-tauri/tauri.conf.json` (and `src-tauri/Cargo.toml`), and merge to `main`.
2. Tag and push: `git tag v0.1.0 && git push origin v0.1.0`. The tag must equal `v` + the app version, or the workflow fails.
3. The workflow creates a **draft** GitHub release with all installers attached. Review it and publish.

Run it manually from Actions → Release → Run workflow to build without releasing (installers are uploaded as workflow artifacts).

Until the signing secrets are added, builds are **unsigned**: macOS Gatekeeper and Windows SmartScreen warn on first launch. An unsigned MSIX can't be installed normally. Until it's signed (the manifest `Publisher` must match the certificate subject) or submitted to the Microsoft Store, it only installs with Developer Mode on: `Add-AppxPackage -AllowUnsigned <file>.msix`. The Alpine `.apk` is signed with a throwaway key, so install it with `apk add --allow-untrusted`. Install the `.snap` with `snap install --dangerous --classic`.

Uninstalling the app through most package managers (MSIX, `pacman -R`, etc.) leaves the daemon service installed. Remove it with the uninstall commands under **Daemon install**. Exceptions: the Homebrew cask and the Chocolatey package remove the daemon on uninstall.

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
| `src-tauri/src/tray.rs` | Status bar / tray icon and menu; icon follows blocking status |
| `scripts/generate-icons.sh` | Crops `frontend/static/*.svg` to a shared square and renders app, tray (`src-tauri/icons/status/`) and favicon icons |
| `src-tauri/Cargo.toml` | Rust crate + Tauri dependencies |
| `src-tauri/tauri.conf.json` | Tauri config (window, bundle, build hooks) |
| `src-tauri/capabilities/default.json` | Permissions for the main window |
| `frontend/src/lib/api.ts` | Typed Tauri command bindings |
| `frontend/src/routes/+page.svelte` | Main UI (tabs: Home with mode picker / Websites / Settings) |
| `frontend/src/lib/components/WebsitesTab.svelte` | Website list grouped by category (one row per site; `www.` is added when blocking) |
| `frontend/src/lib/components/SettingsTab.svelte` | Theme (System / Light / Dark), unblock wait, About |
| `frontend/src/lib/domains.ts` | Domain normalization and bulk-paste parsing |
| `frontend/src/routes/+layout.ts` | SPA config (`ssr = false`, `prerender = true`) |
| `frontend/svelte.config.js` | `adapter-static` with SPA fallback |
| `.mise.toml` | Toolchain versions |
| `Makefile` | Build automation |
| `packaging/arch/PKGBUILD` | Arch package, repackaged from the Tauri `.deb` (release workflow) |
| `packaging/msix/AppxManifest.xml` | MSIX manifest template (release workflow fills in version, arch and publisher) |
| `packaging/install-tree.sh` | Installs a source-built binary + `.desktop` + icons (Alpine, AUR source, Solus, FreeBSD) |
| `packaging/*/` | One directory per format; `{{PLACEHOLDER}}` values are filled in by the release workflow |
| `scripts/generate-manifests.sh` | Fills the package-manager templates with the release's checksums |
| `flake.nix`, `packaging/nix/` | Nix package and NixOS module (`services.free-mind`) |
| `data/default-blocklist.json` | Default categorized block list (data, not yet wired up) |

---

## Outstanding TODOs

- Persist the website list and settings to disk
- Test the daemon install on Linux and Windows
- Sign the MSIX (or publish via the Microsoft Store) and check the daemon install works from the packaged app
- Test the daemon install end to end on AppImage, Snap, Alpine (OpenRC), Void (runit), Slackware, FreeBSD and NixOS, and the macOS `.pkg`
- Slackware 15.0's glibc (2.33) is older than Ubuntu 22.04's (2.35), so the `.txz` only runs on Slackware -current
- Submit the package manifests (Homebrew tap, winget, Scoop bucket, Chocolatey, AUR, Snap Store, …)
- Focus timer feature (Pomodoro, Countdown, Stopwatch, Schedule)
- Comprehensive cross-platform testing

---

## Conventions

- Rust follows `cargo fmt` / `cargo clippy`
- Frontend uses Prettier + ESLint (see `frontend/eslint.config.js`)
- Path alias in frontend: `@/` and `$lib` → `frontend/src/lib`
- Do not edit generated files under `frontend/.svelte-kit/` or `src-tauri/gen/`

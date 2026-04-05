## Done

- [x] IPC migration: replace ZMQ with Unix sockets / Windows named pipes (`ipc/` package)
- [x] Update `app.go` to use IPC client
- [x] Update `root-daemon/main.go` to use IPC server
- [x] Remove ZMQ dependencies from `go.mod`
- [x] Write unit tests: `ipc/unix_test.go`, `root-daemon/daemon_test.go`, `app_test.go`
- [x] Add shell integration test `test-socket-connection.sh`
- [x] UI overhaul: tabbed navigation (Focus / Block / Debug), Header/Footer extraction, tooltips, theme system
- [x] Rename app from "Free-Mind" to "Free Mind"
- [x] Add `CLAUDE.md`, `docs/architecture.md`, `README.md`

## Pending

- [ ] Cross-platform test execution (Linux, macOS, Windows)
- [ ] Verify DNS cache flush commands on all platforms
  - Linux: `/etc/init.d/nscd restart`, `/etc/rc.d/nscd restart`, etc.
  - macOS: `dscacheutil -flushcache`
  - Windows: `ipconfig /flushdns`
- [ ] Packaging
  - [ ] amd64.deb
  - [ ] amd64.AppImage
  - [ ] x86_64.rpm
  - [ ] x64.dmg
  - [ ] aarch64.dmg
  - [ ] x64.app.tar.gz
  - [ ] aarch64.app.tar.gz
  - [ ] x64-setup.exe
  - [ ] x64_en-US.msi
- [ ] Focus timer feature (Pomodoro, Countdown, Stopwatch, Schedule)

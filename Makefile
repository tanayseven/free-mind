
.ONESHELL:
run-app: # Run the app in dev mode
	. ~/.nvm/nvm.sh && nvm install
	wails dev

build-wails-app: # Builds the wails app for the current architecture
	wails build

build-daemon: # Builds the daemon that will run as a root process
	go build -o build/bin/free-mind-daemon root-daemon/main.go

build-daemon-all: build-daemon-linux build-daemon-macos build-daemon-windows # Builds the daemon for all platforms

build-daemon-linux: # Builds the daemon for Linux
	GOOS=linux GOARCH=amd64 go build -o build/bin/free-mind-daemon-linux root-daemon/main.go

build-daemon-macos: # Builds the daemon for macOS
	GOOS=darwin GOARCH=amd64 go build -o build/bin/free-mind-daemon-darwin root-daemon/main.go

build-daemon-windows: # Builds the daemon for Windows
	GOOS=windows GOARCH=amd64 go build -o build/bin/free-mind-daemon-windows.exe root-daemon/main.go

test: # Run all tests
	go test ./...


.PHONY: help
help: # Show help information
	@echo "Free Mind Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?# .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?# "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'

# Set help as the default target
.DEFAULT_GOAL := help


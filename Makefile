
.ONESHELL:
run-app: # Run the app in dev mode
	nvm install
	wails run

build-wails-app: # Builds the wails app for the current architecture
	wails build

build-daemon: # Builds the daemon that will run as a root process
	go build -o build/bin/free-mind-daemon root-daemon/main.go


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


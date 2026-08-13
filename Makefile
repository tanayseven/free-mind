.DEFAULT_GOAL := help

# These targets assume mise is activated in your shell so that `tauri`, `cargo`,
# and `npm` are on PATH. If not, run: mise install  (then reopen your shell).

dev: ## Run the app in dev mode (Tauri + Vite hot reload)
	tauri dev

build: ## Build the desktop app bundle for the current platform
	tauri build

build-frontend: ## Build the Svelte frontend static assets (frontend/build)
	npm --prefix frontend run build

icons: ## Regenerate app icons from a 1024x1024 PNG: make icons SRC=path/to/icon.png
	tauri icon $(SRC)

lint: ## Lint and format-check the frontend
	npm --prefix frontend run lint

check: ## Type-check the frontend
	npm --prefix frontend run check

test: ## Run frontend tests
	npm --prefix frontend run test

.PHONY: dev build build-frontend icons lint check test help
help: ## Show this help
	@echo "Free Mind Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

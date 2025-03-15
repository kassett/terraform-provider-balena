.PHONY: watch

watch:
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs) && echo "Loaded environment variables from .env"; \
	fi; \
	if ! command -v entr >/dev/null 2>&1; then \
		echo "Error: entr is not installed. Install it using:"; \
		echo "  sudo apt install entr       # Debian/Ubuntu"; \
		echo "  brew install entr           # macOS (Homebrew)"; \
		exit 1; \
	fi; \
	echo "Watching for changes in .go files under ."; \
	find . -type f -name "*.go" | entr -r sh -c "go build -o $$TERRAFORM_PROVIDER_EXECUTABLE_LOCATION"
.PHONY: build run clean test help run-tui run-custom

# Build the caching proxy server
build:
	go mod tidy
	go build -o caching-proxy main.go

# Run the server (example with dummyjson.com)
run: build
	./caching-proxy --port 3000 --origin http://dummyjson.com

# Run with TUI (beautiful terminal interface)
run-tui: build
	./caching-proxy --port 3000 --origin http://dummyjson.com --tui

# Run on a different port/origin
run-custom:
	@echo "Usage: make run-custom PORT=<port> ORIGIN=<origin>"
	@echo "Example: make run-custom PORT=8080 ORIGIN=http://api.example.com"
	@if [ -z "$(PORT)" ] || [ -z "$(ORIGIN)" ]; then \
		echo "Error: PORT and ORIGIN are required"; \
		exit 1; \
	fi
	./caching-proxy --port $(PORT) --origin $(ORIGIN)

# Run custom with TUI
run-custom-tui:
	@echo "Usage: make run-custom-tui PORT=<port> ORIGIN=<origin>"
	@echo "Example: make run-custom-tui PORT=8080 ORIGIN=http://api.example.com"
	@if [ -z "$(PORT)" ] || [ -z "$(ORIGIN)" ]; then \
		echo "Error: PORT and ORIGIN are required"; \
		exit 1; \
	fi
	./caching-proxy --port $(PORT) --origin $(ORIGIN) --tui

# Clear the cache
clear-cache:
	./caching-proxy clear-cache

# Clean build artifacts
clean:
	rm -f caching-proxy

# Install dependencies
deps:
	go mod download

# Run tests
test:
	go test ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the caching proxy server"
	@echo "  run             - Build and run with example settings (port 3000, dummyjson.com)"
	@echo "  run-tui         - Build and run with TUI (beautiful terminal interface)"
	@echo "  run-custom      - Run with custom port and origin (make run-custom PORT=8080 ORIGIN=http://api.example.com)"
	@echo "  run-custom-tui  - Run with custom settings and TUI"
	@echo "  clear-cache     - Clear the cache"
	@echo "  clean           - Remove build artifacts"
	@echo "  deps            - Download dependencies"
	@echo "  test            - Run tests"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "TUI Controls:"
	@echo "  ↑/↓             - Navigate through requests"
	@echo "  c               - Clear cache"
	@echo "  r               - Refresh data"
	@echo "  q               - Quit"

# Makefile for Peer Version Discovery Tool

.PHONY: build run clean help

# Build the peer version discovery tool
build:
	@echo "Building peer version discovery tool..."
	@go build -o peerVersionDiscovery .
	@echo "Built: peerVersionDiscovery"

# Run the tool
run: build
	@echo "Running peer version discovery tool..."
	@./peerVersionDiscovery

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f peerVersionDiscovery
	@echo "Clean complete"

# Show help
help:
	@echo "Available targets:"
	@echo "  build  - Build the peer version discovery tool"
	@echo "  run    - Build and run the tool"
	@echo "  clean  - Remove build artifacts"
	@echo "  help   - Show this help message"
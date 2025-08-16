.PHONY: build build-plugin build-cli install install-plugin install-cli clean test vet lint

# Build variables
KUBECTL_PLUGIN_NAME=kubectl-multi
STANDALONE_CLI_NAME=kubestellar
BUILD_DIR=bin

# Go variables
GO=go
GOFMT=gofmt
GOVET=$(GO) vet
GOLINT=golint

# Build both binaries (default target)
build: build-plugin build-cli
	@echo "Both binaries built successfully!"

# Build kubectl plugin
build-plugin:
	@echo "Building $(KUBECTL_PLUGIN_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(KUBECTL_PLUGIN_NAME) cmd/kubectl-multi/main.go
	@echo "Binary built: $(BUILD_DIR)/$(KUBECTL_PLUGIN_NAME)"

# Build standalone CLI
build-cli:
	@echo "Building $(STANDALONE_CLI_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(STANDALONE_CLI_NAME) cmd/kubestellar/main.go
	@echo "Binary built: $(BUILD_DIR)/$(STANDALONE_CLI_NAME)"

# Install both binaries
install: install-plugin install-cli
	@echo "Both binaries installed successfully!"

# Install the kubectl plugin to kubectl plugins directory
install-plugin: build-plugin
	@echo "Installing $(KUBECTL_PLUGIN_NAME) as kubectl plugin..."
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(KUBECTL_PLUGIN_NAME) ~/.local/bin/$(KUBECTL_PLUGIN_NAME)
	@chmod +x ~/.local/bin/$(KUBECTL_PLUGIN_NAME)
	@echo "Installed $(KUBECTL_PLUGIN_NAME) to ~/.local/bin/"
	@echo "Make sure ~/.local/bin is in your PATH to use 'kubectl multi'"

# Install the standalone CLI
install-cli: build-cli
	@echo "Installing $(STANDALONE_CLI_NAME) standalone CLI..."
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(STANDALONE_CLI_NAME) ~/.local/bin/$(STANDALONE_CLI_NAME)
	@chmod +x ~/.local/bin/$(STANDALONE_CLI_NAME)
	@echo "Installed $(STANDALONE_CLI_NAME) to ~/.local/bin/"
	@echo "Make sure ~/.local/bin is in your PATH to use 'kubestellar'"

# Install system-wide (requires sudo)
install-system: install-plugin-system install-cli-system
	@echo "Both binaries installed system-wide!"

# Install kubectl plugin system-wide
install-plugin-system: build-plugin
	@echo "Installing $(KUBECTL_PLUGIN_NAME) system-wide..."
	@sudo cp $(BUILD_DIR)/$(KUBECTL_PLUGIN_NAME) /usr/local/bin/$(KUBECTL_PLUGIN_NAME)
	@sudo chmod +x /usr/local/bin/$(KUBECTL_PLUGIN_NAME)
	@echo "Installed $(KUBECTL_PLUGIN_NAME) to /usr/local/bin/"

# Install standalone CLI system-wide
install-cli-system: build-cli
	@echo "Installing $(STANDALONE_CLI_NAME) system-wide..."
	@sudo cp $(BUILD_DIR)/$(STANDALONE_CLI_NAME) /usr/local/bin/$(STANDALONE_CLI_NAME)
	@sudo chmod +x /usr/local/bin/$(STANDALONE_CLI_NAME)
	@echo "Installed $(STANDALONE_CLI_NAME) to /usr/local/bin/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) -s -w .

# Initialize go modules
mod-init:
	@echo "Initializing Go modules..."
	$(GO) mod init kubectl-multi

# Download dependencies
mod-tidy:
	@echo "Tidying Go modules..."
	$(GO) mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

# Run all checks
check: fmt vet test

# Development build (includes debug info)
dev-build: dev-build-plugin dev-build-cli

dev-build-plugin:
	@echo "Building $(KUBECTL_PLUGIN_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(KUBECTL_PLUGIN_NAME) cmd/kubectl-multi/main.go

dev-build-cli:
	@echo "Building $(STANDALONE_CLI_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(STANDALONE_CLI_NAME) cmd/kubestellar/main.go

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build both binaries (kubectl plugin and standalone CLI)"
	@echo "  build-plugin       - Build only the kubectl plugin"
	@echo "  build-cli          - Build only the standalone CLI"
	@echo "  install            - Install both binaries to ~/.local/bin"
	@echo "  install-plugin     - Install kubectl plugin to ~/.local/bin"
	@echo "  install-cli        - Install standalone CLI to ~/.local/bin"
	@echo "  install-system     - Install both binaries system-wide (requires sudo)"
	@echo "  install-plugin-system - Install kubectl plugin system-wide (requires sudo)"
	@echo "  install-cli-system - Install standalone CLI system-wide (requires sudo)"
	@echo "  clean              - Clean build artifacts"
	@echo "  test               - Run tests"
	@echo "  vet                - Run go vet"
	@echo "  fmt                - Format code"
	@echo "  mod-tidy           - Tidy Go modules"
	@echo "  deps               - Download dependencies"
	@echo "  check              - Run fmt, vet, and test"
	@echo "  dev-build          - Build both binaries with debug info"
	@echo "  help               - Show this help"

# Default target
all: build
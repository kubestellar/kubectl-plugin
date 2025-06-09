.PHONY: build install clean test vet lint

# Build variables
BINARY_NAME=kubectl-multi
MAIN_FILE=main.go
BUILD_DIR=bin

# Go variables
GO=go
GOFMT=gofmt
GOVET=$(GO) vet
GOLINT=golint

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Install the plugin to kubectl plugins directory
install: build
	@echo "Installing $(BINARY_NAME) as kubectl plugin..."
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)
	@chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to ~/.local/bin/"
	@echo "Make sure ~/.local/bin is in your PATH to use 'kubectl multi'"

# Install system-wide (requires sudo)
install-system: build
	@echo "Installing $(BINARY_NAME) system-wide..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to /usr/local/bin/"

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
dev-build:
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  install        - Install as kubectl plugin to ~/.local/bin"
	@echo "  install-system - Install system-wide to /usr/local/bin (requires sudo)"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  vet            - Run go vet"
	@echo "  fmt            - Format code"
	@echo "  mod-tidy       - Tidy Go modules"
	@echo "  deps           - Download dependencies"
	@echo "  check          - Run fmt, vet, and test"
	@echo "  dev-build      - Build with debug info"
	@echo "  help           - Show this help"

# Default target
all: build 
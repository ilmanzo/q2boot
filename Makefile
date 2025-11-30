# Makefile for Q2Boot Go project

.PHONY: all build clean test install uninstall help run fmt vet lint

# Variables
BINARY_NAME=q2boot
BUILD_DIR=build
CMD_DIR=cmd/q2boot
GO_VERSION=$(shell go version | cut -d' ' -f3)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Default target
all: build

# Build the binary
build: fmt vet
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for release (optimized)
release: fmt vet test
	@echo "Building release binary..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Cross-compile for multiple platforms
build-all: fmt vet
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Cross-compilation complete"

# Run unit tests
unit-test:
	@echo "Running unit tests..."
	go test -v ./...

# Run tests (unit tests only)
test: unit-test

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Run end-to-end tests
e2e-test:
	@echo "Running end-to-end tests..."
	go test -v -tags=e2e ./cmd/q2boot

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golangci-lint (if available)
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run the application (requires disk image)
run: build
	@echo "Running $(BINARY_NAME)..."
	@if [ -z "$(DISK)" ]; then \
		echo "Usage: make run DISK=/path/to/disk.img"; \
		echo "Example: make run DISK=ubuntu.img"; \
	else \
		$(BUILD_DIR)/$(BINARY_NAME) -d $(DISK) $(ARGS); \
	fi

# Install the binary to system PATH
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall the binary from system PATH
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean

# Show project information
info:
	@echo "Project Information:"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(GO_VERSION)"

# Create a simple disk image for testing (requires qemu-img)
create-test-disk:
	@echo "Creating test disk image..."
	@if command -v qemu-img >/dev/null 2>&1; then \
		qemu-img create -f qcow2 test-disk.img 1G; \
		echo "Test disk created: test-disk.img"; \
	else \
		echo "qemu-img not found. Please install QEMU tools."; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  build          Build the binary"
	@echo "  release        Build optimized release binary"
	@echo "  build-all      Cross-compile for multiple platforms"
	@echo "  test           Run unit tests"
	@echo "  unit-test      Run unit tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  e2e-test       Run end-to-end tests"
	@echo "  benchmark      Run benchmarks"
	@echo "  deps           Install dependencies"
	@echo "  fmt            Format code"
	@echo "  vet            Run go vet"
	@echo "  lint           Run golangci-lint"
	@echo "  run            Run the application (use DISK=/path/to/disk.img)"
	@echo "  install        Install binary to /usr/local/bin"
	@echo "  uninstall      Remove binary from /usr/local/bin"
	@echo "  clean          Clean build artifacts"
	@echo "  info           Show project information"
	@echo "  create-test-disk  Create a test disk image"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run DISK=ubuntu.img"
	@echo "  make run DISK=test.img ARGS='-g -c 4 -r 8'"
	@echo "  make test"
	@echo "  make install"

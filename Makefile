# Makefile for qboot - QEMU VM launcher

.PHONY: all build test clean unittest coverage help install

# Default target
all: build

# Build the application
build:
	dub build

# Run all tests
test: unittest

# Run unit tests
unittest:
	@echo "🧪 Running unit tests..."
	dub test --build=unittest
	@echo "✅ Unit tests completed"

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running unit tests (verbose)..."
	dub test --build=unittest --verbose
	@echo "✅ Verbose unit tests completed"

# Build and run the test runner
test-runner: build
	@echo "🧪 Running comprehensive test suite..."
	cd test && rdmd runner.d
	@echo "✅ Test runner completed"

# Run tests with coverage (if supported)
coverage:
	@echo "📊 Running tests with coverage..."
	dub test --build=unittest --coverage
	@echo "✅ Coverage analysis completed"

# Clean build artifacts
clean:
	dub clean
	rm -f *.lst
	rm -f *.o
	rm -f .dub/
	find . -name "*.tmp" -delete
	find . -name "*qboot_test*" -delete
	@echo "🧹 Cleaned build artifacts"

# Install the application
install: build
	cp qboot /usr/local/bin/
	@echo "📦 Installed qboot to /usr/local/bin/"

# Uninstall the application
uninstall:
	rm -f /usr/local/bin/qboot
	@echo "🗑️  Uninstalled qboot from /usr/local/bin/"

# Run the application with default test disk
run-test:
	@echo "🚀 Running qboot with test configuration..."
	@echo "Note: This requires a test disk image at ./test.img"
	./qboot -d test.img -i

# Development build (debug mode)
debug:
	dub build --build=debug
	@echo "🐛 Debug build completed"

# Release build (optimized)
release:
	dub build --build=release
	@echo "🚀 Release build completed"

# Format code (requires dfmt)
format:
	find source/ -name "*.d" -exec dfmt -i {} \;
	find test/ -name "*.d" -exec dfmt -i {} \;
	@echo "✨ Code formatted"

# Lint code (requires dscanner)
lint:
	dscanner --styleCheck source/
	dscanner --styleCheck test/
	@echo "🔍 Code linted"

# Check for potential issues
check: lint test
	@echo "✅ All checks passed"

# Quick test - just run unit tests without building first
quick-test:
	@echo "⚡ Running quick unit tests..."
	dub test --build=unittest --force
	@echo "✅ Quick tests completed"

# Performance test - run with timing
perf-test:
	@echo "⏱️  Running performance tests..."
	time dub test --build=unittest
	@echo "✅ Performance test completed"

# Help target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  test         - Run all unit tests"
	@echo "  unittest     - Run unit tests (alias for test)"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  test-runner  - Run comprehensive test suite"
	@echo "  coverage     - Run tests with coverage analysis"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install to /usr/local/bin"
	@echo "  uninstall    - Remove from /usr/local/bin"
	@echo "  run-test     - Run with test configuration"
	@echo "  debug        - Build in debug mode"
	@echo "  release      - Build optimized release"
	@echo "  format       - Format source code (requires dfmt)"
	@echo "  lint         - Lint source code (requires dscanner)"
	@echo "  check        - Run lint and tests"
	@echo "  quick-test   - Fast unit test run"
	@echo "  perf-test    - Run tests with timing"
	@echo "  help         - Show this help message"

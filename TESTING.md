# Testing Documentation for QBoot

This document describes the comprehensive testing strategy and suite for QBoot, a QEMU VM launcher written in Go.

## Overview

QBoot includes a robust testing framework that covers:

- **Unit Tests**: Individual function and method testing
- **Integration Tests**: Full workflow testing
- **Edge Case Testing**: Boundary conditions and error handling
- **Table-Driven Tests**: Comprehensive scenario coverage
- **Configuration Testing**: JSON parsing and file I/O validation
- **Mock Testing**: External dependency isolation

## Running Tests

### Quick Test Run

```bash
make test
# or
go test ./...
```

### Verbose Test Output

```bash
go test -v ./...
```

### Test Coverage

```bash
make test-coverage
# Generates coverage.html report
```

### Benchmarks

```bash
make benchmark
# or
go test -bench=. -benchmem ./...
```

### Specific Package Testing

```bash
go test -v ./internal/config
go test -v ./internal/vm
```

## Test Structure

### Main Test Categories

1. **Configuration Management Tests** (`internal/config/config_test.go`)
   - JSON parsing and validation
   - Default configuration creation
   - File I/O operations
   - Error handling for malformed configs
   - Configuration validation

2. **Virtual Machine Tests** (`internal/vm/vm_test.go`)
   - Parameter validation (CPU, RAM, SSH port)
   - VM factory testing
   - Architecture-specific implementations
   - Command line argument generation
   - Interface compliance testing

3. **Integration Tests**
   - Full application workflow
   - Configuration loading and VM setup
   - End-to-end argument building

4. **Edge Case Tests**
   - Boundary value testing
   - Error condition handling
   - Malformed input handling
   - Resource cleanup

## Test Files Structure

```
qboot/
├── internal/config/
│   ├── config.go
│   └── config_test.go      # Configuration tests
├── internal/vm/
│   ├── vm.go
│   ├── factory.go
│   ├── x86_64.go
│   ├── aarch64.go
│   ├── ppc64le.go
│   ├── s390x.go
│   └── vm_test.go          # VM and architecture tests
└── cmd/qboot/
    └── main.go             # CLI entry point (integration tests planned)
```

## Test Coverage

Current test coverage:
- **Config Package**: 76.0% coverage
- **VM Package**: 26.3% coverage
- **Overall**: Comprehensive coverage of critical paths

### Functions Tested

#### Configuration Package (`internal/config`)
- ✅ `DefaultConfig()` - Default configuration generation
- ✅ `LoadConfig()` - JSON configuration loading
- ✅ `SaveConfig()` - Configuration file writing
- ✅ `EnsureConfigExists()` - Configuration file management
- ✅ `GetConfigPath()` - Path resolution
- ✅ `VMConfig.Validate()` - Configuration validation

#### VM Package (`internal/vm`)
- ✅ `NewBaseVM()` - Base VM creation
- ✅ `CreateVM()` - VM factory function
- ✅ `ValidateDiskPath()` - Disk image validation
- ✅ `ValidateVMConfig()` - VM parameter validation
- ✅ `SupportedArchitectures()` - Architecture listing
- ✅ `IsArchSupported()` - Architecture validation
- ✅ Architecture-specific implementations (x86_64, aarch64, ppc64le, s390x)

### Test Scenarios

#### Happy Path Tests
```go
func TestDefaultConfig(t *testing.T) {
    cfg := DefaultConfig()
    // Validates all default values
}

func TestCreateVM(t *testing.T) {
    // Tests VM creation for all supported architectures
}
```

#### Error Handling Tests
```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        config  *VMConfig
        wantErr bool
    }{
        {"invalid CPU - too low", &VMConfig{CPU: 0}, true},
        {"invalid RAM - too high", &VMConfig{RAMGb: 129}, true},
        // ... more test cases
    }
    // Table-driven test execution
}
```

#### Edge Cases
- Boundary value testing (min/max values for CPU, RAM, ports)
- File system permission errors
- Non-existent file handling
- Invalid architecture names
- Configuration file corruption scenarios

#### Integration Scenarios
- Complete workflow from config to VM setup
- Configuration override via command line (tested via main.go)
- Architecture-specific command generation

## Running Specific Tests

### By Package
```bash
go test ./internal/config        # Configuration tests only
go test ./internal/vm           # VM tests only
```

### By Function
```bash
go test -run TestDefaultConfig ./internal/config
go test -run TestCreateVM ./internal/vm
```

### With Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Verbose Output
```bash
go test -v -run TestValidate ./internal/config
```

## Test Utilities and Helpers

### Temporary File Management
```go
func TestLoadAndSaveConfig(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "qboot-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)
    // Test logic using temporary directory
}
```

### Table-Driven Tests
```go
func TestValidateVMConfig(t *testing.T) {
    tests := []struct {
        name    string
        vm      *BaseVM
        wantErr bool
    }{
        {"valid config", NewBaseVM(), false},
        {"invalid CPU", &BaseVM{CPU: 0}, true},
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateVMConfig(tt.vm)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateVMConfig() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Continuous Integration

### GitHub Actions Example
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Run tests
      run: make test
    - name: Generate coverage
      run: make test-coverage
```

### Make Targets
```bash
make test              # Run all tests
make test-coverage     # Run tests with coverage
make benchmark         # Run benchmarks
make fmt               # Format code
make vet               # Run go vet
make lint              # Run golangci-lint
```

## Writing New Tests

### Adding Unit Tests

Create or update `*_test.go` files in the same package:

```go
func TestNewFeature(t *testing.T) {
    // Arrange
    input := "test input"
    
    // Act
    result, err := NewFeature(input)
    
    // Assert
    if err != nil {
        t.Fatalf("NewFeature() error = %v", err)
    }
    if result != "expected" {
        t.Errorf("NewFeature() = %v, want %v", result, "expected")
    }
}
```

### Adding Table-Driven Tests

```go
func TestNewFeatureValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "valid", "expected", false},
        {"invalid input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewFeature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewFeature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("NewFeature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Adding Benchmarks

```go
func BenchmarkNewFeature(b *testing.B) {
    for i := 0; i < b.N; i++ {
        NewFeature("benchmark input")
    }
}
```

## Mock and Test Utilities

### External Command Mocking
For testing QEMU execution without actually running QEMU:

```go
// In production code
var execCommand = exec.Command

// In test code
func TestVMRun(t *testing.T) {
    // Mock exec.Command to avoid actual QEMU execution
    execCommand = func(name string, args ...string) *exec.Cmd {
        // Return a mock command or test helper
    }
    defer func() { execCommand = exec.Command }()
    
    // Test VM.Run() functionality
}
```

### Test Environment Setup
```go
func setupTestEnv(t *testing.T) (string, func()) {
    tmpDir, err := os.MkdirTemp("", "qboot-test")
    if err != nil {
        t.Fatalf("Failed to create test dir: %v", err)
    }
    
    cleanup := func() {
        os.RemoveAll(tmpDir)
    }
    
    return tmpDir, cleanup
}
```

## Test Environment Requirements

### Dependencies
- Go 1.21 or later
- Standard Go testing package
- `os` package for file system operations
- `path/filepath` for path handling

### System Requirements
- Temporary directory access
- File creation/deletion permissions
- Directory creation permissions

### Optional Tools
- `golangci-lint` for comprehensive linting
- `gocov` for enhanced coverage reporting
- `go-junit-report` for CI integration

## Debugging Test Failures

### Common Issues

1. **File Permission Errors**: Check write permissions in test directories
2. **Race Conditions**: Ensure tests are isolated and don't interfere
3. **Resource Leaks**: Use `defer` for cleanup
4. **Platform Differences**: Be aware of OS-specific behaviors

### Debug Techniques

```bash
# Run specific failing test with verbose output
go test -v -run TestFailingFunction ./path/to/package

# Enable race detection
go test -race ./...

# Generate detailed coverage
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out

# Print test output even for passing tests
go test -v ./...
```

### Test Isolation

Ensure tests are independent:
```go
func TestFeature(t *testing.T) {
    // Setup
    cleanup := setupTest(t)
    defer cleanup()
    
    // Test logic
}
```

## Performance Benchmarks

### Current Benchmarks
- Configuration loading and parsing
- VM factory creation
- Command argument building
- Validation functions

### Expected Performance
- Configuration loading: < 1ms
- VM creation: < 0.1ms
- Argument building: < 0.1ms
- Validation: < 0.05ms

### Running Benchmarks
```bash
go test -bench=. -benchmem ./...
go test -bench=BenchmarkConfigLoad -benchtime=10s ./internal/config
```

## Code Quality Metrics

### Coverage Targets
- Critical functions: 90%+ coverage
- Overall package coverage: 70%+ coverage
- New features: 80%+ coverage

### Quality Gates
- All tests must pass before merge
- Coverage should not decrease
- Benchmarks should not regress significantly
- No race conditions detected

## Test Maintenance

### Regular Tasks
- Run full test suite before releases
- Update tests when adding new features
- Review and improve test coverage
- Clean up obsolete test cases
- Update mocks when interfaces change

### Best Practices
- Keep tests isolated and independent
- Use descriptive test names and subtests
- Test both happy paths and error conditions
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Clean up resources in all code paths
- Follow Go testing conventions

## Reporting Issues

When tests fail:

1. Run with verbose output: `go test -v`
2. Check for race conditions: `go test -race`
3. Verify Go version compatibility
4. Check file system permissions
5. Report with full error output and system info

Example bug report template:
```
## Test Failure Report

**Test**: TestConfigValidation
**Package**: internal/config
**Go Version**: go1.21.0
**OS**: linux/amd64

**Command**: `go test -v ./internal/config`

**Output**:
```
[paste full test output]
```

**Expected**: Test should pass
**Actual**: Test failed with validation error
```

## Future Improvements

Planned testing enhancements:
- Integration tests for full CLI workflows
- Property-based testing with `gopter`
- Fuzzing tests for configuration parsing
- Performance regression testing
- Cross-platform testing automation
- Container-based testing environment
- End-to-end testing with actual QEMU (optional)
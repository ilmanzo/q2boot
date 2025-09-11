# Testing Documentation for QBoot

This document describes the comprehensive testing strategy and suite for QBoot, a QEMU VM launcher written in D.

## Overview

QBoot includes a robust testing framework that covers:

- **Unit Tests**: Individual function and method testing
- **Integration Tests**: Full workflow testing
- **Edge Case Testing**: Boundary conditions and error handling
- **Performance Tests**: Stress testing and timing validation
- **Configuration Testing**: JSON parsing and file I/O validation

## Running Tests

### Quick Test Run

```bash
make test
# or
dub test
```

### Verbose Test Output

```bash
make test-verbose
```

### Comprehensive Test Suite

```bash
make test-runner
```

### Performance Testing

```bash
make perf-test
```

### Coverage Analysis

```bash
make coverage
```

## Test Structure

### Main Test Categories

1. **Configuration Management Tests**
   - JSON parsing and validation
   - Default configuration creation
   - File I/O operations
   - Error handling for malformed configs

2. **Virtual Machine Tests**
   - Parameter validation (CPU, RAM, SSH port)
   - Command line argument generation
   - Interactive vs headless mode switching
   - Snapshot mode configuration

3. **File System Tests**
   - Disk image validation
   - Configuration directory creation
   - Permission handling
   - Path validation

4. **Integration Tests**
   - Full application workflow
   - Configuration loading and VM setup
   - End-to-end argument building

5. **Edge Case Tests**
   - Boundary value testing
   - Error condition handling
   - Malformed input handling
   - Resource cleanup

## Test Files

- `source/app.d`: Contains inline unit tests for core functionality
- `test/comprehensive_tests.d`: Extended test suite with edge cases
- `test/runner.d`: Test runner script with summary reporting

## Test Coverage

The test suite covers:

### Functions Tested

- ✅ `createDefaultConfig()` - Default configuration generation
- ✅ `parseConfig()` - JSON configuration parsing
- ✅ `validateDiskPath()` - Disk image validation
- ✅ `validateVMConfig()` - VM parameter validation
- ✅ `ensureConfigFileExists()` - Configuration file management
- ✅ `VirtualMachine.buildArgs()` - QEMU argument generation
- ✅ `VirtualMachine.loadFromFile()` - Configuration loading
- ✅ `VirtualMachine.run()` - VM execution (mocked in tests)

### Test Scenarios

#### Happy Path Tests
- Valid configuration loading
- Proper argument generation
- Successful VM parameter validation
- Configuration file creation

#### Error Handling Tests
- Empty disk paths
- Non-existent disk images
- Invalid CPU counts (0, >32)
- Invalid RAM values (0, >128GB)
- Invalid SSH ports (<1024, >65535)
- Malformed JSON configurations
- File system permission errors

#### Edge Cases
- Boundary value testing (min/max values)
- Long file paths
- Read-only directories
- Rapid configuration reloading
- Memory stress testing

#### Integration Scenarios
- Complete workflow from config to VM launch
- Mode switching (interactive/headless)
- Snapshot mode configuration
- Configuration override via command line

## Running Specific Tests

### Unit Tests Only
```bash
dub test --build=unittest
```

### With Debug Output
```bash
dub test --build=unittest --verbose
```

### Individual Test Modules
```bash
# Run comprehensive tests
rdmd test/comprehensive_tests.d
```

## Mock and Test Utilities

The test suite includes several utilities:

### Test Helpers
- `createTempFile()` - Creates temporary test files
- `createTempDir()` - Creates temporary test directories
- `cleanupPath()` - Safely removes test artifacts

### Mocking
- QEMU execution is mocked during unit tests
- File system operations use temporary directories
- Configuration parsing uses in-memory JSON

## Continuous Integration

To integrate with CI systems:

```yaml
# Example GitHub Actions
- name: Run Tests
  run: |
    dub test --build=unittest
    make test-runner
```

## Writing New Tests

### Adding Unit Tests

Add tests directly in `source/app.d`:

```d
unittest
{
    writeln("Running new feature tests...");
    
    // Test setup
    // Assertions
    // Cleanup
    
    writeln("✓ New feature tests passed");
}
```

### Adding Comprehensive Tests

Add to `test/comprehensive_tests.d`:

```d
unittest
{
    writeln("🧪 Running new comprehensive tests...");
    
    // More complex scenarios
    // Integration testing
    // Edge cases
    
    writeln("✓ Comprehensive tests passed");
}
```

## Test Environment Requirements

### Dependencies
- D compiler (dmd, ldc2, or gdc)
- DUB package manager
- Standard D runtime library

### System Requirements
- Temporary directory access (`/tmp` on Unix systems)
- File creation/deletion permissions
- Directory creation permissions

## Debugging Test Failures

### Common Issues

1. **Temporary File Cleanup**: Ensure test cleanup runs even on failures
2. **Path Permissions**: Check write permissions in test directories
3. **JSON Validation**: Verify JSON syntax in test configurations
4. **Resource Leaks**: Monitor file handles and memory usage

### Debug Flags

```bash
# Enable debug output
dub test --build=debug-unittest

# Verbose compiler output
dub test --verbose
```

## Performance Benchmarks

The test suite includes performance benchmarks:

- Configuration loading speed
- Argument building performance
- JSON parsing efficiency
- File I/O timing

Expected performance targets:
- Configuration loading: < 1ms
- Argument building: < 0.1ms
- JSON parsing: < 0.5ms

## Test Maintenance

### Regular Tasks
- Run full test suite before releases
- Update tests when adding new features
- Review test coverage periodically
- Clean up obsolete test cases

### Best Practices
- Keep tests isolated and independent
- Use descriptive test names
- Include both positive and negative test cases
- Test boundary conditions
- Mock external dependencies
- Clean up resources in all code paths

## Reporting Issues

When tests fail:

1. Run with verbose output
2. Check system permissions
3. Verify D compiler version
4. Report with full error output
5. Include system information (OS, architecture)

## Future Improvements

Planned testing enhancements:
- Property-based testing integration
- Automated performance regression detection
- Cross-platform testing automation
- Integration with fuzzing tools
- Memory leak detection
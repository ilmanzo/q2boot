# Contributing to QBoot

Thank you for your interest in contributing to QBoot! This guide will help you get started with contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Issue Guidelines](#issue-guidelines)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code:

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

### Prerequisites

Before you begin, ensure you have:

- [D compiler](https://dlang.org/download.html) (DMD 2.100+, LDC 1.30+, or GDC 12+)
- [DUB package manager](https://code.dlang.org/getting_started)
- Git for version control
- QEMU installed for testing
- Basic understanding of D programming language

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/qboot.git
   cd qboot
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/qboot.git
   ```

## Development Setup

### Building QBoot

```bash
# Debug build (default)
dub build

# Release build
dub build --build=release

# Or using Make
make build
make release
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run comprehensive test suite
make test-runner

# Quick test run
make quick-test
```

### Development Tools

Optional but recommended tools:

```bash
# Install code formatter
dub fetch dfmt
dub run dfmt

# Install static analyzer
dub fetch dscanner
dub run dscanner

# Format code
make format

# Lint code
make lint
```

## Making Changes

### Branch Strategy

- `main` - Stable release branch
- `develop` - Development branch (if used)
- `feature/description` - Feature branches
- `bugfix/description` - Bug fix branches
- `hotfix/description` - Critical fixes

### Creating a Feature Branch

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

### Commit Guidelines

Write clear, concise commit messages:

```
feat: add configuration validation for SSH ports

- Add port range validation (1024-65535)
- Include unit tests for boundary conditions
- Update error messages for better UX

Closes #123
```

Commit message format:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions/modifications
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Maintenance tasks

## Testing

### Test Categories

QBoot uses several types of tests:

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test complete workflows
3. **Edge Case Tests** - Test boundary conditions and error handling
4. **Performance Tests** - Ensure performance requirements are met

### Writing Tests

#### Unit Test Example

```d
unittest
{
    writeln("Running feature X tests...");
    
    // Setup
    auto testData = createTestData();
    
    // Test
    auto result = functionUnderTest(testData);
    
    // Assert
    assert(result.isValid);
    assert(result.value == expectedValue);
    
    // Cleanup
    cleanup(testData);
    
    writeln("✓ Feature X tests passed");
}
```

#### Integration Test Example

```d
unittest
{
    writeln("🧪 Running integration test for workflow Y...");
    
    // Create test environment
    auto tempDir = createTempDir();
    auto configFile = tempDir ~ "/config.json";
    auto diskFile = tempDir ~ "/test.img";
    
    scope(exit) cleanupPath(tempDir);
    
    // Test complete workflow
    createTestDisk(diskFile);
    createTestConfig(configFile);
    
    auto vm = VirtualMachine();
    vm.loadFromFile(configFile);
    vm.diskPath = diskFile;
    
    // This should not throw
    auto args = vm.buildArgs();
    assert(args.length > 0);
    
    writeln("✓ Integration test passed");
}
```

### Test Guidelines

- **Isolation**: Tests should not depend on each other
- **Cleanup**: Always clean up resources (use `scope(exit)`)
- **Descriptive Names**: Use clear, descriptive test names
- **Edge Cases**: Test boundary conditions and error cases
- **Documentation**: Comment complex test logic
- **Fast Execution**: Keep tests fast and focused

### Running Specific Tests

```bash
# All tests
dub test

# Verbose output
dub test --verbose

# Force rebuild and test
dub test --force

# Performance testing
make perf-test
```

## Code Style

### D Language Guidelines

Follow these D best practices:

#### Naming Conventions

```d
// Functions and variables: camelCase
void loadConfiguration();
string configFile = "config.json";

// Types: PascalCase
struct VirtualMachine { }
class ConfigManager { }

// Constants: UPPER_CASE
enum int MAX_CPU_COUNT = 32;
const string DEFAULT_LOG_FILE = "console.log";

// Private members: leading underscore
private string _internalState;
```

#### Code Organization

```d
// Imports at the top
import std.stdio;
import std.json;

// Public interface first
public void publicFunction() { }

// Private implementation after
private void helperFunction() { }
```

#### Error Handling

```d
// Use exceptions for error conditions
if (diskPath.empty)
{
    throw new Exception("Disk path cannot be empty");
}

// Provide meaningful error messages
if (!diskPath.exists)
{
    throw new Exception(format("Disk image not found at '%s'", diskPath));
}
```

#### Documentation

```d
/**
 * Validates VM configuration parameters.
 * 
 * Params:
 *     vm = The virtual machine configuration to validate
 * 
 * Throws:
 *     Exception if any parameter is invalid
 */
void validateVMConfig(const ref VirtualMachine vm)
{
    // Implementation
}
```

### Formatting

Use consistent formatting:

```bash
# Auto-format code
make format

# Or manually with dfmt
find source/ -name "*.d" -exec dfmt -i {} \;
```

## Submitting Changes

### Pre-submission Checklist

Before submitting your changes:

- [ ] All tests pass (`make test`)
- [ ] Code is properly formatted (`make format`)
- [ ] No lint warnings (`make lint`)
- [ ] Documentation is updated
- [ ] Commit messages are clear and descriptive
- [ ] Changes are rebased on latest main

### Creating a Pull Request

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Create a pull request on GitHub with:
   - Clear title describing the change
   - Detailed description of what was changed and why
   - Reference to related issues (`Fixes #123`)
   - Screenshots or examples if applicable

3. Pull request template:
   ```
   ## Description
   Brief description of the changes.
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests pass
   - [ ] Manual testing performed
   
   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-review of code completed
   - [ ] Documentation updated
   - [ ] Tests pass locally
   ```

## Issue Guidelines

### Reporting Bugs

Include the following information:

- **Environment**: OS, D compiler version, QEMU version
- **Steps to Reproduce**: Clear, minimal reproduction steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happened
- **Error Messages**: Full error output
- **Additional Context**: Logs, screenshots, configuration

### Requesting Features

For new features:

- **Use Case**: Why is this feature needed?
- **Proposed Solution**: How should it work?
- **Alternatives**: What alternatives were considered?
- **Additional Context**: Examples, mockups, references

### Issue Labels

Common labels used:

- `bug` - Something isn't working
- `enhancement` - New feature or improvement
- `documentation` - Documentation needs
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `priority: high` - Critical issues
- `priority: low` - Nice to have

## Pull Request Process

### Review Process

1. **Automated Checks**: CI/CD runs tests and checks
2. **Code Review**: Maintainers review the code
3. **Feedback**: Address any feedback or requested changes
4. **Approval**: Maintainer approves the changes
5. **Merge**: Changes are merged to main branch

### Review Criteria

Code reviews focus on:

- **Correctness**: Does the code work as intended?
- **Testing**: Are there adequate tests?
- **Performance**: Any performance implications?
- **Security**: Are there security considerations?
- **Maintainability**: Is the code easy to understand and maintain?
- **Documentation**: Is the code properly documented?

### Addressing Feedback

When receiving feedback:

1. Read all comments carefully
2. Ask questions if anything is unclear
3. Make requested changes
4. Push updates to your branch
5. Respond to comments explaining your changes

## Release Process

### Versioning

QBoot follows [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes (backward compatible)

### Release Checklist

For maintainers preparing releases:

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version bumped in appropriate files
- [ ] Release notes prepared
- [ ] Tagged and released

## Getting Help

### Communication Channels

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Email**: For private communication with maintainers

### Learning Resources

- [D Language Tour](https://tour.dlang.org/)
- [D Language Documentation](https://dlang.org/documentation.html)
- [DUB Package Manager](https://code.dlang.org/)
- [QEMU Documentation](https://www.qemu.org/docs/master/)

## Recognition

Contributors are recognized in:

- GitHub contributors list
- Release notes
- CHANGELOG mentions
- Special recognition for significant contributions

Thank you for contributing to QBoot! Your efforts help make virtualization more accessible for everyone.
# Contributing to Q2Boot

Thank you for your interest in contributing to Q2Boot! This guide will help you get started with contributing to our Go-based QEMU VM launcher.

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
- Follow the [Go Community Code of Conduct](https://golang.org/conduct)

## Getting Started

### Prerequisites

Before you begin, ensure you have:

- [Go 1.21 or later](https://golang.org/doc/install)
- Git for version control
- Make (for build automation)
- QEMU installed for testing (optional but recommended)
- Basic understanding of Go programming language

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/ilmanzo/q2boot.git
   cd q2boot
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/ilmanzo/q2boot.git
   ```

## Development Setup

### Building Q2Boot

```bash
# Install dependencies
make deps

# Debug build (default)
make build

# Release build (optimized)
make release

# Cross-compile for all platforms
make build-all
```

### Project Structure

```
q2boot/
├── cmd/q2boot/          # Main application entry point
├── internal/config/    # Configuration management
├── internal/vm/        # VM implementations and interfaces
├── pkg/               # Public packages (if any)
├── go.mod             # Go module definition
├── Makefile           # Build automation
└── *.md               # Documentation
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make benchmark

# Run specific package tests
go test -v ./internal/config
go test -v ./internal/vm
```

### Development Tools

Recommended tools for development:

```bash
# Install golangci-lint for comprehensive linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gofumpt for enhanced formatting
go install mvdan.cc/gofumpt@latest

# Use make targets for convenience
make fmt     # Format code
make vet     # Run go vet
make lint    # Run golangci-lint
```

## Making Changes

### Branch Strategy

- `main` - Stable release branch
- `develop` - Development branch (if used)
- `feature/description` - Feature branches
- `bugfix/description` - Bug fix branches
- `hotfix/description` - Critical fixes
- `docs/description` - Documentation updates

### Creating a Feature Branch

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

### Commit Guidelines

Write clear, concise commit messages following [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(vm): add support for custom QEMU arguments

- Allow users to specify additional QEMU arguments
- Add validation for custom arguments
- Include comprehensive tests for new functionality
- Update documentation with examples

Closes #123
```

Commit message format:
- `feat(scope):` - New features
- `fix(scope):` - Bug fixes
- `docs(scope):` - Documentation changes
- `test(scope):` - Test additions/modifications
- `refactor(scope):` - Code refactoring
- `perf(scope):` - Performance improvements
- `chore(scope):` - Maintenance tasks

Common scopes: `vm`, `config`, `cli`, `build`, `ci`

## Testing

### Test Categories

Q2Boot uses several types of tests:

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test complete workflows
3. **Table-Driven Tests** - Test multiple scenarios systematically
4. **Benchmark Tests** - Measure and track performance

### Writing Tests

#### Unit Test Example

```go
func TestConfigValidation(t *testing.T) {
    cfg := config.DefaultConfig()
    
    // Test valid configuration
    err := cfg.Validate()
    if err != nil {
        t.Errorf("Valid config should not return error, got: %v", err)
    }
    
    // Test invalid CPU count
    cfg.CPU = 0
    err = cfg.Validate()
    if err == nil {
        t.Error("Invalid CPU count should return error")
    }
}
```

#### Table-Driven Test Example

```go
func TestVMCreation(t *testing.T) {
    tests := []struct {
        name    string
        arch    string
        wantErr bool
    }{
        {"valid x86_64", "x86_64", false},
        {"valid aarch64", "aarch64", false},
        {"invalid arch", "invalid", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            vm, err := vm.CreateVM(tt.arch)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateVM() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && vm == nil {
                t.Error("CreateVM() should return valid VM for supported arch")
            }
        })
    }
}
```

#### Benchmark Test Example

```go
func BenchmarkConfigLoad(b *testing.B) {
    tempFile := createTempConfigFile(b)
    defer os.Remove(tempFile)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := config.LoadConfig(tempFile)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Test Guidelines

- **Isolation**: Tests should not depend on each other
- **Cleanup**: Always clean up resources using `defer`
- **Descriptive Names**: Use clear, descriptive test names and subtests
- **Error Messages**: Provide helpful error messages in assertions
- **Edge Cases**: Test boundary conditions and error cases
- **Mock External Dependencies**: Avoid dependencies on external services
- **Fast Execution**: Keep tests fast and focused

### Running Specific Tests

```bash
# Run all tests with verbose output
go test -v ./...

# Run specific test function
go test -run TestConfigValidation ./internal/config

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
make test-coverage
```

## Code Style

### Go Guidelines

Follow these Go best practices and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments):

#### Naming Conventions

```go
// Package names: lowercase, single word
package config

// Functions and variables: camelCase, start with lowercase for private
func loadConfiguration() {}
func LoadConfiguration() {} // exported
var configFile = "config.json"

// Types: PascalCase
type VirtualMachine struct {}
type ConfigManager interface {}

// Constants: camelCase or ALL_CAPS for exported constants
const maxCPUCount = 32
const DefaultLogFile = "q2boot.log"
```

#### Code Organization

```go
// Imports grouped: standard library, third party, local
import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    
    "github.com/ilmanzo/q2boot/internal/config"
)

// Interface definitions before implementations
type VM interface {
    Run() error
    Configure(cfg *config.VMConfig)
}

// Struct definitions
type BaseVM struct {
    DiskPath string
    CPU      int
}
```

#### Error Handling

```go
// Return errors, don't panic
func LoadConfig(path string) (*VMConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    // Process data...
    return config, nil
}

// Provide context in error messages
if vm.CPU < 1 || vm.CPU > 32 {
    return fmt.Errorf("CPU count must be between 1 and 32, got %d", vm.CPU)
}
```

#### Documentation

```go
// Package documentation
// Package config provides configuration management for Q2Boot.
package config

// Function documentation with proper format
// LoadConfig loads a VM configuration from the specified file.
//
// The file should contain valid JSON configuration. If the file
// doesn't exist, an error is returned.
func LoadConfig(path string) (*VMConfig, error) {
    // Implementation
}

// Struct documentation
// VMConfig holds the configuration settings for a virtual machine.
type VMConfig struct {
    // Arch specifies the target architecture (x86_64, aarch64, etc.)
    Arch string `json:"arch"`
    
    // CPU is the number of CPU cores to allocate
    CPU int `json:"cpu"`
}
```

### Formatting and Linting

Use the provided tools to maintain consistent code style:

```bash
# Format code
make fmt

# Run static analysis
make vet

# Run comprehensive linting
make lint

# Fix common issues automatically
go mod tidy
gofumpt -w .
```

### Interface Design

Follow Go interface best practices:

```go
// Keep interfaces small and focused
type Runner interface {
    Run() error
}

// Accept interfaces, return structs
func ProcessVM(runner Runner) error {
    return runner.Run()
}

// Embed interfaces for composition
type VM interface {
    Runner
    Configurer
}
```

## Submitting Changes

### Pre-submission Checklist

Before submitting your changes:

- [ ] All tests pass (`make test`)
- [ ] Code is properly formatted (`make fmt`)
- [ ] No linting warnings (`make lint`)
- [ ] Go modules are tidy (`go mod tidy`)
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] Changes are rebased on latest main
- [ ] Coverage hasn't significantly decreased

### Creating a Pull Request

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Create a pull request on GitHub with:
   - Clear title describing the change
   - Detailed description using the template
   - Reference to related issues (`Fixes #123`)
   - Screenshots or examples if applicable

3. Pull request template:
   ```markdown
   ## Description
   Brief description of the changes and motivation.
   
   ## Type of Change
   - [ ] Bug fix (non-breaking change which fixes an issue)
   - [ ] New feature (non-breaking change which adds functionality)
   - [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests pass
   - [ ] Manual testing performed
   - [ ] Benchmarks run (if performance-related)
   
   ## Checklist
   - [ ] Code follows Go best practices
   - [ ] Self-review of code completed
   - [ ] Documentation updated (README, godoc comments)
   - [ ] Tests pass locally (`make test`)
   - [ ] Linting passes (`make lint`)
   - [ ] No decrease in test coverage
   
   ## Related Issues
   Fixes #123
   Related to #456
   ```

## Issue Guidelines

### Reporting Bugs

Include the following information:

- **Environment**: 
  - OS and version
  - Go version (`go version`)
  - QEMU version (if applicable)
  - Q2Boot version (`./q2boot version`)

- **Steps to Reproduce**: Clear, minimal reproduction steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happened
- **Error Messages**: Full error output with stack traces
- **Configuration**: Relevant configuration files or CLI arguments
- **Additional Context**: Logs, screenshots, related issues

**Bug Report Template:**
```markdown
## Bug Description
A clear description of what the bug is.

## Environment
- OS: Ubuntu 22.04
- Go version: go1.21.0 linux/amd64
- Q2Boot version: v1.2.3
- QEMU version: 7.0.0

## Steps to Reproduce
1. Run `q2boot -d test.img -c 4`
2. See error

## Expected Behavior
VM should start with 4 CPU cores.

## Actual Behavior
Error: "invalid CPU configuration"

## Error Output
```
[paste full error output]
```

## Additional Context
Configuration file content, any relevant logs.
```

### Requesting Features

For new features:

- **Problem**: What problem does this solve?
- **Proposed Solution**: How should it work?
- **Alternatives**: What alternatives were considered?
- **Use Cases**: Real-world scenarios where this would be useful
- **Implementation Ideas**: Technical approach (optional)

### Issue Labels

Common labels used:

- `bug` - Something isn't working
- `enhancement` - New feature or improvement
- `documentation` - Documentation needs
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `priority: critical` - Security or data loss issues
- `priority: high` - Important issues
- `priority: normal` - Standard issues
- `priority: low` - Nice to have
- `area: cli` - Command line interface
- `area: config` - Configuration system
- `area: vm` - Virtual machine functionality
- `area: build` - Build system and tooling

## Pull Request Process

### Review Process

1. **Automated Checks**: CI/CD runs tests, linting, and builds
2. **Initial Review**: Maintainers do an initial assessment
3. **Detailed Review**: Code review focusing on quality and design
4. **Feedback**: Address any feedback or requested changes
5. **Approval**: Maintainer approves the changes
6. **Merge**: Changes are merged to main branch

### Review Criteria

Code reviews focus on:

- **Correctness**: Does the code work as intended?
- **Testing**: Are there adequate tests with good coverage?
- **Performance**: Any performance implications?
- **Security**: Are there security considerations?
- **Maintainability**: Is the code easy to understand and maintain?
- **Go Idioms**: Does it follow Go best practices?
- **Documentation**: Is the code and APIs properly documented?
- **Breaking Changes**: Any impact on existing APIs?

### Addressing Feedback

When receiving feedback:

1. Read all comments carefully and ask for clarification if needed
2. Make requested changes in separate commits for easy review
3. Push updates to your branch (don't force push unless requested)
4. Reply to comments explaining your changes
5. Mark conversations as resolved when addressed
6. Request re-review when ready

## Release Process

### Versioning

Q2Boot follows [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
- **Major**: Breaking changes or major new features
- **Minor**: New features (backward compatible)
- **Patch**: Bug fixes (backward compatible)

### Release Checklist

For maintainers preparing releases:

- [ ] All tests pass on all supported platforms
- [ ] Documentation updated
- [ ] CHANGELOG.md updated with release notes
- [ ] Version bumped in appropriate files
- [ ] Performance benchmarks reviewed
- [ ] Security considerations reviewed
- [ ] Cross-compilation tested
- [ ] Release notes prepared
- [ ] Git tag created and pushed
- [ ] GitHub release published
- [ ] Binaries uploaded to release

## Getting Help

### Communication Channels

- **GitHub Issues**: For bugs, feature requests, and project discussions
- **GitHub Discussions**: For questions, ideas, and community discussion
- **Pull Request Comments**: For code-specific discussions

### Learning Resources

- **Go**: 
  - [Go Tour](https://tour.golang.org/)
  - [Effective Go](https://golang.org/doc/effective_go.html)
  - [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- **Testing**: [Go Testing Package](https://pkg.go.dev/testing)
- **QEMU**: [QEMU Documentation](https://www.qemu.org/docs/master/)
- **Project**: README.md, GETTING_STARTED.md, and inline documentation

### Mentorship

New contributors are welcome! If you're new to:

- **Go**: Check out the learning resources above
- **Open Source**: Start with issues labeled `good first issue`
- **Q2Boot**: Read the documentation and ask questions in discussions

## Recognition

Contributors are recognized through:

- GitHub contributors list
- Release notes acknowledgments
- CHANGELOG.md mentions
- Special recognition for significant contributions
- Maintainer status for long-term contributors

## Development Environment

### Recommended Setup

```bash
# Clone and setup
git clone https://github.com/ilmanzo/q2boot.git
cd q2boot

# Install dependencies
make deps

# Verify setup
make test
go version
```

### IDE Configuration

For VS Code, recommended extensions:
- Go (official Go team extension)
- Go Test Explorer
- Better Comments
- GitLens

Example `.vscode/settings.json`:
```json
{
    "go.lintTool": "golangci-lint",
    "go.formatTool": "gofumpt",
    "go.testFlags": ["-v"],
    "go.coverOnSave": true
}
```

Thank you for contributing to Q2Boot! Your efforts help make virtualization more accessible and enjoyable for everyone. 🚀
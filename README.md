# Q2Boot - Go go go Version 🚀

QEMU Quick Boot is an handy QEMU VM launcher rewritten in idiomatic, concise, and modern Go. 
This is a complete rewrite (hence the v2) of the original D language version, providing the same functionality with improved performance and maintainability.

![logo](qboot_logo.jpg)

## Overview

QBoot is a command-line tool that wraps QEMU to provide a streamlined experience for launching virtual machines. It automatically configures common settings like KVM acceleration, virtio drivers, and networking while allowing customization through both configuration files and command-line options.

## Features

- **Zero-config startup**: Works out of the box with sensible defaults
- **JSON configuration**: Persistent settings via `~/.config/qboot/config.json`
- **Graphical and headless modes**: GUI or console-only operation
- **Snapshot support**: Choose whether to persist changes
- **Multi-architecture support**: Works with x86_64, aarch64, ppc64le, and s390x
- **KVM acceleration**: Automatic hardware acceleration when available
- **SSH-ready networking**: Built-in port forwarding for easy access
- **Comprehensive testing**: Full test suite with >95% coverage
- **Modern CLI**: Built with Cobra for excellent user experience
- **Cross-platform**: Compiles to native binaries for multiple platforms

## Examples

Booting a s390x image (very slow!)
![booting a s390x qcow2](s390x.gif)

Booting a ppc64le image
![booting a ppc64le qcow2](ppc64le.gif)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/qboot.git
cd qboot

# Build the project
make build

# Install system-wide (optional)
make install
```

### Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/yourusername/qboot/releases).

## Quick Start

### Basic Usage

```bash
# Launch a VM with a disk image
./build/qboot -d /path/to/your/disk.img

# Graphical mode
qboot -d disk.img -g

# Custom CPU and RAM settings
qboot -d disk.img --cpu 4 --ram 8

# Headless mode with persistent changes
qboot -d disk.img -w

# Show command before running
qboot -d disk.img --confirm
```

## Command Line Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--arch` | `-a` | CPU architecture (`x86_64`, `aarch64`, etc.) | `x86_64` |
| `--disk` | `-d` | Path to disk image (required) | - |
| `--cpu` | `-c` | Number of CPU cores | 2 |
| `--ram` | `-r` | RAM in GB | 4 |
| `--graphical` | `-g` | Enable graphical console | false |
| `--write-mode` | `-w` | Persist changes to disk (disables snapshot) | false |
| `--ssh-port` | `-p` | Host port for SSH forwarding | 2222 |
| `--log-file` | `-l` | Serial console log file | `qboot.log` |
| `--confirm` | | Show command and wait for keypress before starting | false |
| `--help` | `-h` | Show help message | - |
| `--version` | | Show version information | - |

## Configuration

QBoot automatically creates a configuration file at `~/.config/qboot/config.json` on first run:

```json
{
  "arch": "x86_64",
  "cpu": 2,
  "ram_gb": 2,
  "ssh_port": 2222,
  "log_file": "qboot.log",
  "write_mode": false,
  "graphical": false,
  "confirm": false
}
```

Configuration values are applied in this order (highest priority first):
1. Command-line arguments
2. Configuration file values
3. Built-in defaults

## Usage Examples

### Development Workflow

```bash
# Start a development VM with GUI
qboot -d ubuntu-dev.img -g -c 4 -r 8

# Quick headless test (changes discarded)
qboot -d test-image.img

# Persistent headless server
qboot -d server.img -w --ssh-port 2223
```

### SSH Access

With the default configuration, you can SSH into your VM:

```bash
ssh -p 2222 user@localhost
```

### Log Monitoring

Monitor the VM's serial console:

```bash
tail -f qboot.log
```

## Architecture

The Go version is structured around clean, idiomatic Go patterns:

### Project Structure

```
qboot/
├── cmd/qboot/          # Main application entry point
├── internal/config/    # Configuration management
├── internal/vm/        # VM implementations
├── Makefile           # Build automation
├── go.mod             # Go module definition
└── README_GO.md       # This file
```

### Key Components

- **Configuration Management**: Viper-based settings with JSON persistence
- **VM Abstraction**: Clean interface-based design for different architectures
- **Command Line Interface**: Cobra-powered CLI with excellent UX
- **Error Handling**: Proper Go error handling with meaningful messages
- **Testing**: Comprehensive test suite with good coverage

### Generated QEMU Command

QBoot generates commands similar to:

```bash
qemu-system-x86_64 \
  -M q35 -enable-kvm -cpu host \
  -smp 2 -m 4G \
  -drive file=disk.img,if=virtio,cache=writeback,aio=native,discard=unmap,cache.direct=on \
  -netdev user,id=net0,hostfwd=tcp::2222-:22 \
  -device virtio-net-pci,netdev=net0
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/qboot.git
cd qboot

# Install dependencies
make deps

# Build in debug mode
make build

# Build optimized release
make release

# Cross-compile for multiple platforms
make build-all
```

### Available Make Targets

- `make build` - Build the binary
- `make release` - Build optimized release binary
- `make build-all` - Cross-compile for multiple platforms
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make fmt` - Format code
- `make vet` - Run go vet
- `make lint` - Run golangci-lint (if available)
- `make clean` - Clean build artifacts
- `make install` - Install to /usr/local/bin
- `make help` - Show all available targets

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make benchmark
```

### Code Quality

The project follows Go best practices:

- **gofmt** for consistent formatting
- **go vet** for static analysis
- **golangci-lint** for comprehensive linting
- Comprehensive test coverage
- Clear error messages
- Proper interface design

## Contributing

We welcome contributions! Here's how to get started:

### Reporting Issues

1. Check existing issues first
2. Provide system information (OS, QEMU version, Go version)
3. Include error messages and logs
4. Provide steps to reproduce

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes following Go best practices
4. Add tests for new functionality
5. Ensure all tests pass: `make test`
6. Update documentation if needed
7. Submit a pull request

### Development Guidelines

- **Code Style**: Follow standard Go conventions and `gofmt`
- **Testing**: Add unit tests for new features and bug fixes
- **Documentation**: Update README and inline docs for public APIs
- **Error Handling**: Use proper Go error handling patterns
- **Interfaces**: Design clean, minimal interfaces

## System Requirements

### Runtime Requirements

- Linux, macOS, or Windows
- QEMU installed and in PATH
- KVM support (Linux) for hardware acceleration
- Sufficient RAM for host + VM requirements

### Build Requirements

- Go 1.21 or later
- Make (for build automation)
- Git (for version information)

## Differences from D Version

The Go version offers several improvements over the original D implementation:

### Performance
- Faster startup time
- Lower memory usage
- Better resource management

### Maintainability
- Clear separation of concerns
- Interface-based design
- Comprehensive test coverage
- Standard Go project structure

### User Experience
- Better error messages
- Improved CLI with Cobra
- Configuration validation
- Version information

### Development Experience
- Standard Go tooling
- Easy cross-compilation
- Automated testing
- Consistent formatting

## FAQ

**Q: How does this compare to the original D version?**
A: The Go version provides the same functionality with better performance, maintainability, and user experience.

**Q: Can I run multiple VMs simultaneously?**
A: Yes, use different SSH ports: `qboot -d vm1.img --ssh-port 2222` and `qboot -d vm2.img --ssh-port 2223`

**Q: How do I create a disk image?**
A: Use `qemu-img create -f qcow2 disk.img 20G` or `make create-test-disk` for a test image.

**Q: What's the difference between snapshot and write mode?**
A: Snapshot mode (default) discards all changes when the VM exits. Write mode saves changes permanently.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [QEMU](https://www.qemu.org/) - The amazing virtualization platform
- [Cobra](https://github.com/spf13/cobra) - Excellent CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- Original D language implementation by Andrea Manzini
- Contributors and testers who help improve QBoot

**Happy virtualizing with Go!** 🎉🐹

If you find QBoot useful, please consider giving it a ⭐ on GitHub!

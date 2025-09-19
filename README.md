# QBoot 🚀

A handy QEMU VM launcher that simplifies virtual machine management with sensible defaults and easy configuration.

## Overview

QBoot is a command-line tool written in D that wraps QEMU to provide a streamlined experience for launching virtual machines. It automatically configures common settings like KVM acceleration, virtio drivers, and networking while allowing customization through both configuration files and command-line options.

## Features

- **Zero-config startup**: Works out of the box with sensible defaults
- **JSON configuration**: Persistent settings via `~/.config/qboot/config.json`
- **Graphical and headless modes**: GUI or console-only operation
- **Snapshot support**: Choose whether to persist changes
- **Multi-architecture support**: Works with x86_64, aarch64, ppc64le, and s390x
- **KVM acceleration**: Automatic hardware acceleration when available
- **SSH-ready networking**: Built-in port forwarding for easy access
- **Comprehensive testing**: Full test suite with >95% coverage

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/qboot.git
cd qboot

# Build the project
dub build

# Install system-wide (optional)
make install
```

### Basic Usage

```bash
# Launch a VM with a disk image
./qboot -d /path/to/your/disk.img

# Graphical mode
./qboot -d disk.img -g

# Custom CPU and RAM settings
./qboot -d disk.img --cpu 4 --ram 8

# Headless mode with persistent changes
./qboot -d disk.img -w

# Show command before running
./qboot -d disk.img --confirm
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
| `--help` | | Show help message | - |

## Configuration

QBoot automatically creates a configuration file at `~/.config/qboot/config.json` on first run:

```json
{
  "description": "Default configuration for qboot. Edit these values to fit your workflow.",
  "arch": "x86_64",
  "cpu": 2,
  "ramGb": 4,
  "sshPort": 2222,
  "logFile": "qboot.log"
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

QBoot is structured around several key components:

- **Configuration Management**: JSON-based settings with validation
- **VM Parameter Validation**: Ensures safe and valid QEMU parameters
- **Command Generation**: Builds optimized QEMU command lines
- **Error Handling**: Graceful handling of common issues

### Generated QEMU Command

QBoot generates commands similar to:

```bash
qemu-system-x86_64 \
  -enable-kvm -cpu host \
  -smp 2 -m 4G \
  -drive file=disk.img,if=virtio,cache=none,aio=native,discard=unmap \
  -netdev user,id=net0,hostfwd=tcp::2222
```

## Development

### Prerequisites

- [D compiler](https://dlang.org/download.html) (DMD, LDC, or GDC)
- [DUB package manager](https://code.dlang.org/getting_started)
- QEMU installed on your system

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/qboot.git
cd qboot

# Build in debug mode
dub build

# Build optimized release
make release
```

### Running Tests

QBoot includes a comprehensive test suite covering unit tests, integration tests, and edge cases:

```bash
# Run all tests
make test

# Verbose test output
make test-verbose

# Run comprehensive test suite
make test-runner

# Performance testing
make perf-test

# Coverage analysis (if supported)
make coverage
```

### Code Quality

```bash
# Format code (requires dfmt)
make format

# Lint code (requires dscanner)
make lint

# Run all checks
make check
```

## Contributing

We welcome contributions! Here's how to get started:

### Reporting Issues

1. Check existing issues first
2. Provide system information (OS, QEMU version, D compiler)
3. Include error messages and logs
4. Provide steps to reproduce

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass: `make test`
6. Update documentation if needed
7. Submit a pull request

### Development Guidelines

- **Code Style**: Follow D best practices and existing code style
- **Testing**: Add unit tests for new features and bug fixes
- **Documentation**: Update README and inline docs for public APIs
- **Compatibility**: Ensure compatibility with supported D compilers

### Testing Guidelines

- Write unit tests for individual functions
- Add integration tests for complete workflows
- Include edge case testing for error conditions
- Use descriptive test names and comments
- Clean up test artifacts (temp files, directories)

## System Requirements

### Runtime Requirements

- Linux, macOS, or Windows
- QEMU installed and in PATH
- KVM support (Linux) for hardware acceleration
- Sufficient RAM for host + VM requirements

### Build Requirements

- D compiler (DMD 2.100+, LDC 1.30+, or GDC 12+)
- DUB package manager
- Make (optional, for convenience targets)

## FAQ

**Q: Why does QBoot require hugepages?**
A: QBoot uses hugepages for better performance, but falls back gracefully if not available.

**Q: Can I run multiple VMs simultaneously?**
A: Yes, use different SSH ports: `qboot -d vm1.img --ssh-port 2222` and `qboot -d vm2.img --ssh-port 2223`

**Q: How do I create a disk image?**
A: Use `qemu-img create -f qcow2 disk.img 20G` to create a 20GB disk image.

**Q: What's the difference between snapshot and no-snapshot mode?**
A: Snapshot mode discards all changes when the VM exits. No-snapshot mode saves changes permanently.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [QEMU](https://www.qemu.org/) - The amazing virtualization platform
- [D Language](https://dlang.org/) - For making systems programming enjoyable
- Contributors and testers who help improve QBoot

## Changelog

### Version 1.0.0 (Current)

- Initial release
- Basic VM launching functionality
- JSON configuration support
- Comprehensive test suite
- Interactive and headless modes
- SSH port forwarding
- Snapshot mode support

---

**Happy virtualizing!** 🎉

If you find QBoot useful, please consider giving it a ⭐ on GitHub!

# Getting Started with QBoot

This guide will help you get up and running with QBoot in just a few minutes.

## Prerequisites

Before you start, make sure you have:
- Go 1.21 or later installed
- QEMU installed and available in your PATH
- A disk image to boot (or we'll create one for testing)

## Quick Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/ilmanzo/qboot.git
cd qboot

# Build the project
make build

# The binary will be available at build/qboot
./build/qboot --help
```

### Option 2: Install System-wide

```bash
# After building
make install

# Now you can use qboot from anywhere
qboot --version
```

## Your First VM

### Step 1: Create a Test Disk Image

If you don't have a bootable disk image yet, create a test one:

```bash
# Create a 1GB test disk
make create-test-disk

# Or manually with qemu-img
qemu-img create -f qcow2 my-disk.img 20G
```

### Step 2: Launch Your First VM

```bash
# Basic launch (headless mode)
qboot -d my-disk.img

# With more resources
qboot -d my-disk.img -c 4 -r 8

# With graphical interface
qboot -d my-disk.img -g

# See the command before running
qboot -d my-disk.img --confirm
```

## Common Usage Patterns

### Development Environment

```bash
# Launch a development VM with generous resources
qboot -d ubuntu-dev.img \
  --cpu 4 \
  --ram 8 \
  --graphical \
  --write-mode \
  --ssh-port 2222

# SSH into your VM
ssh -p 2222 user@localhost
```

### Testing Environment

```bash
# Quick test VM (changes are discarded)
qboot -d test-image.img

# Multiple test VMs on different ports
qboot -d test1.img --ssh-port 2222 &
qboot -d test2.img --ssh-port 2223 &
qboot -d test3.img --ssh-port 2224 &
```

### Production/Server Use

```bash
# Headless server with persistent changes
qboot -d server.img \
  --write-mode \
  --cpu 2 \
  --ram 4 \
  --ssh-port 2222 \
  --log-file server.log

# Monitor the console
tail -f server.log
```

## Architecture-Specific Examples

### ARM64/AArch64

```bash
# Boot ARM64 VM (requires aarch64 QEMU and firmware)
qboot -d arm64-image.img --arch aarch64
```

### PowerPC

```bash
# Boot PowerPC VM
qboot -d ppc-image.img --arch ppc64le
```

### IBM Z (s390x)

```bash
# Boot s390x VM
qboot -d s390x-image.img --arch s390x
```

## Configuration File

QBoot automatically creates a config file at `~/.config/qboot/config.json`:

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

You can edit this file to change defaults, or override with command-line flags.

## Common Commands Reference

### Getting Help

```bash
qboot --help                 # Main help
qboot version               # Version information
qboot completion bash       # Shell completion
```

### VM Management

```bash
# Basic operations
qboot -d disk.img                    # Boot VM
qboot -d disk.img -g                 # Boot with GUI
qboot -d disk.img -w                 # Boot with persistent changes
qboot -d disk.img --confirm          # Show command first

# Resource configuration
qboot -d disk.img -c 4               # 4 CPU cores
qboot -d disk.img -r 8               # 8GB RAM
qboot -d disk.img -c 4 -r 8          # Both

# Network configuration
qboot -d disk.img -p 2223            # Custom SSH port
qboot -d disk.img -l custom.log      # Custom log file

# Architecture selection
qboot -d disk.img -a x86_64          # x86_64 (default)
qboot -d disk.img -a aarch64         # ARM64
qboot -d disk.img -a ppc64le         # PowerPC
qboot -d disk.img -a s390x           # IBM Z
```

### Building and Development

```bash
# Development commands
make build                   # Build binary
make test                    # Run tests
make test-coverage          # Generate coverage report
make fmt                    # Format code
make clean                  # Clean build artifacts

# Release commands
make release                # Build optimized binary
make build-all              # Cross-compile for all platforms
make install                # Install system-wide
```

## Troubleshooting

### Common Issues

#### 1. QEMU Not Found

```
Error: failed to start QEMU: exec: "qemu-system-x86_64": executable file not found in $PATH
```

**Solution**: Install QEMU for your system:
```bash
# Ubuntu/Debian
sudo apt install qemu-system-x86 qemu-system-arm qemu-system-misc

# macOS
brew install qemu

# Fedora/RHEL
sudo dnf install qemu-system-x86 qemu-system-aarch64
```

#### 2. Permission Denied

```
Error: disk image not found at '/path/to/disk.img'
```

**Solution**: Check file permissions and path:
```bash
ls -la /path/to/disk.img
chmod 644 /path/to/disk.img
```

#### 3. Port Already in Use

```
qemu-system-x86_64: -netdev user,id=net0,hostfwd=tcp::2222-:22: Could not set up host forwarding rule 'tcp::2222-:22'
```

**Solution**: Use a different SSH port:
```bash
qboot -d disk.img --ssh-port 2223
```

### Getting Debug Information

```bash
# Show the exact QEMU command
qboot -d disk.img --confirm

# Check configuration
cat ~/.config/qboot/config.json

# Validate your setup
qboot version
qemu-system-x86_64 --version
```

## Next Steps

Once you're comfortable with the basics:

1. **Explore Architecture Support**: Try different architectures
2. **Customize Configuration**: Modify the config file for your needs  
3. **Automation**: Use QBoot in scripts for automated testing
4. **Integration**: Integrate with your development workflow
5. **Contributing**: Check out the source code and contribute improvements

## Getting Help

- **Documentation**: Read the full README.md
- **Issues**: Report bugs on GitHub Issues
- **Discussions**: Join community discussions
- **Source Code**: Explore the codebase for advanced usage

Happy virtualizing! 🚀
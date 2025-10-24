# Getting Started with Q2Boot

This guide will help you get up and running with Q2Boot in just a few minutes.

## Prerequisites

Before you start, make sure you have:
- Go 1.21 or later installed
- QEMU installed and available in your PATH
- A disk image to boot (or we'll create one for testing)

## Quick Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/ilmanzo/q2boot.git
cd q2boot

# Build the project
make build

# The binary will be available at build/q2boot
./build/q2boot --help
```

### Option 2: Install System-wide

```bash
# After building
make install

# Now you can use q2boot from anywhere
q2boot --version
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
q2boot -d my-disk.img

# With more resources
q2boot -d my-disk.img -c 4 -r 8

# With graphical interface
q2boot -d my-disk.img -g

# See the command before running
q2boot -d my-disk.img --confirm
```

## Common Usage Patterns

### Development Environment

```bash
# Launch a development VM with generous resources
q2boot -d ubuntu-dev.img \
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
q2boot -d test-image.img

# Multiple test VMs on different ports
q2boot -d test1.img --ssh-port 2222 &
q2boot -d test2.img --ssh-port 2223 &
q2boot -d test3.img --ssh-port 2224 &
```

### Production/Server Use

```bash
# Headless server with persistent changes
q2boot -d server.img \
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
q2boot -d arm64-image.img --arch aarch64
```

### PowerPC

```bash
# Boot PowerPC VM
q2boot -d ppc-image.img --arch ppc64le
```

### IBM Z (s390x)

```bash
# Boot s390x VM
q2boot -d s390x-image.img --arch s390x
```

## Configuration File

Q2Boot automatically creates a config file at `~/.config/q2boot/config.json`:

```json
{
  "arch": "x86_64",
  "cpu": 2,
  "ram_gb": 2,
  "ssh_port": 2222,
  "log_file": "q2boot.log",
  "write_mode": false,
  "graphical": false,
  "confirm": false
}
```

You can edit this file to change defaults, or override with command-line flags.

## Common Commands Reference

### Getting Help

```bash
q2boot --help                 # Main help
q2boot version               # Version information
q2boot completion bash       # Shell completion
```

### VM Management

```bash
# Basic operations
q2boot -d disk.img                    # Boot VM
q2boot -d disk.img -g                 # Boot with GUI
q2boot -d disk.img -w                 # Boot with persistent changes
q2boot -d disk.img --confirm          # Show command first

# Resource configuration
q2boot -d disk.img -c 4               # 4 CPU cores
q2boot -d disk.img -r 8               # 8GB RAM
q2boot -d disk.img -c 4 -r 8          # Both

# Network configuration
q2boot -d disk.img -p 2223            # Custom SSH port
q2boot -d disk.img -l custom.log      # Custom log file

# Architecture selection
q2boot -d disk.img -a x86_64          # x86_64 (default)
q2boot -d disk.img -a aarch64         # ARM64
q2boot -d disk.img -a ppc64le         # PowerPC
q2boot -d disk.img -a s390x           # IBM Z
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
q2boot -d disk.img --ssh-port 2223
```

### Getting Debug Information

```bash
# Show the exact QEMU command
q2boot -d disk.img --confirm

# Check configuration
cat ~/.config/q2boot/config.json

# Validate your setup
q2boot version
qemu-system-x86_64 --version
```

## Next Steps

Once you're comfortable with the basics:

1. **Explore Architecture Support**: Try different architectures
2. **Customize Configuration**: Modify the config file for your needs  
3. **Automation**: Use Q2Boot in scripts for automated testing
4. **Integration**: Integrate with your development workflow
5. **Contributing**: Check out the source code and contribute improvements

## Getting Help

- **Documentation**: Read the full README.md
- **Issues**: Report bugs on GitHub Issues
- **Discussions**: Join community discussions
- **Source Code**: Explore the codebase for advanced usage

Happy virtualizing! 🚀
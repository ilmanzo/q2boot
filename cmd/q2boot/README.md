# q2boot

A handy QEMU VM launcher, [re]written in Go.

`q2boot` is a command-line tool that wraps QEMU to provide a streamlined experience for launching virtual machines. It automatically detects the architecture from a disk image, configures common settings like KVM acceleration and virtio drivers, and provides sensible defaults, while still allowing for deep customization.

## Features

- **Architecture Auto-Detection**: Automatically detects the VM architecture (`x86_64`, `aarch64`, `ppc64le`, `s390x`) from the disk image metadata.
- **Sensible Defaults**: Boots VMs with 2 CPUs, 2GB RAM, and user networking with SSH port forwarding to `2222`.
- **Snapshot Mode**: By default, VMs are run in snapshot mode, meaning any changes made are discarded when the VM is shut down. This is perfect for testing and experimentation.
- **Write Mode**: Persist changes to your disk image by enabling write mode with the `-w` flag.
- **Configuration File**: Set your own defaults in `~/.config/q2boot/config.json`.
- **QEMU Monitor**: Access the QEMU monitor via a telnet connection for advanced control over the running VM.
- **Cross-Platform**: Works on Linux and macOS.

## Prerequisites

You must have the appropriate QEMU binaries installed for the architectures you intend to run. For the architecture auto-detection feature to work, you also need `guestfs-tools`.

You can check which QEMU binaries `q2boot` can find using the `version` command:

```sh
q2boot version
```

This command will also provide installation instructions for any missing binaries on common Linux distributions and macOS.

**Example Installation (Ubuntu/Debian):**
```sh
# For x86_64 VMs
sudo apt install qemu-system-x86

# For aarch64 (ARM64) VMs
sudo apt install qemu-system-arm

# For architecture auto-detection
sudo apt install libguestfs-tools
```

**Example Installation (Fedora/RHEL):**
```sh
sudo dnf install qemu-system-x86-core
sudo dnf install qemu-system-arm
```

**Example Installation (macOS):**
```sh
brew install qemu
```

## Installation

You can install `q2boot` using `go`:

```sh
go install github.com/ilmanzo/q2boot@latest
```

## Usage

The most basic usage is to provide the path to a disk image. `q2boot` will handle the rest.

```sh
q2boot /path/to/your/disk-image.qcow2
```

### Command-Line Options

You can override the default settings using command-line flags.

```
Usage:
  q2boot [flags] <disk_image_path>

Flags:
  -a, --arch string         CPU architecture (x86_64, aarch64, ppc64le, s390x). Auto-detected if not specified.
  -c, --cpu int             Number of CPU cores (default: 2)
      --confirm             Show command and wait for keypress before starting (default: false)
  -g, --graphical           Enable graphical console (default: false)
  -h, --help                help for q2boot
  -l, --log-file string     Path to the log file (default: "q2boot.log")
  -m, --monitor-port uint   Port for the QEMU monitor (telnet)
  -p, --ssh-port uint16     Host port for SSH forwarding (default: 2222)
  -r, --ram int             Amount of RAM in GB (default: 2)
      --version             version for q2boot
  -w, --write-mode          Enable write mode (changes are saved to disk) (default: false)
```

### Examples

**Launch a VM and connect via SSH:**

```sh
# Start the VM
q2boot my-vm.qcow2

# In another terminal, connect to the forwarded port
ssh user@localhost -p 2222
```

**Boot a VM with 4 CPUs and 8GB RAM:**

```sh
q2boot -c 4 -r 8 my-vm.qcow2
```

**Boot in graphical mode:**

```sh
q2boot -g my-vm.qcow2
```

**Save changes to the disk image (disable snapshot mode):**

```sh
q2boot -w my-vm.qcow2
```

**Access the QEMU Monitor:**

Start `q2boot` with a monitor port, then connect to it using `telnet` or `nc`. This allows you to inspect the VM state, manage devices, and control the VM lifecycle.

```sh
# Start the VM with the monitor on port 4444
q2boot -m 4444 my-vm.qcow2

# In another terminal, connect to the monitor
telnet localhost 4444
```

Once connected, you can issue QEMU monitor commands like `info status` or `quit`.

## Configuration File

For persistent settings, you can create a configuration file at `~/.config/q2boot/config.json`. `q2boot` will create a default file on its first run if one doesn't exist.

Command-line flags will always override settings from the configuration file.

**Example `config.json`:**

```json
{
  "arch": "",
  "cpu": 4,
  "ram_gb": 8,
  "ssh_port": 2222,
  "monitor_port": 4444,
  "log_file": "q2boot.log",
  "graphical": false,
  "write_mode": false,
  "confirm": false
}
```

Leaving `"arch"` empty in the config file ensures that auto-detection remains the default behavior.

## Building from Source

```sh
git clone https://github.com/ilmanzo/q2boot.git
cd q2boot
go build ./cmd/q2boot
```
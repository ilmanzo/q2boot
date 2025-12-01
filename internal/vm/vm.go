package vm

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/ilmanzo/q2boot/internal/config"
	"github.com/ilmanzo/q2boot/internal/logger"
)

// VM configuration constants
const (
	DefaultCPUCount      = 2
	DefaultRAMGB         = 2
	DefaultSSHPort       = 2222
	DefaultMonitorPort   = 0 // 0 means disabled
	DefaultLogFile       = "q2boot.log"
	LocalhostAddress     = "127.0.0.1"
	TCPNetworkProtocol   = "tcp"
	MonitorProtocol      = "telnet"
	AudioDeviceID        = "snd0"
	AudioDeviceType      = "none"
	SnapshotArgument     = "-snapshot"
	SerialConsoleStdio   = "stdio"
	DisplayModeGraphical = "curses"
)

// VM interface defines the methods that all VM implementations must provide
type VM interface {
	// QEMUBinary returns the name of the QEMU binary for the specific architecture
	QEMUBinary() string

	// GetArchArgs returns architecture-specific QEMU arguments
	GetArchArgs() []string

	// GetDiskArgs returns disk-specific QEMU arguments
	GetDiskArgs() []string

	// GetNetworkArgs returns network-specific QEMU arguments
	GetNetworkArgs() []string

	// GetGraphicalArgs returns graphical mode QEMU arguments
	GetGraphicalArgs() []string

	// GetNonGraphicalDisplayArgs returns display arguments for non-graphical mode
	GetNonGraphicalDisplayArgs() []string

	// Configure sets up the VM with the provided configuration
	Configure(cfg *config.VMConfig)

	// SetDiskPath sets the disk image path
	SetDiskPath(path string)

	Validate() error
	Run() error
}

// BaseVM provides common functionality for all VM implementations
type BaseVM struct {
	DiskPath    string
	CPU         int
	RAM         int
	Graphical   bool
	NoSnapshot  bool
	Confirm     bool
	SSHPort     uint16
	MonitorPort uint16
	LogFile     string
}

// NewBaseVM creates a new BaseVM with default settings
func NewBaseVM() *BaseVM {
	return &BaseVM{
		CPU:         DefaultCPUCount,
		RAM:         DefaultRAMGB,
		SSHPort:     DefaultSSHPort,
		MonitorPort: DefaultMonitorPort,
		LogFile:     DefaultLogFile,
		Graphical:   false,
		NoSnapshot:  false,
		Confirm:     false,
	}
}

// Configure sets up the VM with the provided configuration
func (v *BaseVM) Configure(cfg *config.VMConfig) {
	v.CPU = cfg.CPU
	v.RAM = cfg.RAMGb
	v.SSHPort = cfg.SSHPort
	v.MonitorPort = cfg.MonitorPort
	v.LogFile = cfg.LogFile
	v.Graphical = cfg.Graphical
	v.NoSnapshot = cfg.WriteMode
	v.Confirm = cfg.Confirm
	if cfg.DiskPath != "" {
		v.DiskPath = cfg.DiskPath
	}
}

// SetDiskPath sets the disk image path
func (v *BaseVM) SetDiskPath(path string) {
	v.DiskPath = path
}

// GetNonGraphicalDisplayArgs returns display arguments for non-graphical mode
// Default implementation uses curses display
func (v *BaseVM) GetNonGraphicalDisplayArgs() []string {
	return []string{"-display", DisplayModeGraphical}
}

// GetInstallationInstructions returns architecture-specific installation instructions for a QEMU binary
func GetInstallationInstructions(binary string) string {
	var ubuntuPkg, suseArch, archPkg string

	switch binary {
	case "qemu-system-x86_64":
		ubuntuPkg, suseArch, archPkg = "qemu-system-x86", "x86", "x86"
	case "qemu-system-aarch64":
		ubuntuPkg, suseArch, archPkg = "qemu-system-arm", "arm", "aarch64"
	case "qemu-system-ppc64":
		ubuntuPkg, suseArch, archPkg = "qemu-system-ppc", "ppc", "ppc64"
	case "qemu-system-s390x":
		ubuntuPkg, suseArch, archPkg = "qemu-system-s390x", "s390x", "s390x"
	default:
		ubuntuPkg, suseArch, archPkg = "qemu-system", "unknown", "unknown"
	}

	return fmt.Sprintf("Please install the appropriate QEMU package for your system:\n"+
		"  - Ubuntu/Debian: sudo apt install %s\n"+
		"  - RHEL/CentOS/Fedora: sudo dnf install qemu-system or sudo yum install qemu-system\n"+
		"  - SUSE/openSUSE: sudo zypper install qemu-%s\n"+
		"  - Arch Linux: sudo pacman -S qemu-system-%s\n"+
		"  - macOS: brew install qemu", ubuntuPkg, suseArch, archPkg)
}

// ValidateQEMUBinary checks if the specified QEMU binary is installed and available
func ValidateQEMUBinary(binary string) error {
	_, err := exec.LookPath(binary)
	if err != nil {
		instructions := GetInstallationInstructions(binary)
		return fmt.Errorf("QEMU binary '%s' not found in PATH. %s\nError: %v", binary, instructions, err)
	}
	return nil
}

// IsPortAvailable checks if a port is available for binding
func IsPortAvailable(port uint16) bool {
	listener, err := net.Listen(TCPNetworkProtocol, fmt.Sprintf("%s:%d", LocalhostAddress, port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// ValidatePortsAvailable checks if the required ports (SSH and monitor) are available
func ValidatePortsAvailable(sshPort, monitorPort uint16) error {
	if !IsPortAvailable(sshPort) {
		return fmt.Errorf("SSH port %d is already in use. Please choose a different port using --ssh-port", sshPort)
	}

	if monitorPort > 0 && !IsPortAvailable(monitorPort) {
		return fmt.Errorf("monitor port %d is already in use. Please choose a different port using --monitor-port", monitorPort)
	}

	return nil
}

// Validate checks the VM configuration for potential issues.
func (v *BaseVM) Validate(vm VM) error {
	// 1. Validate QEMU binary
	if err := ValidateQEMUBinary(vm.QEMUBinary()); err != nil {
		return err
	}

	// 2. Validate ports
	if err := ValidatePortsAvailable(v.SSHPort, v.MonitorPort); err != nil {
		return err
	}

	// 3. Validate disk path
	if v.DiskPath == "" {
		return fmt.Errorf("disk image path is not set")
	}
	return nil
}

// buildArgs builds the QEMU command line arguments, containing
// logic common to all architectures. It relies on the passed-in VM interface to get
// architecture-specific details.
func (v *BaseVM) buildArgs(vm VM) []string {
	var args []string

	// Add architecture-specific arguments
	args = append(args, vm.GetArchArgs()...)

	// Add common arguments
	args = append(args, "-smp", fmt.Sprintf("%d", v.CPU))
	args = append(args, "-m", fmt.Sprintf("%dG", v.RAM))

	// Add disk arguments
	args = append(args, vm.GetDiskArgs()...)

	// Add network arguments
	args = append(args, vm.GetNetworkArgs()...)

	// Add audio device (disabled)
	args = append(args, "-audiodev", fmt.Sprintf("%s,id=%s", AudioDeviceType, AudioDeviceID))

	// Handle display mode
	if v.Graphical {
		graphicalArgs := vm.GetGraphicalArgs()
		args = append(args, graphicalArgs...)
		// If graphical mode is implemented via -nographic (e.g., for s390x),
		// we must disable the default monitor to avoid stdio conflicts.
		if slices.Contains(graphicalArgs, "-nographic") {
			args = append(args, "-monitor", "none")
		}
	} else {
		nonGraphicalDisplayArgs := vm.GetNonGraphicalDisplayArgs()
		args = append(args, nonGraphicalDisplayArgs...)
		if !v.NoSnapshot {
			args = append(args, SnapshotArgument)
		}
	}

	// Handle monitor configuration
	if v.MonitorPort > 0 {
		args = append(args, "-monitor", fmt.Sprintf("%s:%s:%d,server,nowait", MonitorProtocol, LocalhostAddress, v.MonitorPort))
	} else if !v.Graphical {
		// For console modes, disable the interactive monitor on stdio by default
		// unless it's already handled (e.g. for s390x).
		if !slices.Contains(args, "-monitor") {
			args = append(args, "-monitor", "none")
		}
	}
	// For graphical modes, the default monitor is usually in the GUI window, which is fine.

	return args
}

// run is a helper to execute the VM, containing logic common to all architectures.
func (v *BaseVM) run(vm VM) error {
	args := v.buildArgs(vm)
	return RunVM(vm.QEMUBinary(), args, v.Confirm)
}

// RunVM executes the VM with the given binary and arguments
func RunVM(binary string, args []string, confirm bool) error {
	logger.Info("🚀 Starting QEMU with the following command:")
	logger.Info("Command", "binary", binary, "args", strings.Join(args, " "))

	if confirm {
		fmt.Print("Press Enter to continue...")
		var input string
		fmt.Scanln(&input)
	}

	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			logger.Error("QEMU exited with error", "status", exitError.ExitCode())
			return fmt.Errorf("QEMU exited with status %d", exitError.ExitCode())
		}
		logger.Error("Failed to start QEMU", "error", err)
		return fmt.Errorf("failed to start QEMU: %w", err)
	}

	return nil
}

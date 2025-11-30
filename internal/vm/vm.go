package vm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ilmanzo/q2boot/internal/config"
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

	// BuildArgs builds the complete QEMU command line arguments
	BuildArgs() []string

	// Run executes the virtual machine
	Run() error

	// Configure sets up the VM with the provided configuration
	Configure(cfg *config.VMConfig)

	// SetDiskPath sets the disk image path
	SetDiskPath(path string)
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
		CPU:         2,
		RAM:         2,
		SSHPort:     2222,
		MonitorPort: 0,
		LogFile:     "q2boot.log",
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

// run is a helper to execute the VM, containing logic common to all architectures.
// It relies on the passed-in VM interface to get architecture-specific details.
func (v *BaseVM) run(vm VM) error {
	args := vm.BuildArgs()
	return RunVM(vm.QEMUBinary(), args, v.Confirm)
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

// buildArgs is a helper to build the QEMU command line arguments, containing
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
	args = append(args, "-audiodev", "none,id=snd0")

	// Handle display mode
	if v.Graphical {
		args = append(args, vm.GetGraphicalArgs()...)
	} else {
		args = append(args, "-display", "curses")
		if !v.NoSnapshot {
			args = append(args, "-snapshot")
		}
		args = append(args, "-serial", "stdio")
	}

	// Handle monitor configuration
	if v.MonitorPort > 0 {
		args = append(args, "-monitor", fmt.Sprintf("telnet:127.0.0.1:%d,server,nowait", v.MonitorPort))
	} else if !v.Graphical {
		// For console modes, disable the interactive monitor on stdio by default
		// This prevents conflicts with the serial console
		args = append(args, "-monitor", "none")
	}
	// For graphical modes, the default monitor is usually in the GUI window, which is fine.

	return args
}

// RunVM executes the VM with the given binary and arguments
func RunVM(binary string, args []string, confirm bool) error {
	fmt.Printf("🚀 Starting QEMU with the following command:\n")
	fmt.Printf("%s %s\n", binary, strings.Join(args, " "))

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
			return fmt.Errorf("QEMU exited with status %d", exitError.ExitCode())
		}
		return fmt.Errorf("failed to start QEMU: %w", err)
	}

	return nil
}

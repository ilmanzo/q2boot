package vm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ilmanzo/qboot/internal/config"
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
	DiskPath   string
	CPU        int
	RAM        int
	Graphical  bool
	NoSnapshot bool
	Confirm    bool
	SSHPort    uint16
	LogFile    string
}

// NewBaseVM creates a new BaseVM with default settings
func NewBaseVM() *BaseVM {
	return &BaseVM{
		CPU:        2,
		RAM:        2,
		SSHPort:    2222,
		LogFile:    "qboot.log",
		Graphical:  false,
		NoSnapshot: false,
		Confirm:    false,
	}
}

// Configure sets up the VM with the provided configuration
func (v *BaseVM) Configure(cfg *config.VMConfig) {
	v.CPU = cfg.CPU
	v.RAM = cfg.RAMGb
	v.SSHPort = cfg.SSHPort
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

// ValidateDiskPath validates the disk path and returns an error if it's invalid
func ValidateDiskPath(diskPath string) error {
	if diskPath == "" {
		return fmt.Errorf("disk path cannot be empty. Use -d or --disk")
	}

	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return fmt.Errorf("disk image not found at '%s'", diskPath)
	}

	return nil
}

// ValidateVMConfig validates the VM configuration parameters
func ValidateVMConfig(vm *BaseVM) error {
	if vm.CPU < 1 || vm.CPU > 32 {
		return fmt.Errorf("CPU count must be between 1 and 32, got %d", vm.CPU)
	}

	if vm.RAM < 1 || vm.RAM > 128 {
		return fmt.Errorf("RAM must be between 1 and 128 GB, got %d", vm.RAM)
	}

	if vm.SSHPort < 1024 || vm.SSHPort > 65535 {
		return fmt.Errorf("SSH port must be between 1024 and 65535, got %d", vm.SSHPort)
	}

	return nil
}

// BuildCommonArgs builds the common QEMU arguments for all architectures
func (v *BaseVM) BuildCommonArgs(archArgs, diskArgs, netArgs []string) []string {
	var args []string

	// Add architecture-specific arguments
	args = append(args, archArgs...)

	// Add common arguments
	args = append(args, "-smp", fmt.Sprintf("%d", v.CPU))
	args = append(args, "-m", fmt.Sprintf("%dG", v.RAM))

	// Add disk arguments
	args = append(args, diskArgs...)

	// Add network arguments
	args = append(args, netArgs...)

	// Add audio device (disabled)
	args = append(args, "-audiodev", "none,id=snd0")

	// Handle display mode
	if v.Graphical {
		// Graphical args will be added by specific implementations
	} else {
		args = append(args, "-nographic")
		if !v.NoSnapshot {
			args = append(args, "-snapshot")
		}
		args = append(args, "-serial", "stdio", "-monitor", "none")
	}

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

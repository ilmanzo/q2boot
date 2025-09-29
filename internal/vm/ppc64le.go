package vm

import (
	"fmt"

	"github.com/ilmanzo/qboot/internal/config"
)

// PPC64LEVM implements VM for ppc64le architecture
type PPC64LEVM struct {
	*BaseVM
}

// NewPPC64LEVM creates a new PPC64LEVM instance
func NewPPC64LEVM() *PPC64LEVM {
	return &PPC64LEVM{
		BaseVM: NewBaseVM(),
	}
}

// QEMUBinary returns the QEMU binary name for ppc64le
func (vm *PPC64LEVM) QEMUBinary() string {
	return "qemu-system-ppc64"
}

// GetArchArgs returns architecture-specific arguments for ppc64le
func (vm *PPC64LEVM) GetArchArgs() []string {
	return []string{"-M", "pseries", "-cpu", "POWER9"}
}

// GetDiskArgs returns disk-specific arguments for ppc64le
func (vm *PPC64LEVM) GetDiskArgs() []string {
	return []string{
		"-drive",
		fmt.Sprintf("file=%s,if=virtio,cache=writeback,aio=native,discard=unmap,cache.direct=on", vm.DiskPath),
	}
}

// GetNetworkArgs returns network-specific arguments for ppc64le
func (vm *PPC64LEVM) GetNetworkArgs() []string {
	return []string{
		"-netdev",
		fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:22", vm.SSHPort),
		"-device",
		"virtio-net-pci,netdev=net0",
	}
}

// GetGraphicalArgs returns arguments for graphical mode on ppc64le
func (vm *PPC64LEVM) GetGraphicalArgs() []string {
	return []string{"-device", "virtio-vga", "-display", "sdl"}
}

// BuildArgs builds the complete argument list for ppc64le
func (vm *PPC64LEVM) BuildArgs() []string {
	var args []string

	// Add architecture-specific arguments
	args = append(args, vm.GetArchArgs()...)

	// Add common arguments
	args = append(args, "-smp", fmt.Sprintf("%d", vm.CPU))
	args = append(args, "-m", fmt.Sprintf("%dG", vm.RAM))

	// Add disk arguments
	args = append(args, vm.GetDiskArgs()...)

	// Add network arguments
	args = append(args, vm.GetNetworkArgs()...)

	// Add audio device (disabled)
	args = append(args, "-audiodev", "none,id=snd0")

	// Handle display mode
	if vm.Graphical {
		args = append(args, vm.GetGraphicalArgs()...)
	} else {
		args = append(args, "-nographic")
		if !vm.NoSnapshot {
			args = append(args, "-snapshot")
		}
		args = append(args, "-serial", "stdio", "-monitor", "none")
	}

	return args
}

// Run executes the ppc64le VM
func (vm *PPC64LEVM) Run() error {
	// Validate QEMU binary is available
	if err := ValidateQEMUBinary(vm.QEMUBinary()); err != nil {
		return err
	}

	if err := ValidateDiskPath(vm.DiskPath); err != nil {
		return err
	}

	if err := ValidateVMConfig(vm.BaseVM); err != nil {
		return err
	}

	args := vm.BuildArgs()
	return RunVM(vm.QEMUBinary(), args, vm.Confirm)
}

// Configure sets up the VM with the provided configuration
func (vm *PPC64LEVM) Configure(cfg *config.VMConfig) {
	vm.BaseVM.Configure(cfg)
}

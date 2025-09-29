package vm

import (
	"fmt"

	"github.com/ilmanzo/qboot/internal/config"
)

// AARCH64VM implements VM for aarch64 architecture
type AARCH64VM struct {
	*BaseVM
}

// NewAARCH64VM creates a new AARCH64VM instance
func NewAARCH64VM() *AARCH64VM {
	return &AARCH64VM{
		BaseVM: NewBaseVM(),
	}
}

// QEMUBinary returns the QEMU binary name for aarch64
func (vm *AARCH64VM) QEMUBinary() string {
	return "qemu-system-aarch64"
}

// GetArchArgs returns architecture-specific arguments for aarch64
func (vm *AARCH64VM) GetArchArgs() []string {
	// This requires a UEFI firmware file. A common path is provided.
	// Users might need to install it via their package manager
	// (e.g., qemu-efi-aarch64 on Debian/Ubuntu).
	return []string{
		"-machine", "virt",
		"-cpu", "max",
		"-bios", "/usr/share/qemu/aavmf-aarch64-code.bin",
	}
}

// GetDiskArgs returns disk-specific arguments for aarch64
func (vm *AARCH64VM) GetDiskArgs() []string {
	return []string{
		"-drive",
		fmt.Sprintf("file=%s,if=virtio,cache=writeback,aio=native,discard=unmap,cache.direct=on", vm.DiskPath),
	}
}

// GetNetworkArgs returns network-specific arguments for aarch64
func (vm *AARCH64VM) GetNetworkArgs() []string {
	return []string{
		"-netdev",
		fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:22", vm.SSHPort),
		"-device",
		"virtio-net-pci,netdev=net0",
	}
}

// GetGraphicalArgs returns arguments for graphical mode on aarch64
func (vm *AARCH64VM) GetGraphicalArgs() []string {
	return []string{"-device", "virtio-gpu-pci", "-display", "sdl"}
}

// BuildArgs builds the complete argument list for aarch64
func (vm *AARCH64VM) BuildArgs() []string {
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

// Run executes the aarch64 VM
func (vm *AARCH64VM) Run() error {
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
func (vm *AARCH64VM) Configure(cfg *config.VMConfig) {
	vm.BaseVM.Configure(cfg)
}

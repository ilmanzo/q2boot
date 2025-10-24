package vm

import (
	"fmt"

	"github.com/ilmanzo/q2boot/internal/config"
)

// X86_64VM implements VM for x86_64 architecture
type X86_64VM struct {
	*BaseVM
}

// NewX86_64VM creates a new X86_64VM instance
func NewX86_64VM() *X86_64VM {
	return &X86_64VM{
		BaseVM: NewBaseVM(),
	}
}

// QEMUBinary returns the QEMU binary name for x86_64
func (vm *X86_64VM) QEMUBinary() string {
	return "qemu-system-x86_64"
}

// GetArchArgs returns architecture-specific arguments for x86_64
func (vm *X86_64VM) GetArchArgs() []string {
	return []string{"-M", "q35", "-enable-kvm", "-cpu", "host"}
}

// GetDiskArgs returns disk-specific arguments for x86_64
func (vm *X86_64VM) GetDiskArgs() []string {
	return []string{
		"-drive",
		fmt.Sprintf("file=%s,if=virtio,cache=writeback,aio=native,discard=unmap,cache.direct=on", vm.DiskPath),
	}
}

// GetNetworkArgs returns network-specific arguments for x86_64
func (vm *X86_64VM) GetNetworkArgs() []string {
	return []string{
		"-netdev",
		fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:22", vm.SSHPort),
		"-device",
		"virtio-net-pci,netdev=net0",
	}
}

// GetGraphicalArgs returns arguments for graphical mode on x86_64
func (vm *X86_64VM) GetGraphicalArgs() []string {
	return []string{"-device", "virtio-vga-gl", "-display", "sdl,gl=on"}
}

// BuildArgs builds the complete argument list for x86_64
func (vm *X86_64VM) BuildArgs() []string {
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

// Run executes the x86_64 VM
func (vm *X86_64VM) Run() error {
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
func (vm *X86_64VM) Configure(cfg *config.VMConfig) {
	vm.BaseVM.Configure(cfg)
}

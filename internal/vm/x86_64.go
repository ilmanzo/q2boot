package vm

import (
	"fmt"
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
		fmt.Sprintf("file=%s,if=none,id=disk0,cache=none,aio=native,discard=unmap", vm.DiskPath),
		"-device",
		fmt.Sprintf("virtio-blk-pci,drive=disk0,num-queues=%d", vm.CPU),
	}
}

// GetNetworkArgs returns network-specific arguments for x86_64
func (vm *X86_64VM) GetNetworkArgs() []string {
	return []string{
		"-netdev",
		fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:22", vm.SSHPort),
		"-device",
		"virtio-net-pci,netdev=net0,mq=on",
	}
}

// GetGraphicalArgs returns arguments for graphical mode on x86_64
func (vm *X86_64VM) GetGraphicalArgs() []string {
	return []string{"-device", "virtio-vga-gl", "-display", "sdl,gl=on"}
}

// GetNonGraphicalDisplayArgs returns display arguments for non-graphical mode on x86_64
// x86_64 uses the default curses display for headless mode
func (vm *X86_64VM) GetNonGraphicalDisplayArgs() []string {
	return []string{"-display", "curses"}
}

// BuildArgs builds the complete argument list for x86_64
func (vm *X86_64VM) BuildArgs() []string {
	return vm.buildArgs(vm)
}

// Run executes the x86_64 VM
func (vm *X86_64VM) Run() error {
	return vm.run(vm)
}

package vm

import (
	"fmt"
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
		fmt.Sprintf("file=%s,if=none,id=disk0,cache=none,aio=native,discard=unmap", vm.DiskPath),
		"-device",
		fmt.Sprintf("virtio-blk-pci,drive=disk0,num-queues=%d", vm.CPU),
	}
}

// GetNetworkArgs returns network-specific arguments for aarch64
func (vm *AARCH64VM) GetNetworkArgs() []string {
	return []string{
		"-netdev",
		fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:22", vm.SSHPort),
		"-device",
		"virtio-net-pci,netdev=net0,mq=on",
	}
}

// GetGraphicalArgs returns arguments for graphical mode on aarch64
func (vm *AARCH64VM) GetGraphicalArgs() []string {
	return []string{"-device", "virtio-gpu-pci", "-display", "gtk"}
}

// GetNonGraphicalDisplayArgs returns display arguments for non-graphical mode on aarch64
// aarch64 uses the default curses display for headless mode
func (vm *AARCH64VM) GetNonGraphicalDisplayArgs() []string {
	return []string{"-display", "curses"}
}

// BuildArgs builds the complete argument list for aarch64
func (vm *AARCH64VM) BuildArgs() []string {
	return vm.buildArgs(vm)
}

// Run executes the VM
func (vm *AARCH64VM) Run() error {
	return vm.RunVM(vm)
}

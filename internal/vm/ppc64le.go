package vm

import (
	"fmt"
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
	return []string{"-M", "pseries", "-cpu", "power10"}
}

// GetDiskArgs returns disk-specific arguments for ppc64le
func (vm *PPC64LEVM) GetDiskArgs() []string {
	return []string{
		"-drive",
		fmt.Sprintf("file=%s,id=disk0,if=none,cache=none,aio=native,discard=unmap", vm.DiskPath),
		"-device",
		fmt.Sprintf("virtio-blk-pci,drive=disk0,id=dr0,bootindex=1,num-queues=%d", vm.CPU),
	}
}

// GetNetworkArgs returns network-specific arguments for ppc64le
func (vm *PPC64LEVM) GetNetworkArgs() []string {
	return []string{
		"-netdev",
		fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:22", vm.SSHPort),
		"-device",
		"virtio-net-pci,netdev=net0,mq=on",
	}
}

// GetGraphicalArgs returns arguments for graphical mode on ppc64le
func (vm *PPC64LEVM) GetGraphicalArgs() []string {
	// ppc64le does not support virtio-vga for graphical output.
	// Instead, provide an interactive serial console, which is the
	// expected behavior for this architecture.
	return []string{"-nographic", "-serial", "stdio"}
}

// GetNonGraphicalDisplayArgs returns display arguments for non-graphical mode on ppc64le
// ppc64le works better with nographic mode and serial stdio instead of curses
func (vm *PPC64LEVM) GetNonGraphicalDisplayArgs() []string {
	return []string{"-nographic"}
}

// Validate checks the VM configuration and satisfies the VM interface.
func (vm *PPC64LEVM) Validate() error {
	return vm.BaseVM.Validate(vm)
}

// Run executes the VM and satisfies the VM interface.
func (vm *PPC64LEVM) Run() error {
	return vm.run(vm)
}

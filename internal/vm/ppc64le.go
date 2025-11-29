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
	return vm.buildArgs(vm)
}

// Run executes the ppc64le VM
func (vm *PPC64LEVM) Run() error {
	return vm.run(vm)
}

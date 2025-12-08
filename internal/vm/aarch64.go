package vm

import (
	"fmt"
	"os"
)

// AARCH64VM implements VM for aarch64 architecture
type AARCH64VM struct {
	*BaseVM
}

// aarch64 UEFI firmware paths
var aavmfCodePaths = []string{
	"/usr/share/qemu/aavmf-aarch64-code.bin", // SUSE
	"/usr/share/AAVMF/AAVMF_CODE.fd",         // Debian/Ubuntu
}

// NewAARCH64VM creates a new AARCH64VM instance
func NewAARCH64VM() *AARCH64VM {
	vm := &AARCH64VM{
		BaseVM: NewBaseVM(),
	}

	// Set default firmware path if not already set
	if vm.FirmwarePath == "" {
		for _, path := range aavmfCodePaths {
			if _, err := os.Stat(path); err == nil {
				vm.FirmwarePath = path
				break
			}
		}
	}
	return vm
}

// QEMUBinary returns the QEMU binary name for aarch64
func (vm *AARCH64VM) QEMUBinary() string {
	return "qemu-system-aarch64"
}

// GetArchArgs returns architecture-specific arguments for aarch64
func (vm *AARCH64VM) GetArchArgs() []string {
	args := []string{"-M", "virt", "-cpu", "max"}

	if vm.FirmwarePath != "" {
		// The variable store needs to be the same size as the code store.
		firmwareInfo, err := os.Stat(vm.FirmwarePath)
		if err != nil {
			// If we can't stat the firmware, we can't proceed with pflash.
			// This should be caught by other validations, but we'll be safe.
			return args
		}

		// Create a temporary file for UEFI variables.
		varsFile, err := os.CreateTemp("", "q2boot-aavmf-vars-*.fd")
		if err == nil {
			// Resize the empty file to match the firmware size.
			varsFile.Truncate(firmwareInfo.Size())
			varsFile.Close() // Close the file handle.

			// QEMU needs two pflash devices for UEFI: one for code (readonly) and one for vars.
			args = append(args,
				"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", vm.FirmwarePath),
				"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", varsFile.Name()),
			)
		}
	}

	return args
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
	return []string{"-device", "virtio-vga-gl", "-display", "gtk,gl=on"}
}

// GetNonGraphicalDisplayArgs returns display arguments for non-graphical mode on aarch64
// For non-graphical mode, we disable the display and redirect the serial console.
func (vm *AARCH64VM) GetNonGraphicalDisplayArgs() []string {
	if vm.LogFile != "" {
		return []string{"-nographic"}
	}
	return []string{
		"-nographic",
		"-serial",
		"mon:stdio",
	}
}

// Validate checks the VM configuration and satisfies the VM interface.
func (vm *AARCH64VM) Validate() error {
	return vm.BaseVM.Validate(vm)
}

// Run executes the VM and satisfies the VM interface.
func (vm *AARCH64VM) Run() error {
	return vm.run(vm)
}

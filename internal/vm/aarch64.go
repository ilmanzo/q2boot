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
	vm := &AARCH64VM{BaseVM: NewBaseVM()}
	vm.self = vm

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

// GetArchArgs returns architecture-specific arguments for aarch64.
// UEFI requires two pflash devices: read-only code + writable vars (NVRAM).
// The vars file is persisted next to the disk image so boot variables survive
// across runs (matching real hardware behaviour) and we don't leak temp files.
func (vm *AARCH64VM) GetArchArgs() []string {
	args := []string{"-M", "virt", "-cpu", "max"}

	if vm.FirmwarePath == "" || vm.DiskPath == "" {
		// Without a disk we have nowhere to anchor the vars file; skip pflash.
		return args
	}

	firmwareInfo, err := os.Stat(vm.FirmwarePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot stat aarch64 firmware %s: %v (skipping UEFI pflash setup)\n", vm.FirmwarePath, err)
		return args
	}

	varsPath := vm.DiskPath + ".vars"
	if err := ensureVarsFile(varsPath, firmwareInfo.Size()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot prepare UEFI vars file %s: %v (skipping pflash setup)\n", varsPath, err)
		return args
	}

	return append(args,
		"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", vm.FirmwarePath),
		"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", varsPath),
	)
}

// ensureVarsFile creates path as a zero-filled file of exactly size bytes if
// it does not already exist. Existing files are left untouched.
func ensureVarsFile(path string, size int64) error {
	if info, err := os.Stat(path); err == nil {
		if info.Size() != size {
			return fmt.Errorf("existing vars file %s has size %d, expected %d (delete it to reinitialise)", path, info.Size(), size)
		}
		return nil
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Truncate(size)
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

package vm

import (
	"fmt"
	"os"
)

// ArchSpec captures every per-architecture knob q2boot needs to compose a
// QEMU command line. Adding a new arch is one entry in the archSpecs table.
type ArchSpec struct {
	Arch        string
	Binary      string   // qemu-system-* binary name
	MachineArgs []string // -M ... -cpu ... and any per-arch acceleration flags

	DiskBus       string // virtio-blk-pci or virtio-blk-ccw
	DiskID        string // drive id (also used by the device line)
	DiskDevSuffix string // extra device options between drive= and num-queues= (e.g. ",id=dr0,bootindex=1")

	NetBus string // virtio-net-pci or virtio-net-ccw
	NetID  string // netdev id

	// Graphical lists args used in --graphical mode on archs that have real
	// graphical output. ppc64le and s390x have no graphical mode; for them
	// SerialGraphical=true makes --graphical fall back to an interactive
	// serial console.
	Graphical       []string
	SerialGraphical bool

	// UEFIFirmware lists candidate paths for the read-only UEFI code blob
	// (aarch64 only). The first that exists wins; if none exist the firmware
	// is silently omitted (the VM may still boot via direct kernel/initrd).
	// Per-disk NVRAM vars file lives at <diskpath>.vars and is created on
	// first boot, then reused.
	UEFIFirmware []string

	// InstallPkgs names the per-distro package providing this arch's QEMU
	// binary, used by the version/check output. Order: Ubuntu/Debian, SUSE
	// suffix (for qemu-<x>), Arch suffix (for qemu-system-<x>).
	InstallPkgs InstallPkgs
}

type InstallPkgs struct{ Ubuntu, SUSE, Arch string }

var archSpecs = []ArchSpec{
	{
		Arch:        "x86_64",
		Binary:      "qemu-system-x86_64",
		MachineArgs: []string{"-M", "q35", "-enable-kvm", "-cpu", "host"},
		DiskBus:     "virtio-blk-pci", DiskID: "disk0",
		NetBus: "virtio-net-pci", NetID: "net0",
		Graphical:   []string{"-device", "virtio-vga-gl", "-display", "gtk,gl=on"},
		InstallPkgs: InstallPkgs{"qemu-system-x86", "x86", "x86"},
	},
	{
		Arch:        "aarch64",
		Binary:      "qemu-system-aarch64",
		MachineArgs: []string{"-M", "virt", "-cpu", "max"},
		DiskBus:     "virtio-blk-pci", DiskID: "disk0",
		NetBus: "virtio-net-pci", NetID: "net0",
		Graphical: []string{"-device", "virtio-vga-gl", "-display", "gtk,gl=on"},
		UEFIFirmware: []string{
			"/usr/share/qemu/aavmf-aarch64-code.bin", // SUSE
			"/usr/share/AAVMF/AAVMF_CODE.fd",         // Debian/Ubuntu
		},
		InstallPkgs: InstallPkgs{"qemu-system-arm", "arm", "aarch64"},
	},
	{
		Arch:        "ppc64le",
		Binary:      "qemu-system-ppc64",
		MachineArgs: []string{"-M", "pseries", "-cpu", "power10"},
		DiskBus:     "virtio-blk-pci", DiskID: "disk0", DiskDevSuffix: ",id=dr0,bootindex=1",
		NetBus: "virtio-net-pci", NetID: "net0",
		SerialGraphical: true,
		InstallPkgs:     InstallPkgs{"qemu-system-ppc", "ppc", "ppc64"},
	},
	{
		Arch:        "s390x",
		Binary:      "qemu-system-s390x",
		MachineArgs: []string{"-machine", "s390-ccw-virtio", "-cpu", "max"},
		DiskBus:     "virtio-blk-ccw", DiskID: "disk1", DiskDevSuffix: ",id=dr1,bootindex=1",
		NetBus: "virtio-net-ccw", NetID: "net1",
		SerialGraphical: true,
		InstallPkgs:     InstallPkgs{"qemu-system-s390x", "s390x", "s390x"},
	},
}

func specFor(arch string) (ArchSpec, bool) {
	for _, s := range archSpecs {
		if s.Arch == arch {
			return s, true
		}
	}
	return ArchSpec{}, false
}

// archVM is the single VM implementation; behaviour is driven entirely by spec.
type archVM struct {
	*BaseVM
	spec ArchSpec
}

func newArchVM(spec ArchSpec) *archVM {
	v := &archVM{BaseVM: NewBaseVM(), spec: spec}
	v.self = v
	for _, p := range spec.UEFIFirmware {
		if _, err := os.Stat(p); err == nil {
			v.FirmwarePath = p
			break
		}
	}
	return v
}

func (v *archVM) QEMUBinary() string { return v.spec.Binary }

func (v *archVM) GetArchArgs() []string {
	args := append([]string(nil), v.spec.MachineArgs...)
	return append(args, v.uefiArgs()...)
}

// uefiArgs returns the two pflash drive args for archs with UEFI firmware,
// reusing <diskpath>.vars across boots so NVRAM behaves like real hardware
// and we never leak temp files (see ensureVarsFile).
func (v *archVM) uefiArgs() []string {
	if v.FirmwarePath == "" || v.DiskPath == "" {
		return nil
	}
	info, err := os.Stat(v.FirmwarePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot stat firmware %s: %v (skipping UEFI pflash)\n", v.FirmwarePath, err)
		return nil
	}
	varsPath := v.DiskPath + ".vars"
	if err := ensureVarsFile(varsPath, info.Size()); err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot prepare UEFI vars file %s: %v (skipping UEFI pflash)\n", varsPath, err)
		return nil
	}
	return []string{
		"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", v.FirmwarePath),
		"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", varsPath),
	}
}

func (v *archVM) GetDiskArgs() []string {
	return []string{
		"-drive", fmt.Sprintf("file=%s,if=none,id=%s,cache=none,aio=native,discard=unmap", v.DiskPath, v.spec.DiskID),
		"-device", fmt.Sprintf("%s,drive=%s%s,num-queues=%d", v.spec.DiskBus, v.spec.DiskID, v.spec.DiskDevSuffix, v.CPU),
	}
}

func (v *archVM) GetNetworkArgs() []string {
	return []string{
		"-netdev", fmt.Sprintf("user,id=%s,hostfwd=tcp::%d-:22", v.spec.NetID, v.SSHPort),
		"-device", fmt.Sprintf("%s,netdev=%s,mq=on", v.spec.NetBus, v.spec.NetID),
	}
}

func (v *archVM) GetGraphicalArgs() []string {
	if !v.spec.SerialGraphical {
		return v.spec.Graphical
	}
	if v.LogFile != "" {
		return []string{"-nographic"}
	}
	return []string{"-nographic", "-serial", "stdio"}
}

func (v *archVM) GetNonGraphicalDisplayArgs() []string {
	if v.spec.SerialGraphical || v.LogFile != "" {
		return []string{"-nographic"}
	}
	return []string{"-nographic", "-serial", "mon:stdio"}
}

// ensureVarsFile creates path as a zero-filled file of exactly size bytes if
// it does not exist; existing files are left untouched (an unexpected size
// is surfaced as an error so we don't silently feed QEMU a mismatched blob).
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

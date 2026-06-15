package vm

import "fmt"

// CreateVM creates a VM instance based on the specified architecture.
// Declared as a var so tests can substitute it (see internal/vm/mock_vm.go).
var CreateVM = func(arch string) (VM, error) {
	spec, ok := specFor(arch)
	if !ok {
		return nil, fmt.Errorf("unsupported architecture: %s", arch)
	}
	return newArchVM(spec), nil
}

// SupportedArchitectures returns the list of architectures q2boot knows about.
func SupportedArchitectures() []string {
	out := make([]string, len(archSpecs))
	for i, s := range archSpecs {
		out[i] = s.Arch
	}
	return out
}

func IsArchSupported(arch string) bool {
	_, ok := specFor(arch)
	return ok
}

func GetQEMUBinaryForArch(arch string) (string, error) {
	spec, ok := specFor(arch)
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
	return spec.Binary, nil
}

// CheckAvailableQEMUBinaries returns arch → whether qemu-system-<arch> is on PATH.
func CheckAvailableQEMUBinaries() map[string]bool {
	out := make(map[string]bool, len(archSpecs))
	for _, s := range archSpecs {
		out[s.Arch] = ValidateQEMUBinary(s.Binary) == nil
	}
	return out
}

func GetMissingQEMUBinaries() []string {
	var missing []string
	for arch, ok := range CheckAvailableQEMUBinaries() {
		if !ok {
			missing = append(missing, arch)
		}
	}
	return missing
}

// GetInstallationInstructions returns per-distro install hints for an arch's
// QEMU binary. Lookup is by binary name (not arch) to keep the public API
// stable for the version/check commands.
func GetInstallationInstructions(binary string) string {
	pkgs := InstallPkgs{"qemu-system", "unknown", "unknown"}
	for _, s := range archSpecs {
		if s.Binary == binary {
			pkgs = s.InstallPkgs
			break
		}
	}
	return fmt.Sprintf("Please install the appropriate QEMU package for your system:\n"+
		"  - Ubuntu/Debian: sudo apt install %s\n"+
		"  - RHEL/CentOS/Fedora: sudo dnf install qemu-system or sudo yum install qemu-system\n"+
		"  - SUSE/openSUSE: sudo zypper install qemu-%s\n"+
		"  - Arch Linux: sudo pacman -S qemu-system-%s\n"+
		"  - macOS: brew install qemu", pkgs.Ubuntu, pkgs.SUSE, pkgs.Arch)
}

package vm

import "fmt"

// CreateVM is a var so tests can swap it (see mock_vm.go).
var CreateVM = func(arch string) (VM, error) {
	spec, ok := specFor(arch)
	if !ok {
		return nil, fmt.Errorf("unsupported architecture: %s", arch)
	}
	return newArchVM(spec), nil
}

func SupportedArchitectures() []string {
	out := make([]string, len(archSpecs))
	for i, s := range archSpecs {
		out[i] = s.Arch
	}
	return out
}

func GetQEMUBinaryForArch(arch string) (string, error) {
	spec, ok := specFor(arch)
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
	return spec.Binary, nil
}

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

// GetInstallationInstructions is keyed by binary name (not arch) so the
// version/check commands can call it with the QEMU binary they probed.
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

package vm

import (
	"fmt"
	"slices"
	"strings"
)

// CreateVM creates a VM instance based on the specified architecture
func CreateVM(arch string) (VM, error) {
	switch arch {
	case "x86_64":
		return NewX86_64VM(), nil
	case "aarch64":
		return NewAARCH64VM(), nil
	case "ppc64le":
		return NewPPC64LEVM(), nil
	case "s390x":
		return NewS390XVM(), nil
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", arch)
	}
}

// SupportedArchitectures returns a list of supported architectures
func SupportedArchitectures() []string {
	return []string{"x86_64", "aarch64", "ppc64le", "s390x"}
}

// IsArchSupported checks if the given architecture is supported
func IsArchSupported(arch string) bool {
	return slices.Contains(SupportedArchitectures(), arch)
}

// GetQEMUBinaryForArch returns the QEMU binary name for the given architecture
func GetQEMUBinaryForArch(arch string) (string, error) {
	vm, err := CreateVM(arch)
	if err != nil {
		return "", err
	}
	return vm.QEMUBinary(), nil
}

// CheckAvailableQEMUBinaries checks which QEMU binaries are available on the system
func CheckAvailableQEMUBinaries() map[string]bool {
	availability := make(map[string]bool)

	for _, arch := range SupportedArchitectures() {
		binary, err := GetQEMUBinaryForArch(arch)
		if err != nil {
			availability[arch] = false
			continue
		}

		err = ValidateQEMUBinary(binary)
		availability[arch] = (err == nil)
	}

	return availability
}

// GetMissingQEMUBinaries returns a list of architectures that are missing their QEMU binaries
func GetMissingQEMUBinaries() []string {
	var missing []string
	availability := CheckAvailableQEMUBinaries()

	for arch, available := range availability {
		if !available {
			missing = append(missing, arch)
		}
	}

	return missing
}

// ValidateArchitectureSupport checks if the given architecture is supported
func ValidateArchitectureSupport(arch string) error {
	if !IsArchSupported(arch) {
		return fmt.Errorf("unsupported architecture: %s. Supported architectures: %s",
			arch, strings.Join(SupportedArchitectures(), ", "))
	}
	return nil
}

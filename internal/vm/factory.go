package vm

import (
	"fmt"
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
	for _, supported := range SupportedArchitectures() {
		if arch == supported {
			return true
		}
	}
	return false
}

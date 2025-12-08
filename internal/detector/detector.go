// Package detector provides automatic architecture detection from disk images
package detector

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// SupportedArchitectures lists all architectures that can be detected
var SupportedArchitectures = []string{"x86_64", "aarch64", "ppc64le", "s390x"}

// DetectArchitecture attempts to detect the architecture from a disk image.
// It tries multiple detection methods in order of reliability.
// Returns the detected architecture or an error if detection fails.
var DetectArchitecture = func(diskPath string) (string, error) {
	if diskPath == "" {
		return "", fmt.Errorf("disk path is empty")
	}

	// Method 1: Use virt-cat (most reliable when available)
	if arch, err := detectByVirtCat(diskPath); err == nil {
		return arch, nil
	}
	// If virt-cat fails, we log it but don't error out, allowing fallback.

	// Method 2: Fallback to filename inspection
	if arch, err := detectByFilename(diskPath); err == nil {
		return arch, nil
	}

	return "", fmt.Errorf("could not detect architecture from disk image '%s'. Please specify it explicitly with --arch flag", diskPath)
}

func detectByFilename(diskPath string) (string, error) {
	lowerCasePath := strings.ToLower(diskPath)

	// Iterate over supported architectures and check if they are in the filename
	for _, arch := range SupportedArchitectures {
		// Use word boundaries or common separators to avoid partial matches (e.g., "s390" in a version number)
		if strings.Contains(lowerCasePath, "@"+arch) || strings.Contains(lowerCasePath, "-"+arch) || strings.Contains(lowerCasePath, "_"+arch) {
			return arch, nil
		}
	}

	return "", fmt.Errorf("could not deduce architecture from filename for '%s'", diskPath)
}

// detectByVirtCat uses virt-cat (libguestfs) to extract a small file and run `file -` on it
// This is reliable because `file` reports the ELF architecture from the guest binary.
func detectByVirtCat(diskPath string) (string, error) {
	// Check virt-cat availability
	if _, err := exec.LookPath("virt-cat"); err != nil {
		return "", fmt.Errorf("virt-cat not found; please install guestfs-tools (package name may be 'guestfs-tools' or 'libguestfs-tools')")
	}

	// Inform the user this may take some time
	fmt.Fprintln(os.Stderr, "Detecting architecture using virt-cat (this may take a while)...")

	// Prepare commands: virt-cat <disk> /bin/sh  | file -
	cmdVirt := exec.Command("virt-cat", diskPath, "/bin/sh")
	stdoutPipe, err := cmdVirt.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to prepare virt-cat pipe: %w", err)
	}

	var buf bytes.Buffer
	cmdFile := exec.Command("file", "-")
	cmdFile.Stdin = stdoutPipe
	cmdFile.Stdout = &buf
	cmdFile.Stderr = &buf

	// Start file first so it's ready to read
	if err := cmdFile.Start(); err != nil {
		return "", fmt.Errorf("failed to start file command: %w", err)
	}

	if err := cmdVirt.Start(); err != nil {
		// ensure file process is cleaned up
		cmdFile.Wait()
		return "", fmt.Errorf("virt-cat failed to start: %w", err)
	}

	// Wait for virt-cat to finish
	if err := cmdVirt.Wait(); err != nil {
		cmdFile.Wait()
		return "", fmt.Errorf("virt-cat failed: %w; output: %s", err, strings.TrimSpace(buf.String()))
	}

	// Wait for file to finish and collect output
	if err := cmdFile.Wait(); err != nil {
		return "", fmt.Errorf("file command failed: %w; output: %s", err, strings.TrimSpace(buf.String()))
	}

	output := strings.ToLower(buf.String())

	// Check for specific ELF types first for accuracy
	if strings.Contains(output, "elf") && strings.Contains(output, "aarch64") {
		return "aarch64", nil
	}
	if strings.Contains(output, "elf") && (strings.Contains(output, "powerpc") || strings.Contains(output, "ppc64")) {
		return "ppc64le", nil
	}
	if strings.Contains(output, "elf") && (strings.Contains(output, "s390") || strings.Contains(output, "s/390")) {
		return "s390x", nil
	}
	if strings.Contains(output, "elf") && strings.Contains(output, "x86-64") {
		return "x86_64", nil
	}

	// If we don't find a clear ELF match, the output is ambiguous.
	return "", fmt.Errorf("virt-cat/file did not reveal a clear ELF architecture for '%s'; output: %s", diskPath, strings.TrimSpace(output))
}

// IsArchSupported checks if the given architecture is supported
func IsArchSupported(arch string) bool {
	return slices.Contains(SupportedArchitectures, arch)
}

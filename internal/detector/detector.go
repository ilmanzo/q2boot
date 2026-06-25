// Package detector provides automatic architecture detection from disk images
package detector

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// archFilenameRe requires a `-`, `_`, or `@` separator before the arch token
// to avoid matching inside version numbers or unrelated paths.
var archFilenameRe = regexp.MustCompile(`[-_@](x86_64|aarch64|ppc64le|s390x)\b`)

// Filename is tried first because it's instant and matches the conventional
// `-arch.qcow2` / `_arch.img` naming the SUSE/Ubuntu/Fedora image servers
// already use. virt-cat is the fallback for unhelpfully named images, but it
// shells out to libguestfs and can take 10+ seconds on a cold cache.
var DetectArchitecture = func(diskPath string) (string, error) {
	if diskPath == "" {
		return "", fmt.Errorf("disk path is empty")
	}
	if arch, err := detectByFilename(diskPath); err == nil {
		return arch, nil
	}
	if arch, err := detectByVirtCat(diskPath); err == nil {
		return arch, nil
	}
	return "", fmt.Errorf("could not detect architecture from disk image '%s'. Please specify it explicitly with --arch flag", diskPath)
}

func detectByFilename(diskPath string) (string, error) {
	if m := archFilenameRe.FindStringSubmatch(strings.ToLower(diskPath)); m != nil {
		return m[1], nil
	}
	return "", fmt.Errorf("could not deduce architecture from filename for '%s'", diskPath)
}

// HasLUKSPartition reports whether the disk image contains a LUKS-encrypted
// block device. A *closed* LUKS container is not a mountable filesystem, so
// `virt-filesystems --filesystems` (and --all/--extra/--long) never lists it;
// the only reliable signal is libguestfs' vfs-type, which returns
// "crypto_LUKS" for an encrypted device. We enumerate partitions + whole
// devices, then probe each. Returns false when guestfish is missing or the
// probe fails, so a missing guestfs-tools install never blocks a launch.
//
// ponytail: two appliance boots (~3s) on every non-graphical launch when
// guestfish is installed. Acceptable for the safety net; if launch latency
// matters, fold this into the same appliance boot as arch detection.
var HasLUKSPartition = func(diskPath string) bool {
	if _, err := exec.LookPath("guestfish"); err != nil {
		return false
	}
	out, err := exec.Command("guestfish", "--ro", "-a", diskPath,
		"run", ":", "list-partitions", ":", "list-devices").Output()
	if err != nil {
		return false
	}
	args := []string{"--ro", "-a", diskPath, "run"}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if dev := strings.TrimSpace(line); dev != "" {
			args = append(args, ":", "vfs-type", dev)
		}
	}
	if len(args) == 4 { // nothing to probe (empty/unreadable image)
		return false
	}
	types, err := exec.Command("guestfish", args...).Output()
	if err != nil {
		return false
	}
	return bytes.Contains(types, []byte("crypto_LUKS"))
}

// detectByVirtCat pipes `virt-cat <disk> /bin/sh | file -`; the ELF header
// `file` reports is the only reliable cross-distro arch signal.
func detectByVirtCat(diskPath string) (string, error) {
	if _, err := exec.LookPath("virt-cat"); err != nil {
		return "", fmt.Errorf("virt-cat not found; please install guestfs-tools (package name may be 'guestfs-tools' or 'libguestfs-tools')")
	}

	fmt.Fprintln(os.Stderr, "Detecting architecture using virt-cat (this may take a while)...")

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

	if err := cmdFile.Start(); err != nil {
		return "", fmt.Errorf("failed to start file command: %w", err)
	}
	if err := cmdVirt.Start(); err != nil {
		cmdFile.Wait()
		return "", fmt.Errorf("virt-cat failed to start: %w", err)
	}
	if err := cmdVirt.Wait(); err != nil {
		cmdFile.Wait()
		return "", fmt.Errorf("virt-cat failed: %w; output: %s", err, strings.TrimSpace(buf.String()))
	}
	if err := cmdFile.Wait(); err != nil {
		return "", fmt.Errorf("file command failed: %w; output: %s", err, strings.TrimSpace(buf.String()))
	}

	output := strings.ToLower(buf.String())
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
	return "", fmt.Errorf("virt-cat/file did not reveal a clear ELF architecture for '%s'; output: %s", diskPath, strings.TrimSpace(output))
}

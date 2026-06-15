package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// Aliased for test injection.
var (
	osReadFile = os.ReadFile
	osStat     = os.Stat
	osOpenFile = os.OpenFile
)

func NewCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Performs a pre-flight check of needed dependencies",
		Long: `The check command verifies that all required and recommended dependencies 
for running QEMU virtual machines are correctly installed and configured.

It checks for:
- KVM availability and permissions (Linux-only).
- QEMU system binaries for various architectures.
- Optional but recommended UEFI firmware files.

It also provides installation hints for your specific operating system.`,
		Run: func(cmd *cobra.Command, args []string) {
			runChecks()
		},
	}
	return cmd
}

func runChecks() {
	fmt.Println("🚀 Running pre-flight checks for q2boot dependencies...")

	kvmOk := checkKVM()
	qemuArches := checkQEMU()
	checkFirmware()
	virtCatOk := checkVirtCat()
	printInstallHints(kvmOk, len(qemuArches) > 0, virtCatOk)

	fmt.Println("\n✅ Pre-flight check complete.")
}

func checkKVM() bool {
	fmt.Println("\n1. Verifying KVM availability (Linux only)")
	if runtime.GOOS != "linux" {
		fmt.Println("   - KVM check is not applicable on this OS.")
		return true
	}

	cpuinfo, err := osReadFile("/proc/cpuinfo")
	if err != nil {
		fmt.Printf("   ❌ Could not read /proc/cpuinfo: %v\n", err)
		return false
	}
	if !strings.Contains(string(cpuinfo), "vmx") && !strings.Contains(string(cpuinfo), "svm") {
		fmt.Println("   ❌ KVM acceleration is not supported by this CPU.")
		fmt.Println("      -> Hint: Ensure virtualization (VT-x or AMD-V) is enabled in your BIOS/UEFI settings.")
		return false
	}
	fmt.Println("   - CPU virtualization support is enabled.")

	if _, err := osStat("/dev/kvm"); os.IsNotExist(err) {
		fmt.Println("   ❌ KVM kernel module is not loaded.")
		fmt.Println("      -> Hint: Run 'sudo modprobe kvm_intel' or 'sudo modprobe kvm_amd'.")
		return false
	}
	fmt.Println("   - KVM kernel module is loaded.")

	file, err := osOpenFile("/dev/kvm", os.O_RDWR, 0)
	if err != nil {
		fmt.Println("   ❌ /dev/kvm device is not accessible by the current user.")
		fmt.Println("      -> Hint: Add your user to the 'kvm' group with 'sudo usermod -aG kvm $USER'.")
		fmt.Println("      -> Note: You may need to log out and back in for the group change to take effect.")
		return false
	}
	file.Close()

	fmt.Println("   ✅ KVM is available and ready to use.")
	return true
}

func checkQEMU() []string {
	fmt.Println("\n2. Checking for QEMU binaries")
	qemuPrefix := "qemu-system-"
	foundArches := make(map[string]struct{})

	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range pathDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), qemuPrefix) {
				if _, err := exec.LookPath(file.Name()); err == nil {
					arch := strings.TrimPrefix(file.Name(), qemuPrefix)
					foundArches[arch] = struct{}{}
				}
			}
		}
	}

	if len(foundArches) == 0 {
		fmt.Println("   ❌ No QEMU system binaries found in your PATH.")
		return nil
	}

	archList := make([]string, 0, len(foundArches))
	for arch := range foundArches {
		archList = append(archList, arch)
	}
	fmt.Printf("   ✅ Found QEMU binaries for architectures: %s\n", strings.Join(archList, ", "))
	return archList
}

func checkFirmware() {
	fmt.Println("\n3. Checking for optional UEFI firmware")

	aarch64EfiDir := "/usr/share/qemu-efi-aarch64/"
	if files, err := os.ReadDir(aarch64EfiDir); err == nil && len(files) > 0 {
		for _, file := range files {
			if !file.IsDir() {
				fmt.Printf("   ✅ Found aarch64 UEFI firmware: %s\n", filepath.Join(aarch64EfiDir, file.Name()))
				return
			}
		}
	}

	qemuDir := "/usr/share/qemu/"
	if files, err := os.ReadDir(qemuDir); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".bin") {
				fmt.Printf("   ✅ Found firmware file: %s\n", filepath.Join(qemuDir, file.Name()))
				return
			}
		}
	}

	fmt.Println("   - UEFI firmware not found in common locations (optional but recommended).")
	fmt.Println("     -> Hint: For aarch64, install 'qemu-efi-aarch64' or 'edk2-aarch64'.")
}

func checkVirtCat() bool {
	fmt.Println("\n4. Checking for virt-cat (for architecture auto-detection)")
	if _, err := exec.LookPath("virt-cat"); err == nil {
		fmt.Println("   ✅ virt-cat is installed and available in your PATH.")
		return true
	}
	fmt.Println("   - virt-cat not found (optional, but needed for auto-detecting image architecture).")
	return false
}

func printInstallHints(kvmOk, qemuFound, virtCatOk bool) {
	if kvmOk && qemuFound && virtCatOk {
		return
	}

	fmt.Println("\n5. Installation Hints")
	osName := runtime.GOOS

	switch osName {
	case "linux":
		distro := getLinuxDistro()
		fmt.Printf("   - Detected OS: Linux (%s)\n", distro)
		switch distro {
		case "ubuntu", "debian":
			fmt.Println("     -> To install QEMU: 'sudo apt update && sudo apt install qemu-system qemu-utils'")
			fmt.Println("     -> To install UEFI firmware: 'sudo apt install qemu-efi-aarch64'")
		case "fedora", "centos", "rhel":
			fmt.Println("     -> To install QEMU: 'sudo dnf install qemu-system-x86 qemu-system-aarch64'")
			if !virtCatOk {
				fmt.Println("     -> To install virt-cat: 'sudo dnf install libguestfs-tools'")
			}
			fmt.Println("     -> To install UEFI firmware: 'sudo dnf install edk2-aarch64'")
		case "arch":
			fmt.Println("     -> To install QEMU and firmware: 'sudo pacman -S qemu-full'")
		default:
			if strings.HasPrefix(distro, "opensuse") {
				fmt.Println("     -> To install QEMU and firmware: 'sudo zypper install qemu-system-x86 qemu-system-aarch64 qemu-uefi-aarch64'")
				break
			}
			fmt.Println("     -> Please use your distribution's package manager to install 'qemu' and related firmware packages.")
		}

		if !virtCatOk && (distro == "ubuntu" || distro == "debian" || strings.HasPrefix(distro, "opensuse")) {
			fmt.Println("     -> To install virt-cat: 'sudo <package_manager> install guestfs-tools'")
		}
	case "darwin":
		fmt.Println("   - Detected OS: macOS")
		fmt.Println("     -> To install QEMU: 'brew install qemu'")
	case "windows":
		fmt.Println("   - Detected OS: Windows")
		fmt.Println("     -> Download and run the QEMU installer from the official website:")
		fmt.Println("        https://www.qemu.org/download/#windows")
	default:
		fmt.Printf("   - OS '%s' is not fully supported for automatic hints.\n", osName)
	}
}

func getLinuxDistro() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if after, ok := strings.CutPrefix(line, "ID="); ok {
			return after
		}
	}
	return "unknown"
}

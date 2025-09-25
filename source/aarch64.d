module aarch64;

import vm;
import std.format;

/**
 * Concrete implementation of the VirtualMachine for aarch64 architecture.
 */
class AARCH64_VM : VirtualMachine
{
    /// Returns the name of the QEMU binary for the specific architecture.
    override string qemuBinary()
    {
        return "qemu-system-aarch64";
    }

    /// Returns an array of architecture-specific QEMU arguments.
    override string[] getArchArgs()
    {
        // This requires a UEFI firmware file. A common path is provided.
        // Users might need to install it via their package manager
        // (e.g., qemu-efi-aarch64 on Debian/Ubuntu).
        return [
            "-machine", "virt",
            "-cpu", "max",
            "-bios", "/usr/share/qemu/aavmf-aarch64-code.bin",
        ];
    }

    /// Returns an array of QEMU arguments for attaching the disk.
    override string[] getDiskArgs()
    {
        return [
            "-drive",
            format("file=%s,if=virtio,cache=writeback,aio=native,discard=unmap,cache.direct=on", diskPath)
        ];
    }

    /// Returns an array of QEMU arguments for networking.
    override string[] getNetworkArgs()
    {
        return [
            "-netdev",
            format("user,id=net0,hostfwd=tcp::%d-:22", sshPort),
            "-device",
            "virtio-net-pci,netdev=net0"
        ];
    }
}

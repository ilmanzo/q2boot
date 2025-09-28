module ppc64le;

import vm;
import std.format;

/**
 * Concrete implementation of the VirtualMachine for PPC64LE architecture.
 */
class PPC64LE_VM : VirtualMachine
{
    /// Returns the name of the QEMU binary for the specific architecture.
    override string qemuBinary()
    {
        return "qemu-system-ppc64";
    }

    /// Returns an array of architecture-specific QEMU arguments.
    override string[] getArchArgs()
    {
        return ["-M", "pseries", "-cpu", "POWER9"];
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

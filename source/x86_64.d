module x86_64;

import vm;
import std.format;

/**
 * Concrete implementation of the VirtualMachine for x86_64 architecture.
 */
class X86_64_VM : VirtualMachine
{
    /// Returns the name of the QEMU binary for the specific architecture.
    override string qemuBinary()
    {
        return "qemu-system-x86_64";
    }

    /// Returns an array of architecture-specific QEMU arguments.
    override string[] getArchArgs()
    {
        return ["-M", "q35", "-enable-kvm", "-cpu", "host"];
    }

    /// Returns an array of QEMU arguments for graphical output.
    override string[] getGraphicalArgs()
    {
        return ["-device", "virtio-vga-gl", "-display", "sdl,gl=on"];
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

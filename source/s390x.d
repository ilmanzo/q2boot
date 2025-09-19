module s390x;

import vm;
import std.format;

/**
 * Concrete implementation of the VirtualMachine for s390x architecture.
 */
class S390X_VM : VirtualMachine
{
    /// Returns the name of the QEMU binary for the specific architecture.
    override string qemuBinary()
    {
        return "qemu-system-s390x";
    }

    /// Returns an array of architecture-specific QEMU arguments.
    override string[] getArchArgs()
    {
        return [
            "-machine", "s390-ccw-virtio",
            "-object", "rng-random,id=rng0,filename=/dev/random",
            "-device", "virtio-rng-ccw,rng=rng0"
        ];
    }

    /// Returns s390x-specific arguments for attaching the disk.
    override string[] getDiskArgs()
    {
        return [
            "-drive",
            format("file=%s,id=disk1,if=none,cache=none,aio=native,discard=unmap", diskPath),
            "-device",
            "virtio-blk-ccw,drive=disk1"
        ];
    }

    /// Returns s390x-specific arguments for networking.
    override string[] getNetworkArgs()
    {
        return [
            "-netdev",
            format("user,id=net1,hostfwd=tcp::%d-:22", sshPort),
            "-device",
            "virtio-net-ccw,netdev=net1"
        ];
    }

    /// Returns s390x-specific arguments for graphical mode.
    /// On s390x, this provides an interactive session in the terminal,
    /// multiplexing the serial console and the QEMU monitor.
    override string[] getGraphicalArgs()
    {
        return [
            "-nographic",
            "-serial", "mon:stdio"
        ];
    }
}

module vm;

import config;
import std.stdio;
import std.process : spawnProcess, wait;
import std.string;
import std.conv;
import std.file;
import std.format;
import std.path;
import std.json;
import std.array;
import std.exception;

version (unittest)
{
    import std.algorithm;
}

/**
 * Validates disk path and throws descriptive errors
 */
void validateDiskPath(string diskPath)
{
    if (diskPath.empty)
    {
        throw new Exception("Disk path cannot be empty. Use -d or --disk.");
    }
    if (!diskPath.exists)
    {
        throw new Exception(format("Disk image not found at '%s'", diskPath));
    }
}

/**
 * Validates VM configuration parameters
 */
void validateVMConfig(const VirtualMachine vm)
{
    if (vm.cpu < 1 || vm.cpu > 32)
    {
        throw new Exception(format("CPU count must be between 1 and 32, got %d", vm.cpu));
    }
    if (vm.ram < 1 || vm.ram > 128)
    {
        throw new Exception(format("RAM must be between 1 and 128 GB, got %d", vm.ram));
    }
    if (vm.sshPort < 1024 || vm.sshPort > 65535)
    {
        throw new Exception(format("SSH port must be between 1024 and 65535, got %d", vm.sshPort));
    }
}

/**
 * Abstract base class for a virtual machine.
 * This class provides common functionality for running a QEMU VM,
 * but requires subclasses to provide architecture-specific details.
 */
abstract class VirtualMachine
{
    string diskPath;
    int cpu;
    int ram;
    bool graphical;
    bool noSnapshot;
    bool confirm;
    ushort sshPort;
    string logFile;

    /**
     * Constructor to initialize the VM with default settings.
     */
    this()
    {
        // Initialize with default values from a default config
        VMConfig config;
        this.cpu = config.cpu;
        this.ram = config.ramGb;
        this.sshPort = config.sshPort;
        this.logFile = config.logFile;
        this.graphical = config.graphical;
        this.noSnapshot = config.writeMode;
        this.confirm = config.confirm;
    }

    /**
     * Loads settings from the specified JSON configuration file.
     */
    void loadFromFile(string path)
    {
        writeln("Found configuration at '", path, "'. Loading defaults...");
        try
        {
            auto text = path.readText();
            auto json = parseJSON(text);
            auto config = parseConfig(json);

            this.cpu = config.cpu;
            this.ram = config.ramGb;
            this.logFile = config.logFile;
            this.sshPort = config.sshPort;
            this.noSnapshot = config.writeMode;
        }
        catch (Exception e)
        {
            stderr.writeln("Warning: Could not parse config file '", path, "': ", e.msg);
        }
    }

    /**
     * Builds the complete QEMU command line arguments.
     * This combines common arguments with architecture-specific ones.
     */
    string[] buildArgs()
    {
        validateDiskPath(diskPath);
        validateVMConfig(this);

        string[] args;

        // Add architecture-specific arguments
        args ~= getArchArgs();

        // Add common arguments
        args ~= ["-smp", to!string(cpu), "-m", format("%dG", ram)];

        // Add disk arguments, allowing for architecture-specific overrides
        args ~= getDiskArgs();

        // Add network arguments
        args ~= getNetworkArgs();

        args ~= ["-audiodev", "none,id=snd0"];

        if (graphical)
        {
            args ~= getGraphicalArgs();
        }
        else
        {
            args ~= ["-nographic"];
            if (!noSnapshot)
            {
                args ~= ["-snapshot"];
            }
            args ~= ["-serial", "stdio", "-monitor", "none"];
        }

        return args;
    }

    /**
     * Runs the virtual machine.
     */
    void run()
    {
        auto args = this.buildArgs();
        auto binary = this.qemuBinary();

        writeln("🚀 Starting QEMU with the following command:");
        writeln(binary, " ", args.join(" "));

        if (confirm)
        {
            write("Press Enter to continue...");
            stdin.readln();
        }

        auto pid = spawnProcess([binary] ~ args);
        auto status = wait(pid);
        if (status != 0)
            stderr.writeln("QEMU exited with a non-zero status: ", status);
    }

    /// Returns the name of the QEMU binary for the specific architecture (e.g., "qemu-system-x86_64").
    protected abstract string qemuBinary();

    /// Returns an array of architecture-specific QEMU arguments.
    protected abstract string[] getArchArgs();

    /**
     * Returns an array of QEMU arguments for attaching the disk.
     * This can be overridden by subclasses for special handling.
     */
    protected string[] getDiskArgs()
    {
        return [
            "-drive",
            format("file=%s,if=virtio,cache=none,aio=native,discard=unmap", diskPath)
        ];
    }

    /**
     * Returns an array of QEMU arguments for networking.
     * This can be overridden by subclasses for special handling.
     */
    protected string[] getNetworkArgs()
    {
        return [
            "-netdev",
            format("user,id=net0,hostfwd=tcp::%d-:22", sshPort),
            "-device",
            "virtio-net-pci,netdev=net0"
        ];
    }

    /**
     * Returns an array of QEMU arguments for graphical mode.
     * This can be overridden by subclasses for special handling.
     */
    protected string[] getGraphicalArgs()
    {
        return [];
    }
}

// ============================================================================
// UNIT TESTS
// ============================================================================

// Dummy VM for testing
version (unittest) class TestVM : VirtualMachine
{
    override string qemuBinary()
    {
        return "qemu-test";
    }

    override string[] getArchArgs()
    {
        return ["-machine", "test"];
    }
}

unittest
{
    writeln("Running validateDiskPath tests...");

    // Test empty path
    assertThrown!Exception(validateDiskPath(""));

    // Test non-existent path
    assertThrown!Exception(validateDiskPath("/non/existent/path.img"));

    // Create a temporary file for testing
    auto tempFile = tempDir ~ "/test_disk.img";
    scope (exit)
        if (tempFile.exists)
            tempFile.remove();

    std.file.write(tempFile, "test disk content");
    assert(tempFile.exists);

    // This should not throw
    validateDiskPath(tempFile);

    writeln("✓ validateDiskPath tests passed");
}

unittest
{
    writeln("Running validateVMConfig tests...");

    auto vm = new TestVM();

    // Test invalid CPU counts
    vm.cpu = 0;
    vm.ram = 4;
    vm.sshPort = 2222;
    assertThrown!Exception(validateVMConfig(vm));

    vm.cpu = 33;
    assertThrown!Exception(validateVMConfig(vm));

    // Test invalid RAM values
    vm.cpu = 2;
    vm.ram = 0;
    assertThrown!Exception(validateVMConfig(vm));

    vm.ram = 129;
    assertThrown!Exception(validateVMConfig(vm));

    // Test invalid SSH port
    vm.ram = 4;
    vm.sshPort = 1023;
    assertThrown!Exception(validateVMConfig(vm));

    // Test valid config
    vm.cpu = 4;
    vm.ram = 8;
    vm.sshPort = 2222;
    // This should not throw
    validateVMConfig(vm);

    vm.sshPort = 65535;
    validateVMConfig(vm);

    writeln("✓ validateVMConfig tests passed");
}

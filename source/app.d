import std.stdio;
import std.getopt;
import std.process : spawnProcess, wait, environment;
import std.string;
import std.conv;
import std.file;
import std.format;
import std.path;
import std.json;
import std.array;

version (unittest)
{
    import std.exception;
    import std.algorithm;
    import std.random;
}

/**
 * Configuration structure for VM settings
 */
struct VMConfig
{
    int cpu = 1;
    int ramGb = 2;
    ushort sshPort = 2222;
    string logFile = "console.log";
    bool headlessSavesChanges = false;
}

/**
 * Creates a default configuration JSON structure
 */
JSONValue createDefaultConfig()
{
    return JSONValue([
        "description": JSONValue("Default configuration for qboot. Edit these values to fit your workflow."),
        "cpu": JSONValue(2),
        "ram_gb": JSONValue(4),
        "ssh_port": JSONValue(2222),
        "log_file": JSONValue("console.log"),
        "headless_saves_changes": JSONValue(false),
    ]);
}

/**
 * Parses configuration from JSON
 */
VMConfig parseConfig(JSONValue json)
{
    VMConfig config;

    if ("cpu" in json)
        config.cpu = json["cpu"].get!int;
    if ("ram_gb" in json)
        config.ramGb = json["ram_gb"].get!int;
    if ("log_file" in json)
        config.logFile = json["log_file"].get!string;
    if ("ssh_port" in json)
        config.sshPort = json["ssh_port"].get!ushort;
    if ("headless_saves_changes" in json)
        config.headlessSavesChanges = json["headless_saves_changes"].get!bool;

    return config;
}

/**
 * Ensures the configuration directory and a default config.json file exist.
 */
void ensureConfigFileExists(string dirPath, string filePath)
{
    if (filePath.exists)
        return;

    try
    {
        writeln("Configuration file not found. Creating a default at '", filePath, "'...");
        dirPath.mkdirRecurse();

        auto defaultConfig = createDefaultConfig();
        std.file.write(filePath, defaultConfig.toPrettyString());

    }
    catch (Exception e)
    {
        stderr.writeln("Warning: Could not create default config file: ", e.msg);
    }
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
void validateVMConfig(const ref VirtualMachine vm)
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

struct VirtualMachine
{
    string diskPath;
    int cpu;
    int ram;
    bool noSnapshot;
    string logFile;
    ushort sshPort;
    bool interactive;

    /// Loads settings from the specified JSON configuration file.
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
            this.noSnapshot = config.headlessSavesChanges;

        }
        catch (Exception e)
        {
            stderr.writeln("Warning: Could not parse config file '", path, "': ", e.msg);
        }
    }

    /// Builds QEMU command line arguments
    string[] buildArgs()
    {
        validateDiskPath(diskPath);
        validateVMConfig(this);

        string[] args;
        args ~= ["-enable-kvm", "-cpu", "host"];
        args ~= ["-smp", to!string(cpu), "-m", format("%dG", ram)];
        args ~= ["-drive", format("file=%s,if=virtio,cache=none,aio=native,discard=unmap", diskPath)];
        args ~= ["-audiodev", "none,id=snd0"];

        if (interactive)
        {
            args ~= ["-display", "default,show-cursor=on"];
        }
        else
        {
            args ~= ["-nographic"];
            if (!noSnapshot)
            {
                args ~= ["-snapshot"];
            }
        }

        args ~= ["-netdev", format("user,id=net0,hostfwd=tcp::%d-:22", sshPort)];
        args ~= ["-device", "virtio-net-pci,netdev=net0"];
        args ~= ["-serial", format("file:%s", logFile)];

        return args;
    }

    /// Runs the virtual machine
    void run()
    {
        auto args = this.buildArgs();
        writeln("🚀 Starting QEMU with the following command:");
        writeln("qemu-system-x86_64 ", args.join(" "));
        writeln("-------------------------------------------------");

        version (unittest)
        {
            // In unit tests, don't actually spawn QEMU
            return;
        }

        string[] fullArgs = ["qemu-system-x86_64"] ~ args;
        auto pid = spawnProcess(fullArgs);
        auto status = wait(pid);
        if (status != 0)
            stderr.writeln("QEMU exited with a non-zero status: ", status);
    }
}

void main(string[] args)
{
    version (unittest)
    {
        // Skip main when running unit tests
        return;
    }

    auto vm = VirtualMachine(
        cpu: 1, ram: 2, logFile: "console.log",
        sshPort: 2222, noSnapshot: false, interactive: false
    );

    string configDir = buildPath(environment.get("HOME"), ".config", "qboot");
    string configFile = buildPath(configDir, "config.json");

    ensureConfigFileExists(configDir, configFile);

    if (configFile.exists)
    {
        vm.loadFromFile(configFile);
    }

    try
    {
        auto helpInfo = getopt(
            args,
            "disk|d", &vm.diskPath,
            "cpu|c", &vm.cpu,
            "ram|r", &vm.ram,
            "interactive|i", &vm.interactive,
            "no-snapshot|S", &vm.noSnapshot,
            "log|l", &vm.logFile,
            "ssh-port", &vm.sshPort
        );

        if (helpInfo.helpWanted || vm.diskPath.empty)
        {
            defaultGetoptPrinter("A handy QEMU VM launcher.", helpInfo.options);
            return;
        }

        vm.run();

    }
    catch (Exception e)
    {
        stderr.writeln("Error: ", e.msg);
    }
}

// ============================================================================
// UNIT TESTS
// ============================================================================

unittest
{
    writeln("Running createDefaultConfig tests...");

    auto config = createDefaultConfig();
    assert(config["cpu"].get!int == 2);
    assert(config["ram_gb"].get!int == 4);
    assert(config["ssh_port"].get!ushort == 2222);
    assert(config["log_file"].get!string == "console.log");
    assert(config["headless_saves_changes"].get!bool == false);

    writeln("✓ createDefaultConfig tests passed");
}

unittest
{
    writeln("Running parseConfig tests...");

    // Test valid config
    auto json = JSONValue([
        "cpu": JSONValue(4),
        "ram_gb": JSONValue(8),
        "ssh_port": JSONValue(3333),
        "log_file": JSONValue("test.log"),
        "headless_saves_changes": JSONValue(true)
    ]);

    auto config = parseConfig(json);
    assert(config.cpu == 4);
    assert(config.ramGb == 8);
    assert(config.sshPort == 3333);
    assert(config.logFile == "test.log");
    assert(config.headlessSavesChanges == true);

    // Test partial config (should use defaults)
    auto partialJson = JSONValue([
        "cpu": JSONValue(6)
    ]);

    auto partialConfig = parseConfig(partialJson);
    assert(partialConfig.cpu == 6);
    assert(partialConfig.ramGb == 2); // default
    assert(partialConfig.sshPort == 2222); // default

    writeln("✓ parseConfig tests passed");
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

    VirtualMachine vm;

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

    vm.sshPort = 65535;
    vm.ram = 4;
    // Test valid config
    vm.cpu = 4;
    vm.ram = 8;
    vm.sshPort = 2222;
    // This should not throw
    validateVMConfig(vm);

    writeln("✓ validateVMConfig tests passed");
}

unittest
{
    writeln("Running VirtualMachine.buildArgs tests...");

    // Create a temporary disk file
    auto tempFile = tempDir ~ "/test_vm_disk.img";
    scope (exit)
        if (tempFile.exists)
            tempFile.remove();
    std.file.write(tempFile, "test disk content");

    VirtualMachine vm;
    vm.diskPath = tempFile;
    vm.cpu = 2;
    vm.ram = 4;
    vm.sshPort = 2222;
    vm.logFile = "test.log";
    vm.interactive = false;
    vm.noSnapshot = false;

    auto args = vm.buildArgs();

    // Check that essential arguments are present
    assert(args.canFind("-enable-kvm"));
    assert(args.canFind("-cpu"));
    assert(args.canFind("host"));
    assert(args.canFind("-smp"));
    assert(args.canFind("2"));
    assert(args.canFind("-m"));
    assert(args.canFind("4G"));
    assert(args.canFind("-nographic"));
    assert(args.canFind("-snapshot"));

    // Test interactive mode
    vm.interactive = true;
    auto interactiveArgs = vm.buildArgs();
    assert(interactiveArgs.canFind("-display"));
    assert(interactiveArgs.canFind("default,show-cursor=on"));
    assert(!interactiveArgs.canFind("-nographic"));

    // Test no snapshot mode
    vm.interactive = false;
    vm.noSnapshot = true;
    auto noSnapshotArgs = vm.buildArgs();
    assert(!noSnapshotArgs.canFind("-snapshot"));

    writeln("✓ VirtualMachine.buildArgs tests passed");
}

unittest
{
    writeln("Running VirtualMachine.loadFromFile tests...");

    // Create a temporary config file
    auto tempConfigFile = tempDir ~ "/test_config.json";
    scope (exit)
        if (tempConfigFile.exists)
            tempConfigFile.remove();

    auto testConfig = JSONValue([
        "cpu": JSONValue(8),
        "ram_gb": JSONValue(16),
        "ssh_port": JSONValue(3333),
        "log_file": JSONValue("custom.log"),
        "headless_saves_changes": JSONValue(true)
    ]);

    std.file.write(tempConfigFile, testConfig.toPrettyString());

    VirtualMachine vm;
    vm.loadFromFile(tempConfigFile);

    assert(vm.cpu == 8);
    assert(vm.ram == 16);
    assert(vm.sshPort == 3333);
    assert(vm.logFile == "custom.log");
    assert(vm.noSnapshot == true);

    writeln("✓ VirtualMachine.loadFromFile tests passed");
}

unittest
{
    writeln("Running ensureConfigFileExists tests...");

    // Create a temporary directory
    auto tempTestDir = tempDir ~ "/qboot_test_" ~ to!string(uniform(1000, 9999));
    auto tempConfigFile = tempTestDir ~ "/config.json";

    scope (exit)
    {
        if (tempConfigFile.exists)
            tempConfigFile.remove();
        if (tempTestDir.exists)
            tempTestDir.rmdir();
    }

    // Initially, neither directory nor file should exist
    assert(!tempTestDir.exists);
    assert(!tempConfigFile.exists);

    // Call the function
    ensureConfigFileExists(tempTestDir, tempConfigFile);

    // Now both should exist
    assert(tempTestDir.exists);
    assert(tempConfigFile.exists);

    // Verify the content is valid JSON
    auto content = tempConfigFile.readText();
    auto json = parseJSON(content);
    assert(json["cpu"].get!int == 2);

    // Call again - should not overwrite
    auto originalContent = tempConfigFile.readText();
    ensureConfigFileExists(tempTestDir, tempConfigFile);
    assert(tempConfigFile.readText() == originalContent);

    writeln("✓ ensureConfigFileExists tests passed");
}

// Integration test
unittest
{
    writeln("Running integration test...");

    // Create temporary files
    auto tempTestDir = tempDir ~ "/qboot_integration_test";
    auto tempDisk = tempTestDir ~ "/test.img";
    auto tempConfig = tempTestDir ~ "/config.json";

    scope (exit)
    {
        if (tempDisk.exists)
            tempDisk.remove();
        if (tempConfig.exists)
            tempConfig.remove();
        if (tempTestDir.exists)
            tempTestDir.rmdir();
    }

    tempTestDir.mkdirRecurse();
    std.file.write(tempDisk, "fake disk image");

    // Create VM and set it up
    VirtualMachine vm;
    vm.diskPath = tempDisk;
    vm.cpu = 2;
    vm.ram = 4;
    vm.sshPort = 2222;
    vm.logFile = "integration_test.log";
    vm.interactive = false;
    vm.noSnapshot = false;

    // Should be able to build args without throwing
    auto args = vm.buildArgs();
    assert(args.length > 0);

    // Should be able to run (though it won't actually spawn QEMU in unittest)
    vm.run();

    writeln("✓ Integration test passed");
}

static this()
{
    version (unittest)
    {
        writeln("🧪 Starting qboot unit tests...");
    }
}

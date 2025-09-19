import vm;
import config;
import x86_64;
import ppc64le;
import s390x;
import std.stdio;
import std.getopt;
import std.process : environment;
import std.string;
import std.conv;
import std.file;
import std.path;
import std.json;
import std.exception;

version (unittest)
{
    import std.algorithm;
    import std.random;
}

void main(string[] args)
{
    version (unittest)
    {
        // Skip main when running unit tests
        return;
    }

    string configDir = buildPath(environment.get("HOME"), ".config", "qboot");
    string configFile = buildPath(configDir, "config.json");

    ensureConfigFileExists(configDir, configFile);

    VMConfig config;
    if (configFile.exists)
    {
        try
        {
            auto text = configFile.readText();
            auto json = parseJSON(text);
            config = parseConfig(json);
        }
        catch (Exception e)
        {
            stderr.writeln("Warning: Could not parse config file '", configFile, "': ", e.msg);
        }
    }

    string diskPath, cpu, ram, logFile, arch;
    bool interactive, noSnapshot;
    ushort sshPort;

    try
    {
        auto helpInfo = getopt(args, "arch|a", &arch, "disk|d", &diskPath, "cpu|c", &cpu, "ram|r",
            &ram, "interactive|i", &interactive, "no-snapshot|S", &noSnapshot,
            "log|l", &logFile, "ssh-port", &sshPort, );

        if (helpInfo.helpWanted || diskPath.empty)
        {
            writeln("qboot - A handy QEMU VM launcher");
            writeln();
            writeln("USAGE:");
            writeln("    qboot [OPTIONS] --disk <path>");
            writeln();
            writeln("EXAMPLES:");
            writeln("    qboot --disk ubuntu.qcow2");
            writeln("    qboot --disk fedora.img --cpu 4 --ram 8 --interactive");
            writeln("    qboot --arch ppc64le --disk debian.qcow2 --ssh-port 2223");
            writeln("    qboot --disk test.img --no-snapshot --log vm.log");
            writeln();
            writeln("OPTIONS:");

            // Custom help formatting with proper option descriptions
            writeln("    -a, --arch <arch>        Target architecture (default: x86_64, options: x86_64, ppc64le)");
            writeln("    -d, --disk <path>        Path to disk image file (required)");
            writeln("    -c, --cpu <count>        Number of CPU cores (default: 2)");
            writeln("    -r, --ram <gb>           RAM size in GB (default: 4)");
            writeln("    -i, --interactive        Enable interactive mode with QEMU monitor");
            writeln("    -S, --no-snapshot        Disable snapshot mode (changes will be saved to disk)");
            writeln("    -l, --log <file>         Log file path (default: console.log)");
            writeln("        --ssh-port <port>    SSH port forwarding (default: 2222)");
            writeln("    -h, --help               Show this help message");

            writeln();
            writeln("Configuration file: ~/.config/qboot/config.json");
            writeln("Command line options override configuration file settings.");
            return;
        }
    }
    catch (Exception e)
    {
        stderr.writeln("Error: ", e.msg);
        return;
    }

    if (!arch.empty)
    {
        config.arch = arch;
    }

    VirtualMachine vm;
    switch (config.arch)
    {
        case "x86_64":
            vm = new X86_64_VM();
            break;
        case "ppc64le":
            vm = new PPC64LE_VM();
            break;
        case "s390x":
            vm = new S390X_VM();
            break;

        default:
            stderr.writeln("Error: Unsupported architecture '", config.arch,
                "' in config file.");
            return;
    }

    if (configFile.exists)
    {
        vm.loadFromFile(configFile);
    }

    vm.diskPath = diskPath;
    if (!cpu.empty)
        vm.cpu = cpu.to!int;
    if (!ram.empty)
        vm.ram = ram.to!int;
    vm.interactive = interactive;
    vm.noSnapshot = noSnapshot;
    if (!logFile.empty)
        vm.logFile = logFile;
    if (sshPort != 0)
        vm.sshPort = sshPort;

    try
    {
        vm.run();
    }
    catch (Exception e)
    {
        stderr.writeln("Error: ", e.msg);
    }
}

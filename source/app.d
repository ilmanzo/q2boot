import vm;
import config;
import x86_64;
import ppc64le;
import s390x;
import aarch64;
import std.stdio;
import std.getopt;
import std.process : environment;
import std.string;
import std.conv;
import std.file;
import std.path;
import std.json;
import std.exception;

void main(string[] args)
{

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
            stderr.writeln("Error parsing config file: ", e.msg);
        }
    }

    string diskPath, cpu, ram, logFile, arch;
    bool graphical, noSnapshot, confirm;
    ushort sshPort;

    try
    {
        auto options = getopt(
            args,
            "d", "disk", &diskPath,
            "c", "cpu", &cpu,
            "r", "ram", &ram,
            "g", "graphical", &graphical,
            "w", "write-mode", &noSnapshot,
            "p", "ssh-port", &sshPort,
            "l", "log-file", &logFile,
            "a", "arch", &arch,
            "confirm", &confirm,
            "help", {
            writeln("Usage: qboot [options]");
            writeln("Options:");
            writeln("  -d, --disk <path>      Path to the qcow2 disk image (required)");
            writeln("  -c, --cpu <cores>      Number of CPU cores (default: ", config.cpu, ")");
            writeln("  -r, --ram <GB>         Amount of RAM in GB (default: ", config.ramGb, ")");
            writeln("  -g, --graphical        Enable graphical console (default: disabled)");
            writeln("  -w, --write-mode       Enable write mode (changes are saved to disk)");
            writeln("  -p, --ssh-port <port>  Host port for SSH forwarding (default: ", config.sshPort, ")");
            writeln("  -l, --log-file <path>  Path to the log file (default: ", config.logFile, ")");
            writeln("  -a, --arch <arch>      Architecture (x86_64, ppc64le, s390x, aarch64) (default: ", config.arch, ")");
            writeln("      --confirm          Show command and wait for keypress before starting");
            writeln("      --help             Show this help message");
            return;
        }
        );
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
    case "aarch64":
        vm = new AARCH64_VM();
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
    {
        try
        {
            vm.cpu = cpu.to!int;
        }
        catch (ConvException e)
        {
            stderr.writeln("Error: Invalid CPU value '", cpu, "'. Please provide a numeric value.");
            return;
        }
    }
    if (!ram.empty)
    {
        try
        {
            vm.ram = ram.to!int;
        }
        catch (ConvException e)
        {
            stderr.writeln("Error: Invalid RAM value '", ram, "'. Please provide a numeric value.");
            return;
        }
    }
    vm.graphical = graphical;
    vm.noSnapshot = noSnapshot;
    vm.confirm = confirm;
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

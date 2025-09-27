module config;

import std.stdio;
import std.json;
import std.file;
import std.path;
import std.conv;

version (unittest)
{
    import std.exception;
    import std.random;
}

/**
 * Configuration structure for VM settings
 */
struct VMConfig
{
    string arch;
    int cpu;
    int ramGb;
    ushort sshPort;
    string logFile;
    bool graphical;
    bool writeMode;
    bool confirm;
}

/**
 * Creates a default configuration JSON structure
 */
JSONValue createDefaultConfig()
{
    return JSONValue([
        "cpu": JSONValue(2),
        "ram_gb": JSONValue(2),
        "ssh_port": JSONValue(2222),
        "log_file": JSONValue("console.log"),
        "write_mode": JSONValue(false),
        "graphical": JSONValue(false),
        "confirm": JSONValue(false),
        "arch": JSONValue("x86_64"),
    ]);
}

/**
 * Parses configuration from JSON
 */
VMConfig parseConfig(JSONValue json)
{
    VMConfig config;

    // Set defaults first
    config.cpu = 2;
    config.ramGb = 2;
    config.sshPort = 2222;
    config.logFile = "console.log";
    config.writeMode = false;
    config.graphical = false;
    config.confirm = false;
    config.arch = "x86_64";

    // Override with values from JSON if they exist
    if ("cpu" in json)
        config.cpu = to!int(json["cpu"].get!long);
    if ("ram_gb" in json)
        config.ramGb = to!int(json["ram_gb"].get!long);
    if ("ssh_port" in json)
        config.sshPort = to!ushort(json["ssh_port"].get!long);
    if ("log_file" in json)
        config.logFile = json["log_file"].get!string;
    if ("write_mode" in json)
        config.writeMode = json["write_mode"].get!bool;
    if ("graphical" in json)
        config.graphical = json["graphical"].get!bool;
    if ("confirm" in json)
        config.confirm = json["confirm"].get!bool;
    if ("arch" in json)
        config.arch = json["arch"].get!string;

    return config;
}

/**
 * Ensures that the configuration directory and file exist.
 * If they don't, it creates them with default values.
 */
void ensureConfigFileExists(string configDir, string configFile)
{
    if (!configDir.exists)
    {
        writeln("Creating config directory at '", configDir, "'");
        configDir.mkdirRecurse();
    }

    if (!configFile.exists)
    {
        writeln("No config file found. Creating default config at '", configFile, "'");
        auto defaultConfig = createDefaultConfig();
        std.file.write(configFile, toJSON(defaultConfig, true));
    }
}

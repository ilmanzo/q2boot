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
    string description;
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
        "description": JSONValue("Default configuration for qboot. Edit these values to fit your workflow."),
        "cpu": JSONValue(2),
        "ram_gb": JSONValue(4),
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
    config.ramGb = 4;
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

// ============================================================================
// UNIT TESTS
// ============================================================================

unittest
{
    writeln("Running createDefaultConfig tests...");
    auto defaultConfig = createDefaultConfig();
    assert(defaultConfig["cpu"].get!long == 2);
    assert(defaultConfig["ram_gb"].get!long == 4);
    assert(defaultConfig["ssh_port"].get!long == 2222);
    assert(defaultConfig["log_file"].get!string == "console.log");
    assert(defaultConfig["write_mode"].get!bool == false);
    assert(defaultConfig["graphical"].get!bool == false);
    assert(defaultConfig["confirm"].get!bool == false);
    assert(defaultConfig["arch"].get!string == "x86_64");
    writeln("✓ createDefaultConfig tests passed");
}

unittest
{
    writeln("Running parseConfig tests...");

    // Test full config
    auto fullJson = JSONValue([
        "cpu": JSONValue(4),
        "ram_gb": JSONValue(8),
        "ssh_port": JSONValue(3333),
        "log_file": JSONValue("test.log"),
        "write_mode": JSONValue(true),
        "graphical": JSONValue(true),
        "confirm": JSONValue(true),
        "arch": JSONValue("aarch64")
    ]);

    auto fullConfig = parseConfig(fullJson);
    assert(fullConfig.cpu == 4);
    assert(fullConfig.ramGb == 8);
    assert(fullConfig.sshPort == 3333);
    assert(fullConfig.logFile == "test.log");
    assert(fullConfig.writeMode == true);
    assert(fullConfig.graphical == true);
    assert(fullConfig.confirm == true);
    assert(fullConfig.arch == "aarch64");

    // Test partial config (should use defaults for missing keys)
    auto partialJson = JSONValue(["cpu": JSONValue(1)]);
    auto partialConfig = parseConfig(partialJson);
    assert(partialConfig.cpu == 1);
    assert(partialConfig.ramGb == 4); // Should be default
    assert(partialConfig.sshPort == 2222); // Should be default
    assert(partialConfig.logFile == "console.log"); // Should be default
    assert(partialConfig.writeMode == false); // Should be default
    assert(partialConfig.graphical == false); // Should be default
    assert(partialConfig.confirm == false); // Should be default
    assert(partialConfig.arch == "x86_64"); // Should be default

    writeln("✓ parseConfig tests passed");
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
    assert(json["cpu"].get!long == 2);
    assert(json["arch"].get!string == "x86_64");

    // Call again - should not overwrite
    auto originalContent = tempConfigFile.readText();
    ensureConfigFileExists(tempTestDir, tempConfigFile);
    assert(tempConfigFile.readText() == originalContent);

    writeln("✓ ensureConfigFileExists tests passed");
}

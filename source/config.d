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
        "headless_saves_changes": JSONValue(false),
        "arch": JSONValue("x86_64"),
    ]);
}

/**
 * Parses configuration from JSON
 */
VMConfig parseConfig(JSONValue json)
{
    VMConfig config;

    if ("cpu" in json)
        config.cpu = to!int(json["cpu"].get!long);
    if ("ram_gb" in json)
        config.ramGb = to!int(json["ram_gb"].get!long);
    if ("log_file" in json)
        config.logFile = json["log_file"].get!string;
    if ("ssh_port" in json)
        config.sshPort = to!ushort(json["ssh_port"].get!long);

    if ("headless_saves_changes" in json)
        config.headlessSavesChanges = json["headless_saves_changes"].get!bool;
    if ("arch" in json)
        config.arch = json["arch"].get!string;

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

// ============================================================================
// UNIT TESTS
// ============================================================================

unittest
{
    writeln("Running createDefaultConfig tests...");

    auto config = createDefaultConfig();
    assert(config["cpu"].get!long == 2);
    assert(config["ram_gb"].get!long == 4);
    assert(config["ssh_port"].get!long == 2222);
    assert(config["log_file"].get!string == "console.log");
    assert(config["headless_saves_changes"].get!bool == false);
    assert(config["arch"].get!string == "x86_64");

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
        "headless_saves_changes": JSONValue(true),
        "arch": JSONValue("aarch64")
    ]);

    auto config = parseConfig(json);
    assert(config.cpu == 4);
    assert(config.ramGb == 8);
    assert(config.sshPort == 3333);
    assert(config.logFile == "test.log");
    assert(config.headlessSavesChanges == true);
    assert(config.arch == "aarch64");

    // Test partial config (should use defaults)
    auto partialJson = JSONValue(["cpu": JSONValue(6)]);

    auto partialConfig = parseConfig(partialJson);
    assert(partialConfig.cpu == 6);
    assert(partialConfig.ramGb == 2); // default
    assert(partialConfig.sshPort == 2222); // default
    assert(partialConfig.arch == "x86_64"); // default

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

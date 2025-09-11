#!/usr/bin/env rdmd
import std.stdio;
import std.process;
import std.string;
import std.conv;

/**
 * Test runner for qboot
 * This script runs all unit tests and provides a summary
 */

void main(string[] args)
{
    writeln("🧪 QBoot Unit Test Runner");
    writeln("========================");

    // Run unit tests using dub
    auto result = execute(["dub", "test", "--build=unittest"]);

    if (result.status == 0)
    {
        writeln("\n✅ All tests passed successfully!");
        writeln("Output:");
        writeln(result.output);
    }
    else
    {
        writeln("\n❌ Tests failed!");
        writeln("Error output:");
        writeln(result.output);
    }

    writeln("\nTest Summary:");
    writeln("- Configuration parsing tests");
    writeln("- Disk path validation tests");
    writeln("- VM configuration validation tests");
    writeln("- Command line argument building tests");
    writeln("- File I/O tests");
    writeln("- Integration tests");
}

//go:build !e2e

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ilmanzo/q2boot/internal/config"
	"github.com/ilmanzo/q2boot/internal/vm"
	"github.com/spf13/cobra"
)

func TestDefaultArchitecture(t *testing.T) {
	// Create a temporary file to act as a disk image
	tempDir, err := os.MkdirTemp("", "q2boot-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test.img")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Use a mock VM creator to prevent actual QEMU execution
	originalCreator := vm.CreateVM
	vm.CreateVM = func(arch string) (vm.VM, error) {
		// We can return a mock that does nothing on Run()
		return vm.NewMockVM(), nil
	}
	defer func() { vm.CreateVM = originalCreator }()

	// We need to mock the final Run() to avoid executing QEMU
	// For this test, we'll create a custom run function that just checks the config.
	testCfg := config.DefaultConfig()

	// This is our testable "run" function that will be executed by the command.
	// It captures the final state of the configuration.
	testRunE := func(cmd *cobra.Command, args []string) error {
		// The core logic from main.go is now in runQ2BootE
		// We pass our test config to it.
		return runQ2BootE(cmd, args, testCfg)
	}

	// Temporarily replace the command's RunE function with our test function
	originalRunE := rootCmd.RunE
	rootCmd.RunE = testRunE
	defer func() { rootCmd.RunE = originalRunE }()

	// Execute the root command with only the required disk flag
	rootCmd.SetArgs([]string{"--disk", tempFile})

	// We only care about the config validation part, so we ignore the QEMU execution error
	_ = rootCmd.Execute()

	if testCfg.Arch != "x86_64" {
		t.Errorf("Expected architecture to default to 'x86_64', but got '%s'", testCfg.Arch)
	}
}

func TestFlagOverridesConfig(t *testing.T) {
	// 1. Create a temporary config file with specific values
	tempDir, err := os.MkdirTemp("", "q2boot-test-config")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set the test config directory so initConfig will use it
	testConfigDir = tempDir
	defer func() { testConfigDir = "" }() // Reset after test

	configContent := `{
		"arch": "aarch64",
		"cpu": 8,
		"ram_gb": 16,
		"ssh_port": 3333,
		"graphical": true
	}`
	configPath := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Create a dummy disk file
	tempDisk := filepath.Join(tempDir, "disk.img")
	if err := os.WriteFile(tempDisk, []byte("disk"), 0644); err != nil {
		t.Fatalf("Failed to create temp disk file: %v", err)
	}

	// Use a mock VM creator to prevent actual QEMU execution
	originalCreator := vm.CreateVM
	vm.CreateVM = func(arch string) (vm.VM, error) {
		// We can return a mock that does nothing on Run()
		return vm.NewMockVM(), nil
	}
	defer func() { vm.CreateVM = originalCreator }()

	// 2. Setup the test run
	testRunE := func(cmd *cobra.Command, args []string) error {
		// Manually call initConfig to load from our temp file
		initConfig()
		// The core logic will unmarshal into the global cfg, then we run our logic
		return runQ2BootE(cmd, args, cfg)
	}

	originalRunE := rootCmd.RunE
	rootCmd.RunE = testRunE
	defer func() { rootCmd.RunE = originalRunE }()

	// 3. Execute with flags that override the config file
	rootCmd.SetArgs([]string{
		"--disk", tempDisk,
		"--arch", "ppc64le", // Override "aarch64"
		"--cpu", "4", // Override 8
		"--ram", "8", // Override 16
		"--ssh-port", "4444", // Override 3333
		"--graphical=false", // Override true
	})

	// We only care about config validation, so ignore QEMU execution error
	_ = rootCmd.Execute()

	// 4. Assert that flag values were used
	if cfg.Arch != "ppc64le" {
		t.Errorf("Expected arch to be 'ppc64le' from flag, but got '%s'", cfg.Arch)
	}
	if cfg.CPU != 4 {
		t.Errorf("Expected cpu to be 4 from flag, but got %d", cfg.CPU)
	}
	if cfg.RAMGb != 8 {
		t.Errorf("Expected ram_gb to be 8 from flag, but got %d", cfg.RAMGb)
	}
	if cfg.SSHPort != 4444 {
		t.Errorf("Expected ssh_port to be 4444 from flag, but got %d", cfg.SSHPort)
	}
	if cfg.Graphical != false {
		t.Errorf("Expected graphical to be false from flag, but got %t", cfg.Graphical)
	}
}

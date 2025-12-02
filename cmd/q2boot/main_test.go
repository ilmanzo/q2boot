//go:build !e2e

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ilmanzo/q2boot/internal/detector"
	"github.com/ilmanzo/q2boot/internal/vm"
	"github.com/spf13/cobra"
)

// setupTest re-initializes the command and configuration for each test run,
// ensuring test isolation.
func setupTest(t *testing.T) {
	// Reset the root command and its flags to a clean state
	rootCmd = &cobra.Command{
		Use:  "q2boot [flags] <disk_image_path>",
		Args: cobra.ExactArgs(1),
		RunE: runQ2Boot,
	}
	// Re-run the flag setup to define all persistent flags on the new command
	setupFlags()
	// Reset the global config object
	initConfig()
}
func TestDefaultArchitecture(t *testing.T) {
	setupTest(t)

	// Use a mock VM creator to prevent actual QEMU execution
	originalCreator := vm.CreateVM
	vm.CreateVM = func(arch string) (vm.VM, error) {
		return vm.NewMockVM(), nil
	}
	defer func() { vm.CreateVM = originalCreator }()

	// Mock architecture detection to avoid running virt-cat in unit tests
	originalDetector := detector.DetectArchitecture
	expectedErr := "detection failed"
	detector.DetectArchitecture = func(diskPath string) (string, error) {
		// Simulate a failed detection
		return "", fmt.Errorf(expectedErr)
	}
	defer func() { detector.DetectArchitecture = originalDetector }()

	// Create a dummy disk file
	tempFile := filepath.Join(t.TempDir(), "test.img")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	rootCmd.SetArgs([]string{tempFile})

	// Execute the command and expect an error
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected an error due to failed architecture detection, but got nil")
	}

	// Check if the error is the one we expect from our mock
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain '%s', but got: %v", expectedErr, err)
	}
}

func TestFlagOverridesConfig(t *testing.T) {
	setupTest(t)

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
		// Return a mock that bypasses QEMU binary validation
		mock := vm.NewMockVM()
		mock.ValidateFunc = func() error { return nil }
		return mock, nil
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
		tempDisk,
		// Use s390x to match the test output log, ensuring consistency.
		"--arch", "s390x", // Override "aarch64"
		"--cpu", "4", // Override 8
		"--ram", "8", // Override 16
		"--ssh-port", "4444", // Override 3333
		"--graphical=false", // Override true
	})

	// We only care about config validation, so ignore QEMU execution error
	_ = rootCmd.Execute()

	// 4. Assert that flag values were used
	if cfg.Arch != "s390x" {
		t.Errorf("Expected arch to be 's390x' from flag, but got '%s'", cfg.Arch)
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

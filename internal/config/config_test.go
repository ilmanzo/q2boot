package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Arch != "x86_64" {
		t.Errorf("Expected arch to be x86_64, got %s", cfg.Arch)
	}

	if cfg.CPU != 2 {
		t.Errorf("Expected CPU to be 2, got %d", cfg.CPU)
	}

	if cfg.RAMGb != 2 {
		t.Errorf("Expected RAMGb to be 2, got %d", cfg.RAMGb)
	}

	if cfg.SSHPort != 2222 {
		t.Errorf("Expected SSHPort to be 2222, got %d", cfg.SSHPort)
	}

	if cfg.LogFile != "qboot.log" {
		t.Errorf("Expected LogFile to be qboot.log, got %s", cfg.LogFile)
	}

	if cfg.WriteMode != false {
		t.Errorf("Expected WriteMode to be false, got %t", cfg.WriteMode)
	}

	if cfg.Graphical != false {
		t.Errorf("Expected Graphical to be false, got %t", cfg.Graphical)
	}

	if cfg.Confirm != false {
		t.Errorf("Expected Confirm to be false, got %t", cfg.Confirm)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *VMConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid CPU - too low",
			config: &VMConfig{
				Arch:    "x86_64",
				CPU:     0,
				RAMGb:   4,
				SSHPort: 2222,
			},
			wantErr: true,
		},
		{
			name: "invalid CPU - too high",
			config: &VMConfig{
				Arch:    "x86_64",
				CPU:     33,
				RAMGb:   4,
				SSHPort: 2222,
			},
			wantErr: true,
		},
		{
			name: "invalid RAM - too low",
			config: &VMConfig{
				Arch:    "x86_64",
				CPU:     2,
				RAMGb:   0,
				SSHPort: 2222,
			},
			wantErr: true,
		},
		{
			name: "invalid RAM - too high",
			config: &VMConfig{
				Arch:    "x86_64",
				CPU:     2,
				RAMGb:   129,
				SSHPort: 2222,
			},
			wantErr: true,
		},
		{
			name: "invalid SSH port - too low",
			config: &VMConfig{
				Arch:    "x86_64",
				CPU:     2,
				RAMGb:   4,
				SSHPort: 1023,
			},
			wantErr: true,
		},
		{
			name: "valid SSH port - maximum",
			config: &VMConfig{
				Arch:    "x86_64",
				CPU:     2,
				RAMGb:   4,
				SSHPort: 65535,
			},
			wantErr: false,
		},
		{
			name: "invalid SSH port - below minimum",
			config: &VMConfig{
				Arch:    "x86_64",
				CPU:     2,
				RAMGb:   4,
				SSHPort: 1023,
			},
			wantErr: true,
		},
		{
			name: "invalid architecture",
			config: &VMConfig{
				Arch:    "invalid",
				CPU:     2,
				RAMGb:   4,
				SSHPort: 2222,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadAndSaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "qboot-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "config.json")

	// Create a test config
	testConfig := &VMConfig{
		Arch:      "aarch64",
		CPU:       4,
		RAMGb:     8,
		SSHPort:   2223,
		LogFile:   "test.log",
		WriteMode: true,
		Graphical: true,
		Confirm:   true,
	}

	// Save the config
	err = SaveConfig(testConfig, configFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config
	loadedConfig, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Compare the configs
	if loadedConfig.Arch != testConfig.Arch {
		t.Errorf("Expected arch %s, got %s", testConfig.Arch, loadedConfig.Arch)
	}
	if loadedConfig.CPU != testConfig.CPU {
		t.Errorf("Expected CPU %d, got %d", testConfig.CPU, loadedConfig.CPU)
	}
	if loadedConfig.RAMGb != testConfig.RAMGb {
		t.Errorf("Expected RAMGb %d, got %d", testConfig.RAMGb, loadedConfig.RAMGb)
	}
	if loadedConfig.SSHPort != testConfig.SSHPort {
		t.Errorf("Expected SSHPort %d, got %d", testConfig.SSHPort, loadedConfig.SSHPort)
	}
	if loadedConfig.LogFile != testConfig.LogFile {
		t.Errorf("Expected LogFile %s, got %s", testConfig.LogFile, loadedConfig.LogFile)
	}
	if loadedConfig.WriteMode != testConfig.WriteMode {
		t.Errorf("Expected WriteMode %t, got %t", testConfig.WriteMode, loadedConfig.WriteMode)
	}
	if loadedConfig.Graphical != testConfig.Graphical {
		t.Errorf("Expected Graphical %t, got %t", testConfig.Graphical, loadedConfig.Graphical)
	}
	if loadedConfig.Confirm != testConfig.Confirm {
		t.Errorf("Expected Confirm %t, got %t", testConfig.Confirm, loadedConfig.Confirm)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	_, err := LoadConfig("/non/existent/file.json")
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
}

func TestSaveConfigInvalidPath(t *testing.T) {
	config := DefaultConfig()
	err := SaveConfig(config, "/invalid/path/config.json")
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestEnsureConfigExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "qboot-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configDir := filepath.Join(tempDir, "qboot")
	configFile := filepath.Join(configDir, "config.json")

	// Ensure config exists (should create both dir and file)
	err = EnsureConfigExists(configDir, configFile)
	if err != nil {
		t.Fatalf("Failed to ensure config exists: %v", err)
	}

	// Check that directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	// Check that file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify the content is valid JSON
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	var config VMConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Errorf("Created config file contains invalid JSON: %v", err)
	}
}

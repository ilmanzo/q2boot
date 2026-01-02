package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Architecture defaults to empty string (triggers auto-detection)
	if cfg.Arch != "" {
		t.Errorf("Expected arch to be empty (auto-detect), got %s", cfg.Arch)
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

	if cfg.LogFile != "q2boot.log" {
		t.Errorf("Expected LogFile to be q2boot.log, got %s", cfg.LogFile)
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
	// Create a temporary file for disk path testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.img")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file for testing: %v", err)
	}

	tests := []struct {
		name    string
		config  *VMConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      2,
				RAMGb:    4,
				SSHPort:  2222,
				DiskPath: tempFile,
			},
			wantErr: false,
		},
		{
			name: "invalid CPU - too low",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      0,
				RAMGb:    4,
				SSHPort:  2222,
				DiskPath: tempFile,
			},
			wantErr: true,
		},
		{
			name: "invalid CPU - too high",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      65,
				RAMGb:    4,
				SSHPort:  2222,
				DiskPath: tempFile,
			},
			wantErr: true,
		},
		{
			name: "invalid RAM - too low",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      2,
				RAMGb:    0,
				SSHPort:  2222,
				DiskPath: tempFile,
			},
			wantErr: true,
		},
		{
			name: "invalid RAM - too high",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      2,
				RAMGb:    129,
				SSHPort:  2222,
				DiskPath: tempFile,
			},
			wantErr: true,
		},
		{
			name: "invalid SSH port - too low",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      2,
				RAMGb:    4,
				SSHPort:  1023,
				DiskPath: tempFile,
			},
			wantErr: true,
		},
		{
			name: "valid SSH port - high value",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      2,
				RAMGb:    4,
				SSHPort:  65535,
				DiskPath: tempFile,
			},
			wantErr: false,
		},
		{
			name: "invalid monitor port - too low",
			config: &VMConfig{
				Arch:        "x86_64",
				CPU:         2,
				RAMGb:       4,
				SSHPort:     2222,
				MonitorPort: 1023,
				DiskPath:    tempFile,
			},
			wantErr: true,
		},
		{
			name: "valid monitor port - disabled",
			config: &VMConfig{
				Arch:        "x86_64",
				CPU:         2,
				RAMGb:       4,
				SSHPort:     2222,
				MonitorPort: 0,
				DiskPath:    tempFile,
			},
			wantErr: false,
		},
		{
			name: "valid monitor port - enabled",
			config: &VMConfig{
				Arch:        "x86_64",
				CPU:         2,
				RAMGb:       4,
				SSHPort:     2222,
				MonitorPort: 1024,
				DiskPath:    tempFile,
			},
			wantErr: false,
		},
		{
			name: "invalid disk path - empty",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      2,
				RAMGb:    4,
				SSHPort:  2222,
				DiskPath: "",
			},
			wantErr: true,
		},
		{
			name: "invalid disk path - not found",
			config: &VMConfig{
				Arch:     "x86_64",
				CPU:      2,
				RAMGb:    4,
				SSHPort:  2222,
				DiskPath: "/path/to/non/existent/disk.img",
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

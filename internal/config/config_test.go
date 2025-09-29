package config

import (
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
			name: "valid SSH port - high value",
			config: &VMConfig{
				Arch:    "x86_64",
				CPU:     2,
				RAMGb:   4,
				SSHPort: 65535,
			},
			wantErr: false,
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

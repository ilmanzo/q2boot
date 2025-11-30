package config

import (
	"fmt"
	"os"
)

// VMConfig holds the configuration settings for the VM
type VMConfig struct {
	Arch        string `json:"arch" mapstructure:"arch"`
	CPU         int    `json:"cpu" mapstructure:"cpu"`
	RAMGb       int    `json:"ram_gb" mapstructure:"ram_gb"`
	SSHPort     uint16 `json:"ssh_port" mapstructure:"ssh_port"`
	MonitorPort uint16 `json:"monitor_port" mapstructure:"monitor_port"`
	LogFile     string `json:"log_file" mapstructure:"log_file"`
	WriteMode   bool   `json:"write_mode" mapstructure:"write_mode"`
	Graphical   bool   `json:"graphical" mapstructure:"graphical"`
	Confirm     bool   `json:"confirm" mapstructure:"confirm"`
	DiskPath    string `json:"disk_path,omitempty" mapstructure:"disk_path"`
}

// DefaultConfig creates a default configuration
// Note: While Viper handles defaults via SetDefault(), this function
// is still useful for testing and programmatic config creation
func DefaultConfig() *VMConfig {
	return &VMConfig{
		Arch:        "x86_64",
		CPU:         2,
		RAMGb:       2,
		SSHPort:     2222,
		MonitorPort: 0, // Default to 0, meaning disabled
		LogFile:     "q2boot.log",
		WriteMode:   false,
		Graphical:   false,
		Confirm:     false,
	}
}

// Validate validates the configuration values
// This provides domain-specific validation logic that Viper doesn't handle
func (c *VMConfig) Validate() error {
	if c.CPU < 1 || c.CPU > 32 {
		return fmt.Errorf("CPU count must be between 1 and 32, got %d", c.CPU)
	}

	if c.RAMGb < 1 || c.RAMGb > 128 {
		return fmt.Errorf("RAM must be between 1 and 128 GB, got %d", c.RAMGb)
	}

	if c.SSHPort < 1024 {
		return fmt.Errorf("SSH port must be >= 1024, got %d", c.SSHPort)
	}

	if c.MonitorPort != 0 && c.MonitorPort < 1024 {
		return fmt.Errorf("monitor port must be >= 1024, got %d", c.MonitorPort)
	}

	validArchs := []string{"x86_64", "aarch64", "ppc64le", "s390x"}
	validArch := false
	for _, arch := range validArchs {
		if c.Arch == arch {
			validArch = true
			break
		}
	}
	if !validArch {
		return fmt.Errorf("invalid architecture '%s'. Valid options: %v", c.Arch, validArchs)
	}

	if c.DiskPath != "" {
		if _, err := os.Stat(c.DiskPath); os.IsNotExist(err) {
			return fmt.Errorf("disk image not found at '%s'", c.DiskPath)
		}
	}

	return nil
}

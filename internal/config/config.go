package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VMConfig holds the configuration settings for the VM
type VMConfig struct {
	Arch      string `json:"arch" mapstructure:"arch"`
	CPU       int    `json:"cpu" mapstructure:"cpu"`
	RAMGb     int    `json:"ram_gb" mapstructure:"ram_gb"`
	SSHPort   uint16 `json:"ssh_port" mapstructure:"ssh_port"`
	LogFile   string `json:"log_file" mapstructure:"log_file"`
	WriteMode bool   `json:"write_mode" mapstructure:"write_mode"`
	Graphical bool   `json:"graphical" mapstructure:"graphical"`
	Confirm   bool   `json:"confirm" mapstructure:"confirm"`
	DiskPath  string `json:"disk_path,omitempty" mapstructure:"disk_path"`
}

// DefaultConfig creates a default configuration
func DefaultConfig() *VMConfig {
	return &VMConfig{
		Arch:      "x86_64",
		CPU:       2,
		RAMGb:     2,
		SSHPort:   2222,
		LogFile:   "qboot.log",
		WriteMode: false,
		Graphical: false,
		Confirm:   false,
	}
}

// LoadConfig loads the configuration from the specified file
func LoadConfig(path string) (*VMConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	config := DefaultConfig()
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to the specified file
func SaveConfig(config *VMConfig, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// EnsureConfigExists ensures the config directory and file exist
// It creates them with default values if they don't
func EnsureConfigExists(configDir, configFile string) error {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		fmt.Printf("Creating config directory at '%s'\n", configDir)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("No config file found. Creating default config at '%s'\n", configFile)
		defaultConfig := DefaultConfig()
		if err := SaveConfig(defaultConfig, configFile); err != nil {
			return fmt.Errorf("failed to create default config file: %w", err)
		}
	}
	return nil
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() (string, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory not found
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "qboot")
	configFile := filepath.Join(configDir, "config.json")
	return configDir, configFile
}

// Validate validates the configuration values
func (c *VMConfig) Validate() error {
	if c.CPU < 1 || c.CPU > 32 {
		return fmt.Errorf("CPU count must be between 1 and 32, got %d", c.CPU)
	}

	if c.RAMGb < 1 || c.RAMGb > 128 {
		return fmt.Errorf("RAM must be between 1 and 128 GB, got %d", c.RAMGb)
	}

	if c.SSHPort < 1024 || c.SSHPort > 65535 {
		return fmt.Errorf("SSH port must be between 1024 and 65535, got %d", c.SSHPort)
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

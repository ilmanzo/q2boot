package config

import (
	"fmt"
	"os"
)

// Validation constants
const (
	MinCPU             = 1
	MaxCPU             = 32
	MinRAM             = 1
	MaxRAM             = 128
	MinPrivilegedPort  = 1024
	DefaultCPU         = 2
	DefaultRAMGb       = 2
	DefaultSSHPort     = 2222
	DefaultMonitorPort = 0 // 0 means disabled
	DefaultLogFile     = "q2boot.log"
)

// VMConfig holds the configuration settings for the VM
type VMConfig struct {
	Arch          string   `json:"arch" mapstructure:"arch"`
	CPU           int      `json:"cpu" mapstructure:"cpu"`
	RAMGb         int      `json:"ram_gb" mapstructure:"ram_gb"`
	SSHPort       uint16   `json:"ssh_port" mapstructure:"ssh_port"`
	MonitorPort   uint16   `json:"monitor_port" mapstructure:"monitor_port"`
	LogFile       string   `json:"log_file" mapstructure:"log_file"`
	SerialLogPath string   `json:"serial_log_path" mapstructure:"serial_log_path"`
	WriteMode     bool     `json:"write_mode" mapstructure:"write_mode"`
	Graphical     bool     `json:"graphical" mapstructure:"graphical"`
	Confirm       bool     `json:"confirm" mapstructure:"confirm"`
	DiskPath      string   `json:"disk_path,omitempty" mapstructure:"disk_path"`
	ExtraQemuArgs []string `json:"extra_qemu_args,omitempty" mapstructure:"extra_qemu_args"`
}

// DefaultConfig creates a default configuration
// Note: While Viper handles defaults via SetDefault(), this function
// is still useful for testing and programmatic config creation
func DefaultConfig() *VMConfig {
	return &VMConfig{
		CPU:           DefaultCPU,
		RAMGb:         DefaultRAMGb,
		SSHPort:       DefaultSSHPort,
		MonitorPort:   DefaultMonitorPort, // Default to 0, meaning disabled
		LogFile:       DefaultLogFile,
		SerialLogPath: "",
		WriteMode:     false,
		Graphical:     false,
		Confirm:       false,
	}
}

// Validate validates the configuration values
// This provides domain-specific validation logic that Viper doesn't handle
func (c *VMConfig) Validate() error {
	if c.CPU < MinCPU || c.CPU > MaxCPU {
		return fmt.Errorf("CPU count must be between %d and %d, got %d", MinCPU, MaxCPU, c.CPU)
	}

	if c.RAMGb < MinRAM || c.RAMGb > MaxRAM {
		return fmt.Errorf("RAM must be between %d and %d GB, got %d", MinRAM, MaxRAM, c.RAMGb)
	}

	if c.SSHPort < MinPrivilegedPort {
		return fmt.Errorf("SSH port must be >= %d, got %d", MinPrivilegedPort, c.SSHPort)
	}

	if c.MonitorPort != 0 && c.MonitorPort < MinPrivilegedPort {
		return fmt.Errorf("monitor port must be >= %d, got %d", MinPrivilegedPort, c.MonitorPort)
	}

	if c.DiskPath == "" {
		return fmt.Errorf("disk path is required (use -d or --disk)")
	}

	if _, err := os.Stat(c.DiskPath); os.IsNotExist(err) {
		return fmt.Errorf("disk image not found at '%s'", c.DiskPath)
	}

	return nil
}

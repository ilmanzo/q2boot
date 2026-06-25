package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	MinCPU             = 1
	MaxCPU             = 64
	MinRAM             = 1
	MaxRAM             = 128
	MinPrivilegedPort  = 1024
	DefaultCPU         = 2
	DefaultRAMGb       = 2
	DefaultSSHPort     = 2222
	DefaultMonitorPort = 0
)

// defaultLogFile returns $XDG_STATE_HOME/q2boot/q2boot.log, falling back to
// ~/.local/state/q2boot/q2boot.log. Falls back to "q2boot.log" only when the
// home directory is unresolvable (rare; e.g. no passwd entry).
func defaultLogFile() string {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "q2boot.log"
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, "q2boot", "q2boot.log")
}

type VMConfig struct {
	Arch          string   `json:"arch"`
	CPU           int      `json:"cpu"`
	RAMGb         int      `json:"ram_gb"`
	SSHPort       uint16   `json:"ssh_port"`
	MonitorPort   uint16   `json:"monitor_port"`
	LogFile       string   `json:"log_file"`
	WriteMode     bool     `json:"write_mode"`
	Graphical     bool     `json:"graphical"`
	Confirm       bool     `json:"confirm"`
	DiskPath      string   `json:"disk_path,omitempty"`
	ExtraQemuArgs []string `json:"extra_qemu_args,omitempty"`
}

func DefaultConfig() *VMConfig {
	return &VMConfig{
		CPU:         DefaultCPU,
		RAMGb:       DefaultRAMGb,
		SSHPort:     DefaultSSHPort,
		MonitorPort: DefaultMonitorPort,
		LogFile:     defaultLogFile(),
	}
}

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
		return fmt.Errorf("disk image path is required (pass it as the positional argument: q2boot <disk_image_path>)")
	}

	return nil
}

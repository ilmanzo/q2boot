package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ilmanzo/q2boot/internal/config"
	"github.com/ilmanzo/q2boot/internal/logger"
	"github.com/ilmanzo/q2boot/internal/vm"
)

var (
	// Version information - set by build flags
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"

	// Command line flags
	diskPath    string
	cpu         int
	ram         int
	arch        string
	sshPort     uint16
	monitorPort uint16
	logFile     string
	graphical   bool
	writeMode   bool
	confirm     bool

	// Configuration
	cfg *config.VMConfig
)

var rootCmd = &cobra.Command{
	Use:     "q2boot",
	Version: version,
	Short:   "A handy QEMU VM launcher",
	Long: `Q2Boot is a command-line tool that wraps QEMU to provide a streamlined
experience for launching virtual machines. It automatically configures common
settings like KVM acceleration, virtio drivers, and networking while allowing
customization through both configuration files and command-line options.`,
	RunE: runQ2Boot,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Q2Boot version", "version", version)
		logger.Info("Git commit", "commit", commit)
		logger.Info("Build time", "time", buildTime)
		logger.Info("Supported architectures", "archs", vm.SupportedArchitectures())

		// Show QEMU binary availability
		logger.Info("QEMU Binary Availability")
		availability := vm.CheckAvailableQEMUBinaries()
		for _, arch := range vm.SupportedArchitectures() {
			status := "❌ Not Available"
			if availability[arch] {
				status = "✅ Available"
			}
			binary, _ := vm.GetQEMUBinaryForArch(arch)
			logger.Info("Architecture", "arch", arch, "binary", binary, "status", status)
		}

		missing := vm.GetMissingQEMUBinaries()
		if len(missing) > 0 {
			logger.Info("To install missing QEMU binaries")
			for _, arch := range missing {
				binary, _ := vm.GetQEMUBinaryForArch(arch)
				logger.Info("Missing binary", "arch", arch, "binary", binary)
				instructions := vm.GetInstallationInstructions(binary)
				lines := strings.Split(instructions, "\n")
				for _, line := range lines {
					if line != "" {
						logger.Info("Instruction", "text", line)
					}
				}
			}
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add subcommands
	rootCmd.AddCommand(versionCmd)

	// Define command line flags
	rootCmd.PersistentFlags().StringVarP(&diskPath, "disk", "d", "", "Path to the disk image (required)")
	rootCmd.PersistentFlags().IntVarP(&cpu, "cpu", "c", 0, "Number of CPU cores (default: 2)")
	rootCmd.PersistentFlags().IntVarP(&ram, "ram", "r", 0, "Amount of RAM in GB (default: 2)")
	rootCmd.PersistentFlags().StringVarP(&arch, "arch", "a", "x86_64", "CPU architecture (x86_64, aarch64, ppc64le, s390x)")
	rootCmd.PersistentFlags().Uint16VarP(&sshPort, "ssh-port", "p", 0, "Host port for SSH forwarding (default: 2222)")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log-file", "l", "", "Path to the log file (default: q2boot.log)")
	rootCmd.PersistentFlags().BoolVarP(&graphical, "graphical", "g", false, "Enable graphical console (default: false)")
	rootCmd.PersistentFlags().BoolVarP(&writeMode, "write-mode", "w", false, "Enable write mode (changes are saved to disk) (default: false)")
	rootCmd.PersistentFlags().BoolVar(&confirm, "confirm", false, "Show command and wait for keypress before starting (default: false)")
	rootCmd.PersistentFlags().Uint16VarP(&monitorPort, "monitor-port", "m", 0, "Port for the QEMU monitor (telnet)")

	// Mark required flags only for root command, not subcommands
	rootCmd.MarkFlagRequired("disk")

	// Bind flags to viper
	viper.BindPFlag("disk", rootCmd.PersistentFlags().Lookup("disk"))
	viper.BindPFlag("cpu", rootCmd.PersistentFlags().Lookup("cpu"))
	viper.BindPFlag("ram", rootCmd.PersistentFlags().Lookup("ram"))
	viper.BindPFlag("arch", rootCmd.PersistentFlags().Lookup("arch"))
	viper.BindPFlag("ssh_port", rootCmd.PersistentFlags().Lookup("ssh-port"))
	viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("graphical", rootCmd.PersistentFlags().Lookup("graphical"))
	viper.BindPFlag("write_mode", rootCmd.PersistentFlags().Lookup("write-mode"))
	viper.BindPFlag("confirm", rootCmd.PersistentFlags().Lookup("confirm"))
	viper.BindPFlag("monitor_port", rootCmd.PersistentFlags().Lookup("monitor-port"))
}

// testConfigDir is used by tests to override the default config location.
var testConfigDir string

func initConfig() {
	var configDir, configFile string

	if testConfigDir != "" {
		configDir = testConfigDir
		configFile = filepath.Join(configDir, "config")
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir = filepath.Join(home, ".config", "q2boot")
		configFile = filepath.Join(configDir, "config")
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		logger.Error("Error creating config directory", "path", configDir, "error", err)
		return
	}

	// Configure viper
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("arch", "x86_64")
	viper.SetDefault("cpu", 2)
	viper.SetDefault("ram_gb", 2)
	viper.SetDefault("ssh_port", 2222)
	viper.SetDefault("log_file", "q2boot.log")
	viper.SetDefault("graphical", false)
	viper.SetDefault("write_mode", false)
	viper.SetDefault("confirm", false)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default
			logger.Info("No config file found. Creating default config", "path", configFile+".json")
			if err := viper.WriteConfigAs(configFile + ".json"); err != nil {
				logger.Error("Error creating config file", "path", configFile+".json", "error", err)
			}
		} else {
			logger.Error("Error reading config file", "error", err)
		}
	}

	// Create config struct
	cfg = &config.VMConfig{}
	if err := viper.Unmarshal(cfg); err != nil {
		logger.Error("Error unmarshaling config", "error", err)
		cfg = config.DefaultConfig()
	}
}

func runQ2Boot(cmd *cobra.Command, args []string) error {
	return runQ2BootE(cmd, args, cfg)
}

// runQ2BootE contains the core logic for running the VM, making it testable.
func runQ2BootE(cmd *cobra.Command, args []string, cfg *config.VMConfig) error {
	if diskPath != "" {
		cfg.DiskPath = diskPath
	}
	if cpu > 0 {
		cfg.CPU = cpu
	}
	if ram > 0 {
		cfg.RAMGb = ram
	}
	if arch != "" {
		cfg.Arch = arch
	}
	if sshPort > 0 {
		cfg.SSHPort = sshPort
	}
	if logFile != "" {
		cfg.LogFile = logFile
	}
	if cmd.Flags().Changed("graphical") {
		cfg.Graphical = graphical
	}
	if cmd.Flags().Changed("write-mode") {
		cfg.WriteMode = writeMode
	}
	if cmd.Flags().Changed("confirm") {
		cfg.Confirm = confirm
	}
	if cmd.Flags().Changed("monitor-port") {
		cfg.MonitorPort = monitorPort
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Validate port availability
	if err := vm.ValidatePortsAvailable(cfg.SSHPort, cfg.MonitorPort); err != nil {
		return err
	}

	// Validate architecture separately to avoid import cycle
	if !vm.IsArchSupported(cfg.Arch) {
		return fmt.Errorf("invalid architecture '%s'. Valid options: %v", cfg.Arch, vm.SupportedArchitectures())
	}

	// Create VM based on architecture
	virtualMachine, err := vm.CreateVM(cfg.Arch)
	if err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	// Validate QEMU binary availability using the VM's specific binary
	if err := vm.ValidateQEMUBinary(virtualMachine.QEMUBinary()); err != nil {
		return fmt.Errorf("QEMU validation failed: %w", err)
	}

	// Configure the VM
	virtualMachine.Configure(cfg)

	// Run the VM
	logger.Info("Starting VM", "arch", cfg.Arch)
	return virtualMachine.Run()
}

func main() {
	// Initialize logger
	_ = logger.Initialize(logger.InfoLevel, "text")

	if err := rootCmd.Execute(); err != nil {
		logger.Error("Fatal error", "error", err)
		os.Exit(1)
	}
}

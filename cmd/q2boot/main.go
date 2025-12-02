package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ilmanzo/q2boot/internal/config"
	"github.com/ilmanzo/q2boot/internal/detector"
	"github.com/ilmanzo/q2boot/internal/logger"
	"github.com/ilmanzo/q2boot/internal/vm"
)

// Configuration directory constants
const (
	ConfigDirPermissions = 0700 // Owner read/write/execute only for security
	ConfigDirName        = ".config"
	AppConfigDirName     = "q2boot"
	ConfigFileName       = "config"
	ConfigFileFormat     = "json"
)

// Flags holds all command-line flag values
type Flags struct {
	CPU         int
	RAM         int
	Arch        string
	SSHPort     uint16
	MonitorPort uint16
	LogFile     string
	Graphical   bool
	WriteMode   bool
	Confirm     bool
}

var (
	// Version information - set by build flags
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"

	// Command line flags
	flags = &Flags{}

	// Configuration
	cfg *config.VMConfig
)

var rootCmd = &cobra.Command{
	Use:     "q2boot [flags] <disk_image_path>",
	Version: version,
	Short:   "A handy QEMU VM launcher",
	Long: `Q2Boot is a command-line tool that wraps QEMU to provide a streamlined
experience for launching virtual machines. It automatically configures common
settings like KVM acceleration, virtio drivers, and networking while allowing
customization through both configuration files and command-line options.`,
	Args: cobra.ExactArgs(1), // Expect exactly one argument: the disk image path
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
	setupFlags()
}

// setupFlags defines and binds all command-line flags.
// It's a separate function to allow re-initialization during testing.
func setupFlags() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().IntVarP(&flags.CPU, "cpu", "c", 0, "Number of CPU cores (default: 2)")
	rootCmd.PersistentFlags().IntVarP(&flags.RAM, "ram", "r", 0, "Amount of RAM in GB (default: 2)")
	rootCmd.PersistentFlags().StringVarP(&flags.Arch, "arch", "a", "", "CPU architecture (x86_64, aarch64, ppc64le, s390x). Auto-detected from disk image if not specified")
	rootCmd.PersistentFlags().Uint16VarP(&flags.SSHPort, "ssh-port", "p", 0, "Host port for SSH forwarding (default: 2222)")
	rootCmd.PersistentFlags().StringVarP(&flags.LogFile, "log-file", "l", "", "Path to the log file (default: q2boot.log)")
	rootCmd.PersistentFlags().BoolVarP(&flags.Graphical, "graphical", "g", false, "Enable graphical console (default: false)")
	rootCmd.PersistentFlags().BoolVarP(&flags.WriteMode, "write-mode", "w", false, "Enable write mode (changes are saved to disk) (default: false)")
	rootCmd.PersistentFlags().BoolVar(&flags.Confirm, "confirm", false, "Show command and wait for keypress before starting (default: false)")
	rootCmd.PersistentFlags().Uint16VarP(&flags.MonitorPort, "monitor-port", "m", 0, "Port for the QEMU monitor (telnet)")

	// Bind flags to viper
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
		configFile = filepath.Join(configDir, ConfigFileName)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir = filepath.Join(home, ConfigDirName, AppConfigDirName)
		configFile = filepath.Join(configDir, ConfigFileName)
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, ConfigDirPermissions); err != nil {
		logger.Error("Error creating config directory", "path", configDir, "error", err)
		return
	}

	// Configure viper
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileFormat)
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("arch", "")
	viper.SetDefault("cpu", config.DefaultCPU)
	viper.SetDefault("ram_gb", config.DefaultRAMGb)
	viper.SetDefault("ssh_port", config.DefaultSSHPort)
	viper.SetDefault("log_file", config.DefaultLogFile)
	viper.SetDefault("graphical", false)
	viper.SetDefault("write_mode", false)
	viper.SetDefault("confirm", false)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default
			logger.Info("No config file found. Creating default config", "path", configFile+"."+ConfigFileFormat)
			if err := viper.WriteConfigAs(configFile + "." + ConfigFileFormat); err != nil {
				logger.Error("Error creating config file", "path", configFile+"."+ConfigFileFormat, "error", err)
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

// applyFlagOverrides applies command-line flag overrides to the configuration.
// It checks which flags were explicitly set and overwrites the corresponding config values.
func applyFlagOverrides(cmd *cobra.Command, f *Flags, cfg *config.VMConfig, diskPath string) {
	cfg.DiskPath = diskPath
	if f.CPU > 0 {
		cfg.CPU = f.CPU
	}
	if f.RAM > 0 {
		cfg.RAMGb = f.RAM
	}
	if f.Arch != "" {
		cfg.Arch = f.Arch
	}
	if f.SSHPort > 0 {
		cfg.SSHPort = f.SSHPort
	}
	if f.LogFile != "" {
		cfg.LogFile = f.LogFile
	}
	if cmd.Flags().Changed("graphical") {
		cfg.Graphical = f.Graphical
	}
	if cmd.Flags().Changed("write-mode") {
		cfg.WriteMode = f.WriteMode
	}
	if cmd.Flags().Changed("confirm") {
		cfg.Confirm = f.Confirm
	}
	if cmd.Flags().Changed("monitor-port") {
		cfg.MonitorPort = f.MonitorPort
	}
}

// detectArchitecture automatically detects the architecture from the disk image.
// This is called when no explicit architecture was provided via flag.
func detectArchitecture(diskPath string) (string, error) {
	logger.Info("Attempting to detect architecture from disk image", "disk", diskPath)
	arch, err := detector.DetectArchitecture(diskPath)
	if err != nil {
		return "", err
	}
	logger.Info("Successfully detected architecture", "arch", arch)
	return arch, nil
}

// runQ2BootE contains the core logic for running the VM, making it testable.
func runQ2BootE(cmd *cobra.Command, args []string, cfg *config.VMConfig) error {
	// Apply flag overrides to configuration
	applyFlagOverrides(cmd, flags, cfg, args[0])

	// If architecture was not explicitly provided via the command-line flag,
	// attempt automatic detection. This correctly ignores any 'arch' from the config file.
	if !cmd.Flags().Changed("arch") {
		detectedArch, err := detectArchitecture(cfg.DiskPath)
		if err != nil {
			return fmt.Errorf("architecture not specified and automatic detection failed: %w", err)
		}
		cfg.Arch = detectedArch
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
		return err
	}

	// Configure the VM
	virtualMachine.Configure(cfg)

	// Validate the configured VM (this will include QEMU binary checks)
	if err := virtualMachine.Validate(); err != nil {
		return fmt.Errorf("VM validation failed: %w", err)
	}

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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ilmanzo/q2boot/internal/config"
	"github.com/ilmanzo/q2boot/internal/detector"
	"github.com/ilmanzo/q2boot/internal/downloader"
	"github.com/ilmanzo/q2boot/internal/vm"
)

// Flags holds all command-line flag values
type Flags struct {
	CPU           int
	RAM           int
	Arch          string
	SSHPort       uint16
	MonitorPort   uint16
	LogFile       string
	Graphical     bool
	WriteMode     bool
	Confirm       bool
	ExtraQemuArgs []string
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
customization through both configuration files and command-line options.

For advanced use cases, you can pass custom arguments directly to QEMU using the -qemu-extra flag.`,
	Args: cobra.ExactArgs(1), // Expect exactly one argument: the disk image path
	RunE: runQ2Boot,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Q2Boot version: %s\n", version)
		fmt.Printf("Git commit:     %s\n", commit)
		fmt.Printf("Build time:     %s\n", buildTime)
		fmt.Printf("Supported architectures: %v\n", vm.SupportedArchitectures())

		fmt.Println("\nQEMU Binary Availability:")
		availability := vm.CheckAvailableQEMUBinaries()
		for _, arch := range vm.SupportedArchitectures() {
			status := "❌ Not Available"
			if availability[arch] {
				status = "✅ Available"
			}
			binary, _ := vm.GetQEMUBinaryForArch(arch)
			fmt.Printf("  %-8s %-22s %s\n", arch, binary, status)
		}

		missing := vm.GetMissingQEMUBinaries()
		if len(missing) > 0 {
			fmt.Println("\nTo install missing QEMU binaries:")
			for _, arch := range missing {
				binary, _ := vm.GetQEMUBinaryForArch(arch)
				fmt.Printf("- %s (%s):\n", arch, binary)
				fmt.Println(vm.GetInstallationInstructions(binary))
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
	rootCmd.AddCommand(NewCheckCmd())

	rootCmd.PersistentFlags().IntVarP(&flags.CPU, "cpu", "c", 0, "Number of CPU cores (default: 2)")
	rootCmd.PersistentFlags().IntVarP(&flags.RAM, "ram", "r", 0, "Amount of RAM in GB (default: 2)")
	rootCmd.PersistentFlags().StringVarP(&flags.Arch, "arch", "a", "", "CPU architecture (x86_64, aarch64, ppc64le, s390x). Auto-detected from disk image if not specified")
	rootCmd.PersistentFlags().Uint16VarP(&flags.SSHPort, "ssh-port", "p", 0, "Host port for SSH forwarding (default: 2222)")
	rootCmd.PersistentFlags().StringVarP(&flags.LogFile, "log-file", "l", "", "Path to the log file (default: q2boot.log)")
	rootCmd.PersistentFlags().BoolVarP(&flags.Graphical, "graphical", "g", false, "Enable graphical console (default: false)")
	rootCmd.PersistentFlags().BoolVarP(&flags.WriteMode, "write-mode", "w", false, "Enable write mode (changes are saved to disk) (default: false)")
	rootCmd.PersistentFlags().BoolVar(&flags.Confirm, "confirm", false, "Show command and wait for keypress before starting (default: false)")
	rootCmd.PersistentFlags().Uint16VarP(&flags.MonitorPort, "monitor-port", "m", 0, "Port for the QEMU monitor (telnet)")
	rootCmd.PersistentFlags().StringSliceVarP(&flags.ExtraQemuArgs, "qemu-extra", "e", []string{}, "Extra arguments to pass to QEMU (can be specified multiple times)")

	// Bind flags to viper. The key MUST match the VMConfig mapstructure tag,
	// otherwise viper.Unmarshal silently ignores the flag value.
	viper.BindPFlag("cpu", rootCmd.PersistentFlags().Lookup("cpu"))
	viper.BindPFlag("ram_gb", rootCmd.PersistentFlags().Lookup("ram"))
	viper.BindPFlag("arch", rootCmd.PersistentFlags().Lookup("arch"))
	viper.BindPFlag("ssh_port", rootCmd.PersistentFlags().Lookup("ssh-port"))
	viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("graphical", rootCmd.PersistentFlags().Lookup("graphical"))
	viper.BindPFlag("write_mode", rootCmd.PersistentFlags().Lookup("write-mode"))
	viper.BindPFlag("confirm", rootCmd.PersistentFlags().Lookup("confirm"))
	viper.BindPFlag("monitor_port", rootCmd.PersistentFlags().Lookup("monitor-port"))
	viper.BindPFlag("extra_qemu_args", rootCmd.PersistentFlags().Lookup("qemu-extra"))
}

// testConfigDir is used by tests to override the default config location.
var testConfigDir string

func initConfig() {
	var configDir string

	if testConfigDir != "" {
		configDir = testConfigDir
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		configDir = filepath.Join(home, ".config", "q2boot")
	}
	configFile := filepath.Join(configDir, "config.json")

	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Printf("Error creating config directory %s: %v\n", configDir, err)
		return
	}

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)

	viper.SetDefault("arch", "")
	viper.SetDefault("cpu", config.DefaultCPU)
	viper.SetDefault("ram_gb", config.DefaultRAMGb)
	viper.SetDefault("ssh_port", config.DefaultSSHPort)
	viper.SetDefault("log_file", config.DefaultLogFile)
	viper.SetDefault("graphical", false)
	viper.SetDefault("write_mode", false)
	viper.SetDefault("confirm", false)
	viper.SetDefault("extra_qemu_args", []string{})

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("No config file found. Creating default at %s\n", configFile)
			if err := viper.WriteConfigAs(configFile); err != nil {
				fmt.Printf("Error creating config file %s: %v\n", configFile, err)
			}
		} else {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	cfg = &config.VMConfig{}
	if err := viper.Unmarshal(cfg); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
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
	if len(f.ExtraQemuArgs) > 0 {
		cfg.ExtraQemuArgs = f.ExtraQemuArgs
	}
}

// detectArchitecture automatically detects the architecture from the disk image.
// This is called when no explicit architecture was provided via flag.
func detectArchitecture(diskPath string) (string, error) {
	fmt.Printf("Attempting to detect architecture from disk image: %s\n", diskPath)
	arch, err := detector.DetectArchitecture(diskPath)
	if err != nil {
		return "", err
	}
	fmt.Printf("Successfully detected architecture: %s\n", arch)
	return arch, nil
}

// runQ2BootE contains the core logic for running the VM, making it testable.
func runQ2BootE(cmd *cobra.Command, args []string, cfg *config.VMConfig) error {
	diskPath := args[0]

	// Handle remote images
	if downloader.IsRemote(diskPath) {
		localPath, cleanup, err := downloader.Download(diskPath)
		if err != nil {
			return fmt.Errorf("failed to download image: %w", err)
		}
		defer cleanup()
		diskPath = localPath
	}

	// Apply flag overrides to configuration
	applyFlagOverrides(cmd, flags, cfg, diskPath)

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

	// Create VM (errors on unsupported architecture).
	virtualMachine, err := vm.CreateVM(cfg.Arch)
	if err != nil {
		return err
	}

	virtualMachine.Configure(cfg)

	// Includes QEMU-binary, port-availability, and disk-path checks.
	if err := virtualMachine.Validate(); err != nil {
		return fmt.Errorf("VM validation failed: %w", err)
	}

	fmt.Printf("Starting VM (arch=%s)\n", cfg.Arch)
	return virtualMachine.Run()
}

func main() {
	// Initialize logger

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

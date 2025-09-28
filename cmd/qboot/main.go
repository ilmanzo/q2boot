package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ilmanzo/qboot/internal/config"
	"github.com/ilmanzo/qboot/internal/vm"
)

var (
	// Version information - set by build flags
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"

	// Command line flags
	diskPath  string
	cpu       int
	ram       int
	arch      string
	sshPort   uint16
	logFile   string
	graphical bool
	writeMode bool
	confirm   bool

	// Configuration
	cfg *config.VMConfig
)

var rootCmd = &cobra.Command{
	Use:     "qboot",
	Version: version,
	Short:   "A handy QEMU VM launcher",
	Long: `QBoot is a command-line tool that wraps QEMU to provide a streamlined
experience for launching virtual machines. It automatically configures common
settings like KVM acceleration, virtio drivers, and networking while allowing
customization through both configuration files and command-line options.`,
	RunE: runQBoot,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("QBoot %s\n", version)
		fmt.Printf("Git commit: %s\n", commit)
		fmt.Printf("Build time: %s\n", buildTime)
		fmt.Printf("Supported architectures: %v\n", vm.SupportedArchitectures())
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
	rootCmd.PersistentFlags().StringVarP(&arch, "arch", "a", "", "CPU architecture (x86_64, aarch64, ppc64le, s390x) (default: x86_64)")
	rootCmd.PersistentFlags().Uint16VarP(&sshPort, "ssh-port", "p", 0, "Host port for SSH forwarding (default: 2222)")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log-file", "l", "", "Path to the log file (default: qboot.log)")
	rootCmd.PersistentFlags().BoolVarP(&graphical, "graphical", "g", false, "Enable graphical console (default: false)")
	rootCmd.PersistentFlags().BoolVarP(&writeMode, "write-mode", "w", false, "Enable write mode (changes are saved to disk) (default: false)")
	rootCmd.PersistentFlags().BoolVar(&confirm, "confirm", false, "Show command and wait for keypress before starting (default: false)")

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
}

func initConfig() {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	// Set up configuration paths
	configDir := filepath.Join(home, ".config", "qboot")
	configFile := filepath.Join(configDir, "config")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
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
	viper.SetDefault("log_file", "qboot.log")
	viper.SetDefault("graphical", false)
	viper.SetDefault("write_mode", false)
	viper.SetDefault("confirm", false)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default
			fmt.Printf("No config file found. Creating default config at '%s.json'\n", configFile)
			if err := viper.WriteConfigAs(configFile + ".json"); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config file: %v\n", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}

	// Create config struct
	cfg = &config.VMConfig{}
	if err := viper.Unmarshal(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling config: %v\n", err)
		cfg = config.DefaultConfig()
	}
}

func runQBoot(cmd *cobra.Command, args []string) error {
	// Override config with command line flags if provided
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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Validate required parameters
	if cfg.DiskPath == "" {
		return fmt.Errorf("disk path is required (use -d or --disk)")
	}

	// Create VM based on architecture
	virtualMachine, err := vm.CreateVM(cfg.Arch)
	if err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	// Configure the VM
	virtualMachine.Configure(cfg)

	// Run the VM
	fmt.Printf("🚀 Starting %s VM...\n", cfg.Arch)
	return virtualMachine.Run()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

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

var (
	// Version information - set by build flags
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"

	// cfg is populated lazily in runQ2BootE from viper (file + flag merge).
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
	Args: cobra.ExactArgs(1),
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
	setupFlags()
}

// setupFlags declares the persistent CLI flags and binds them to viper.
// Flag values are read back via viper.Unmarshal(cfg) in runQ2BootE — there
// is no parallel Flags struct, and viper's pflag precedence does the merge
// (CLI > env > config file > default) automatically.
//
// IMPORTANT: BindPFlag keys MUST match VMConfig mapstructure tags or
// Unmarshal silently drops the flag value. See [[viper-bind-key-mismatch]].
func setupFlags() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(NewCheckCmd())

	pf := rootCmd.PersistentFlags()
	pf.IntP("cpu", "c", 0, "Number of CPU cores (default: 2)")
	pf.IntP("ram", "r", 0, "Amount of RAM in GB (default: 2)")
	pf.StringP("arch", "a", "", "CPU architecture (x86_64, aarch64, ppc64le, s390x). Auto-detected from disk image if not specified")
	pf.Uint16P("ssh-port", "p", 0, "Host port for SSH forwarding (default: 2222)")
	pf.StringP("log-file", "l", "", "Path to the log file (default: q2boot.log)")
	pf.BoolP("graphical", "g", false, "Enable graphical console (default: false)")
	pf.BoolP("write-mode", "w", false, "Enable write mode (changes are saved to disk) (default: false)")
	pf.Bool("confirm", false, "Show command and wait for keypress before starting (default: false)")
	pf.Uint16P("monitor-port", "m", 0, "Port for the QEMU monitor (telnet)")
	pf.StringSliceP("qemu-extra", "e", []string{}, "Extra arguments to pass to QEMU (can be specified multiple times)")

	viper.BindPFlag("cpu", pf.Lookup("cpu"))
	viper.BindPFlag("ram_gb", pf.Lookup("ram"))
	viper.BindPFlag("arch", pf.Lookup("arch"))
	viper.BindPFlag("ssh_port", pf.Lookup("ssh-port"))
	viper.BindPFlag("log_file", pf.Lookup("log-file"))
	viper.BindPFlag("graphical", pf.Lookup("graphical"))
	viper.BindPFlag("write_mode", pf.Lookup("write-mode"))
	viper.BindPFlag("confirm", pf.Lookup("confirm"))
	viper.BindPFlag("monitor_port", pf.Lookup("monitor-port"))
	viper.BindPFlag("extra_qemu_args", pf.Lookup("qemu-extra"))
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
	// Re-unmarshal so that values bound after initConfig (e.g. flags parsed
	// during Execute) are reflected in cfg. cmd argument retained for parity
	// with cobra signatures.
	_ = cmd
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	cfg.DiskPath = args[0]
	remote := downloader.IsRemote(cfg.DiskPath)

	// Reject the combination of a remote image and --write-mode: the local copy
	// lives in a temp file we delete on exit, so any persistence intent is a lie.
	if remote && cfg.WriteMode {
		return fmt.Errorf("--write-mode cannot be used with a remote image (%s): changes would be discarded with the downloaded temp file. Download the image first, then pass the local path", cfg.DiskPath)
	}

	if remote {
		localPath, cleanup, err := downloader.Download(cfg.DiskPath)
		if err != nil {
			return fmt.Errorf("failed to download image: %w", err)
		}
		defer cleanup()
		cfg.DiskPath = localPath
	} else if _, err := os.Stat(cfg.DiskPath); err != nil {
		return fmt.Errorf("disk image not accessible: %w", err)
	}

	// If architecture was not specified anywhere (flag, env, or config file),
	// auto-detect from the disk image.
	if cfg.Arch == "" {
		detectedArch, err := detectArchitecture(cfg.DiskPath)
		if err != nil {
			return fmt.Errorf("architecture not specified and automatic detection failed: %w", err)
		}
		cfg.Arch = detectedArch
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	virtualMachine, err := vm.CreateVM(cfg.Arch)
	if err != nil {
		return err
	}
	virtualMachine.Configure(cfg)

	if err := virtualMachine.Validate(); err != nil {
		return fmt.Errorf("VM validation failed: %w", err)
	}

	fmt.Printf("Starting VM (arch=%s)\n", cfg.Arch)
	return virtualMachine.Run()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

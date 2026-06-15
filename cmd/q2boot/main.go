package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

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

	// cfg holds the merged config: defaults < config file < CLI flags.
	cfg *config.VMConfig

	// flagCfg is the staging struct that CLI flags write into. It is merged
	// into cfg in runQ2BootE based on which flags were Changed.
	flagCfg = config.DefaultConfig()
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

// Each flag writes into flagCfg; precedence (defaults < file < CLI) is
// applied later by applyChangedFlags.
func setupFlags() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(NewCheckCmd())

	pf := rootCmd.PersistentFlags()
	pf.IntVarP(&flagCfg.CPU, "cpu", "c", flagCfg.CPU, "Number of CPU cores")
	pf.IntVarP(&flagCfg.RAMGb, "ram", "r", flagCfg.RAMGb, "Amount of RAM in GB")
	pf.StringVarP(&flagCfg.Arch, "arch", "a", flagCfg.Arch, "CPU architecture (x86_64, aarch64, ppc64le, s390x). Auto-detected from disk image if not specified")
	pf.Uint16VarP(&flagCfg.SSHPort, "ssh-port", "p", flagCfg.SSHPort, "Host port for SSH forwarding")
	pf.StringVarP(&flagCfg.LogFile, "log-file", "l", flagCfg.LogFile, "Path to the log file")
	pf.BoolVarP(&flagCfg.Graphical, "graphical", "g", flagCfg.Graphical, "Enable graphical console")
	pf.BoolVarP(&flagCfg.WriteMode, "write-mode", "w", flagCfg.WriteMode, "Enable write mode (changes are saved to disk)")
	pf.BoolVar(&flagCfg.Confirm, "confirm", flagCfg.Confirm, "Show command and wait for keypress before starting")
	pf.Uint16VarP(&flagCfg.MonitorPort, "monitor-port", "m", flagCfg.MonitorPort, "Port for the QEMU monitor (telnet)")
	pf.StringSliceVarP(&flagCfg.ExtraQemuArgs, "qemu-extra", "e", flagCfg.ExtraQemuArgs, "Extra arguments to pass to QEMU (can be specified multiple times)")
}

var testConfigDir string

func configPath() (string, string) {
	if testConfigDir != "" {
		return testConfigDir, filepath.Join(testConfigDir, "config.json")
	}
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	dir := filepath.Join(home, ".config", "q2boot")
	return dir, filepath.Join(dir, "config.json")
}

// initConfig loads ~/.config/q2boot/config.json, writing DefaultConfig out
// on first run so the user has something to edit.
func initConfig() {
	dir, path := configPath()
	if err := os.MkdirAll(dir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory %s: %v\n", dir, err)
		cfg = config.DefaultConfig()
		return
	}
	loaded, err := config.Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("No config file found. Creating default at %s\n", path)
			cfg = config.DefaultConfig()
			if werr := config.Save(path, cfg); werr != nil {
				fmt.Fprintf(os.Stderr, "Error creating config file %s: %v\n", path, werr)
			}
			return
		}
		fmt.Fprintf(os.Stderr, "Error reading config file (using defaults): %v\n", err)
		cfg = config.DefaultConfig()
		return
	}
	cfg = loaded
}

// applyChangedFlags is the CLI-wins step of the precedence chain.
func applyChangedFlags(cmd *cobra.Command, cfg *config.VMConfig) {
	pf := cmd.Flags()
	if pf.Changed("cpu") {
		cfg.CPU = flagCfg.CPU
	}
	if pf.Changed("ram") {
		cfg.RAMGb = flagCfg.RAMGb
	}
	if pf.Changed("arch") {
		cfg.Arch = flagCfg.Arch
	}
	if pf.Changed("ssh-port") {
		cfg.SSHPort = flagCfg.SSHPort
	}
	if pf.Changed("log-file") {
		cfg.LogFile = flagCfg.LogFile
	}
	if pf.Changed("graphical") {
		cfg.Graphical = flagCfg.Graphical
	}
	if pf.Changed("write-mode") {
		cfg.WriteMode = flagCfg.WriteMode
	}
	if pf.Changed("confirm") {
		cfg.Confirm = flagCfg.Confirm
	}
	if pf.Changed("monitor-port") {
		cfg.MonitorPort = flagCfg.MonitorPort
	}
	if pf.Changed("qemu-extra") {
		cfg.ExtraQemuArgs = flagCfg.ExtraQemuArgs
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

func runQ2BootE(cmd *cobra.Command, args []string, cfg *config.VMConfig) error {
	applyChangedFlags(cmd, cfg)
	cfg.DiskPath = args[0]
	remote := downloader.IsRemote(cfg.DiskPath)

	// Remote + --write-mode is incoherent: the local copy lives in a temp
	// file we delete on exit, so any "persisted" change vanishes.
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

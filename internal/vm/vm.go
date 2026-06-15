package vm

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/ilmanzo/q2boot/internal/config"
)

const (
	LocalhostAddress   = "127.0.0.1"
	TCPNetworkProtocol = "tcp"
	MonitorProtocol    = "telnet"
	AudioDeviceID      = "snd0"
	AudioDeviceType    = "none"
	SnapshotArgument   = "-snapshot"
	SerialConsoleStdio = "stdio"
)

type VM interface {
	QEMUBinary() string
	GetArchArgs() []string
	GetDiskArgs() []string
	GetNetworkArgs() []string
	GetGraphicalArgs() []string
	GetNonGraphicalDisplayArgs() []string
	Configure(cfg *config.VMConfig)
	Validate() error
	Run() error
}

// BaseVM embeds VMConfig so concrete arch types read v.CPU/v.RAMGb/etc.
// directly. `self` holds the embedding concrete type so the promoted
// Validate/Run can dispatch into the arch-specific methods; constructors
// MUST set it or BaseVM.Validate/Run nil-derefs.
type BaseVM struct {
	*config.VMConfig
	self         VM
	FirmwarePath string
}

func NewBaseVM() *BaseVM {
	return &BaseVM{VMConfig: config.DefaultConfig()}
}

func (v *BaseVM) Configure(cfg *config.VMConfig) {
	v.VMConfig = cfg
}

func ValidateQEMUBinary(binary string) error {
	_, err := exec.LookPath(binary)
	if err != nil {
		instructions := GetInstallationInstructions(binary)
		return fmt.Errorf("QEMU binary '%s' not found in PATH. %s\nError: %v", binary, instructions, err)
	}
	return nil
}

func IsPortAvailable(port uint16) bool {
	listener, err := net.Listen(TCPNetworkProtocol, fmt.Sprintf("%s:%d", LocalhostAddress, port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func ValidatePortsAvailable(sshPort, monitorPort uint16) error {
	if !IsPortAvailable(sshPort) {
		return fmt.Errorf("SSH port %d is already in use. Please choose a different port using --ssh-port", sshPort)
	}

	if monitorPort > 0 && !IsPortAvailable(monitorPort) {
		return fmt.Errorf("monitor port %d is already in use. Please choose a different port using --monitor-port", monitorPort)
	}

	return nil
}

func (v *BaseVM) Validate() error {
	if err := ValidateQEMUBinary(v.self.QEMUBinary()); err != nil {
		return err
	}
	if err := ValidatePortsAvailable(v.SSHPort, v.MonitorPort); err != nil {
		return err
	}
	if v.DiskPath == "" {
		return fmt.Errorf("disk image path is not set")
	}
	return nil
}

func (v *BaseVM) buildArgs(vm VM, extraArgs []string) []string {
	var args []string
	args = append(args, vm.GetArchArgs()...)
	args = append(args, "-smp", fmt.Sprintf("%d", v.CPU))
	args = append(args, "-m", fmt.Sprintf("%dG", v.RAMGb))
	args = append(args, vm.GetDiskArgs()...)
	args = append(args, vm.GetNetworkArgs()...)
	if extraArgs != nil {
		args = append(args, extraArgs...)
	}
	if len(v.ExtraQemuArgs) > 0 {
		args = append(args, v.ExtraQemuArgs...)
	}
	args = append(args, "-audiodev", fmt.Sprintf("%s,id=%s", AudioDeviceType, AudioDeviceID))

	if v.Graphical {
		graphicalArgs := vm.GetGraphicalArgs()
		args = append(args, graphicalArgs...)
		// Graphical-via-`-nographic` (s390x) takes over stdio; suppress the
		// default monitor to avoid a conflict.
		if slices.Contains(graphicalArgs, "-nographic") {
			args = append(args, "-monitor", "none")
		}
	} else {
		args = append(args, vm.GetNonGraphicalDisplayArgs()...)
		if !v.WriteMode {
			args = append(args, SnapshotArgument)
		}
	}

	if v.MonitorPort > 0 {
		if !slices.Contains(args, "mon:stdio") {
			args = append(args, "-monitor", fmt.Sprintf("%s:%s:%d,server,nowait", MonitorProtocol, LocalhostAddress, v.MonitorPort))
		}
	} else if !v.Graphical {
		// Console mode without an explicit monitor would otherwise grab stdio.
		if !slices.Contains(args, "-monitor") && !slices.Contains(args, "mon:stdio") {
			args = append(args, "-monitor", "none")
		}
	}

	if v.LogFile != "" {
		args = append(args, "-chardev", fmt.Sprintf("stdio,mux=on,id=char0,logfile=%s,signal=off", v.LogFile))
		args = append(args, "-mon", "chardev=char0,mode=readline")
		args = append(args, "-serial", "chardev:char0")
	}

	return args
}

func (v *BaseVM) Run() error {
	args := v.buildArgs(v.self, nil)
	return RunVM(v.self.QEMUBinary(), args, v.Confirm)
}

func RunVM(binary string, args []string, confirm bool) error {
	fmt.Println("Starting QEMU with the following command:")
	fmt.Printf("  %s %s\n", binary, strings.Join(args, " "))

	if confirm {
		fmt.Print("Press Enter to continue...")
		var input string
		fmt.Scanln(&input)
	}

	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("QEMU exited with status %d", exitError.ExitCode())
		}
		return fmt.Errorf("failed to start QEMU: %w", err)
	}
	return nil
}

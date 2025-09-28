package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ilmanzo/qboot/internal/config"
)

func TestNewBaseVM(t *testing.T) {
	vm := NewBaseVM()

	if vm.CPU != 2 {
		t.Errorf("Expected CPU to be 2, got %d", vm.CPU)
	}

	if vm.RAM != 2 {
		t.Errorf("Expected RAM to be 2, got %d", vm.RAM)
	}

	if vm.SSHPort != 2222 {
		t.Errorf("Expected SSHPort to be 2222, got %d", vm.SSHPort)
	}

	if vm.LogFile != "qboot.log" {
		t.Errorf("Expected LogFile to be qboot.log, got %s", vm.LogFile)
	}

	if vm.Graphical != false {
		t.Errorf("Expected Graphical to be false, got %t", vm.Graphical)
	}

	if vm.NoSnapshot != false {
		t.Errorf("Expected NoSnapshot to be false, got %t", vm.NoSnapshot)
	}

	if vm.Confirm != false {
		t.Errorf("Expected Confirm to be false, got %t", vm.Confirm)
	}
}

func TestBaseConfigure(t *testing.T) {
	vm := NewBaseVM()
	cfg := &config.VMConfig{
		Arch:      "aarch64",
		CPU:       4,
		RAMGb:     8,
		SSHPort:   2223,
		LogFile:   "test.log",
		WriteMode: true,
		Graphical: true,
		Confirm:   true,
		DiskPath:  "/tmp/test.img",
	}

	vm.Configure(cfg)

	if vm.CPU != cfg.CPU {
		t.Errorf("Expected CPU to be %d, got %d", cfg.CPU, vm.CPU)
	}

	if vm.RAM != cfg.RAMGb {
		t.Errorf("Expected RAM to be %d, got %d", cfg.RAMGb, vm.RAM)
	}

	if vm.SSHPort != cfg.SSHPort {
		t.Errorf("Expected SSHPort to be %d, got %d", cfg.SSHPort, vm.SSHPort)
	}

	if vm.LogFile != cfg.LogFile {
		t.Errorf("Expected LogFile to be %s, got %s", cfg.LogFile, vm.LogFile)
	}

	if vm.Graphical != cfg.Graphical {
		t.Errorf("Expected Graphical to be %t, got %t", cfg.Graphical, vm.Graphical)
	}

	if vm.NoSnapshot != cfg.WriteMode {
		t.Errorf("Expected NoSnapshot to be %t, got %t", cfg.WriteMode, vm.NoSnapshot)
	}

	if vm.Confirm != cfg.Confirm {
		t.Errorf("Expected Confirm to be %t, got %t", cfg.Confirm, vm.Confirm)
	}

	if vm.DiskPath != cfg.DiskPath {
		t.Errorf("Expected DiskPath to be %s, got %s", cfg.DiskPath, vm.DiskPath)
	}
}

func TestValidateDiskPath(t *testing.T) {
	tests := []struct {
		name     string
		diskPath string
		wantErr  bool
	}{
		{
			name:     "empty path",
			diskPath: "",
			wantErr:  true,
		},
		{
			name:     "non-existent path",
			diskPath: "/non/existent/path.img",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiskPath(tt.diskPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDiskPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDiskPathWithRealFile(t *testing.T) {
	// Create a temporary file for testing
	tempDir, err := os.MkdirTemp("", "qboot-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test.img")
	err = os.WriteFile(tempFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test with existing file
	err = ValidateDiskPath(tempFile)
	if err != nil {
		t.Errorf("ValidateDiskPath() with existing file should not return error, got: %v", err)
	}
}

func TestValidateVMConfig(t *testing.T) {
	tests := []struct {
		name    string
		vm      *BaseVM
		wantErr bool
	}{
		{
			name:    "valid config",
			vm:      NewBaseVM(),
			wantErr: false,
		},
		{
			name: "invalid CPU - too low",
			vm: &BaseVM{
				CPU:     0,
				RAM:     4,
				SSHPort: 2222,
			},
			wantErr: true,
		},
		{
			name: "invalid CPU - too high",
			vm: &BaseVM{
				CPU:     33,
				RAM:     4,
				SSHPort: 2222,
			},
			wantErr: true,
		},
		{
			name: "invalid RAM - too low",
			vm: &BaseVM{
				CPU:     2,
				RAM:     0,
				SSHPort: 2222,
			},
			wantErr: true,
		},
		{
			name: "invalid RAM - too high",
			vm: &BaseVM{
				CPU:     2,
				RAM:     129,
				SSHPort: 2222,
			},
			wantErr: true,
		},
		{
			name: "invalid SSH port - too low",
			vm: &BaseVM{
				CPU:     2,
				RAM:     4,
				SSHPort: 1023,
			},
			wantErr: true,
		},
		{
			name: "valid SSH port - maximum",
			vm: &BaseVM{
				CPU:     2,
				RAM:     4,
				SSHPort: 65535,
			},
			wantErr: false,
		},
		{
			name: "invalid SSH port - below minimum",
			vm: &BaseVM{
				CPU:     2,
				RAM:     4,
				SSHPort: 1023,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVMConfig(tt.vm)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVMConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateVM(t *testing.T) {
	tests := []struct {
		name     string
		arch     string
		wantType string
		wantErr  bool
	}{
		{
			name:     "x86_64 VM",
			arch:     "x86_64",
			wantType: "*vm.X86_64VM",
			wantErr:  false,
		},
		{
			name:     "aarch64 VM",
			arch:     "aarch64",
			wantType: "*vm.AARCH64VM",
			wantErr:  false,
		},
		{
			name:     "ppc64le VM",
			arch:     "ppc64le",
			wantType: "*vm.PPC64LEVM",
			wantErr:  false,
		},
		{
			name:     "s390x VM",
			arch:     "s390x",
			wantType: "*vm.S390XVM",
			wantErr:  false,
		},
		{
			name:    "unsupported architecture",
			arch:    "unsupported",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, err := CreateVM(tt.arch)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && vm == nil {
				t.Errorf("CreateVM() returned nil VM for supported architecture %s", tt.arch)
			}
		})
	}
}

func TestSupportedArchitectures(t *testing.T) {
	archs := SupportedArchitectures()
	expected := []string{"x86_64", "aarch64", "ppc64le", "s390x"}

	if len(archs) != len(expected) {
		t.Errorf("Expected %d architectures, got %d", len(expected), len(archs))
	}

	for i, arch := range expected {
		if i >= len(archs) || archs[i] != arch {
			t.Errorf("Expected architecture %s at position %d, got %s", arch, i, archs[i])
		}
	}
}

func TestIsArchSupported(t *testing.T) {
	tests := []struct {
		name      string
		arch      string
		supported bool
	}{
		{"x86_64 supported", "x86_64", true},
		{"aarch64 supported", "aarch64", true},
		{"ppc64le supported", "ppc64le", true},
		{"s390x supported", "s390x", true},
		{"unsupported arch", "unsupported", false},
		{"empty arch", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsArchSupported(tt.arch); got != tt.supported {
				t.Errorf("IsArchSupported() = %v, want %v", got, tt.supported)
			}
		})
	}
}

// Test specific VM implementations
func TestX86_64VM(t *testing.T) {
	vm := NewX86_64VM()

	if vm.QEMUBinary() != "qemu-system-x86_64" {
		t.Errorf("Expected QEMU binary to be qemu-system-x86_64, got %s", vm.QEMUBinary())
	}

	archArgs := vm.GetArchArgs()
	expectedArgs := []string{"-M", "q35", "-enable-kvm", "-cpu", "host"}
	if len(archArgs) != len(expectedArgs) {
		t.Errorf("Expected %d arch args, got %d", len(expectedArgs), len(archArgs))
	}
}

func TestAARCH64VM(t *testing.T) {
	vm := NewAARCH64VM()

	if vm.QEMUBinary() != "qemu-system-aarch64" {
		t.Errorf("Expected QEMU binary to be qemu-system-aarch64, got %s", vm.QEMUBinary())
	}

	archArgs := vm.GetArchArgs()
	if len(archArgs) == 0 {
		t.Error("Expected non-empty arch args for aarch64")
	}
}

func TestPPC64LEVM(t *testing.T) {
	vm := NewPPC64LEVM()

	if vm.QEMUBinary() != "qemu-system-ppc64" {
		t.Errorf("Expected QEMU binary to be qemu-system-ppc64, got %s", vm.QEMUBinary())
	}

	archArgs := vm.GetArchArgs()
	expectedArgs := []string{"-M", "pseries", "-cpu", "POWER9"}
	if len(archArgs) != len(expectedArgs) {
		t.Errorf("Expected %d arch args, got %d", len(expectedArgs), len(archArgs))
	}
}

func TestS390XVM(t *testing.T) {
	vm := NewS390XVM()

	if vm.QEMUBinary() != "qemu-system-s390x" {
		t.Errorf("Expected QEMU binary to be qemu-system-s390x, got %s", vm.QEMUBinary())
	}

	archArgs := vm.GetArchArgs()
	expectedArgs := []string{"-machine", "s390-ccw-virtio", "-cpu", "max"}
	if len(archArgs) != len(expectedArgs) {
		t.Errorf("Expected %d arch args, got %d", len(expectedArgs), len(archArgs))
	}
}

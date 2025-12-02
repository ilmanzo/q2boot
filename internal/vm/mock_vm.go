package vm

import "github.com/ilmanzo/q2boot/internal/config"

// MockVM is a mock implementation of the VM interface for testing.
type MockVM struct {
	*BaseVM
	RunFunc              func() error
	ValidateFunc         func() error
	GetGraphicalArgsFunc func() []string
}

// NewMockVM creates a new MockVM instance.
func NewMockVM() *MockVM {
	return &MockVM{
		BaseVM: NewBaseVM(),
	}
}

// QEMUBinary is a mock implementation of the QEMUBinary method.
func (m *MockVM) QEMUBinary() string {
	return "qemu-mock"
}

// GetArchArgs is a mock implementation of the GetArchArgs method.
func (m *MockVM) GetArchArgs() []string {
	return []string{"-machine", "mock"}
}

// GetDiskArgs is a mock implementation of the GetDiskArgs method.
func (m *MockVM) GetDiskArgs() []string {
	return []string{"-drive", "file=mock.img"}
}

// GetNetworkArgs is a mock implementation of the GetNetworkArgs method.
func (m *MockVM) GetNetworkArgs() []string {
	return []string{"-netdev", "user,id=net0"}
}

// GetGraphicalArgs is a mock implementation of the GetGraphicalArgs method.
func (m *MockVM) GetGraphicalArgs() []string {
	if m.GetGraphicalArgsFunc != nil {
		return m.GetGraphicalArgsFunc()
	}
	return []string{"-display", "mock"}
}

// GetNonGraphicalDisplayArgs is a mock implementation of the GetNonGraphicalDisplayArgs method.
func (m *MockVM) GetNonGraphicalDisplayArgs() []string {
	return []string{"-display", "mock-headless"}
}

// Run is a mock implementation of the Run method.
func (m *MockVM) Run() error {
	if m.RunFunc != nil {
		return m.RunFunc()
	}
	return nil
}

// Validate is a mock implementation of the Validate method.
func (m *MockVM) Validate() error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc()
	}
	return m.BaseVM.Validate(m)
}

// Configure is a mock implementation of the Configure method.
func (m *MockVM) Configure(cfg *config.VMConfig) {
	m.BaseVM.Configure(cfg)
}

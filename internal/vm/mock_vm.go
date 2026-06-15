package vm

import "github.com/ilmanzo/q2boot/internal/config"

type MockVM struct {
	*BaseVM
	RunFunc              func() error
	ValidateFunc         func() error
	GetGraphicalArgsFunc func() []string
}

func NewMockVM() *MockVM {
	m := &MockVM{BaseVM: NewBaseVM()}
	m.self = m
	return m
}

func (m *MockVM) QEMUBinary() string       { return "qemu-mock" }
func (m *MockVM) GetArchArgs() []string    { return []string{"-machine", "mock"} }
func (m *MockVM) GetDiskArgs() []string    { return []string{"-drive", "file=mock.img"} }
func (m *MockVM) GetNetworkArgs() []string { return []string{"-netdev", "user,id=net0"} }

func (m *MockVM) GetGraphicalArgs() []string {
	if m.GetGraphicalArgsFunc != nil {
		return m.GetGraphicalArgsFunc()
	}
	return []string{"-display", "mock"}
}

func (m *MockVM) GetNonGraphicalDisplayArgs() []string {
	return []string{"-display", "mock-headless"}
}

func (m *MockVM) Run() error {
	if m.RunFunc != nil {
		return m.RunFunc()
	}
	return nil
}

func (m *MockVM) Validate() error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc()
	}
	return m.BaseVM.Validate()
}

func (m *MockVM) Configure(cfg *config.VMConfig) {
	m.BaseVM.Configure(cfg)
}

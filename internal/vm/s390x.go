package vm

import (
	"fmt"
)

// S390XVM implements VM for s390x architecture
type S390XVM struct {
	*BaseVM
}

// NewS390XVM creates a new S390XVM instance
func NewS390XVM() *S390XVM {
	return &S390XVM{
		BaseVM: NewBaseVM(),
	}
}

// QEMUBinary returns the QEMU binary name for s390x
func (vm *S390XVM) QEMUBinary() string {
	return "qemu-system-s390x"
}

// GetArchArgs returns architecture-specific arguments for s390x
func (vm *S390XVM) GetArchArgs() []string {
	return []string{
		"-machine", "s390-ccw-virtio",
		"-cpu", "max",
	}
}

// GetDiskArgs returns s390x-specific disk arguments
func (vm *S390XVM) GetDiskArgs() []string {
	return []string{
		"-drive",
		fmt.Sprintf("file=%s,id=disk1,if=none,cache=none,aio=native,discard=unmap", vm.DiskPath),
		"-device",
		fmt.Sprintf("virtio-blk-ccw,drive=disk1,id=dr1,bootindex=1,num-queues=%d", vm.CPU),
	}
}

// GetNetworkArgs returns s390x-specific network arguments
func (vm *S390XVM) GetNetworkArgs() []string {
	return []string{
		"-netdev",
		fmt.Sprintf("user,id=net1,hostfwd=tcp::%d-:22", vm.SSHPort),
		"-device",
		"virtio-net-ccw,netdev=net1,mq=on",
	}
}

// GetGraphicalArgs returns s390x-specific graphical mode arguments
// On s390x, this provides an interactive session in the terminal,
// multiplexing the serial console and the QEMU monitor.
func (vm *S390XVM) GetGraphicalArgs() []string {
	return []string{
		"-nographic",
		"-serial", "stdio",
	}
}

// GetNonGraphicalDisplayArgs returns display arguments for non-graphical mode on s390x
// s390x uses nographic mode with serial console instead of curses
func (vm *S390XVM) GetNonGraphicalDisplayArgs() []string {
	return []string{"-nographic"}
}

// BuildArgs builds the complete argument list for s390x
func (vm *S390XVM) BuildArgs() []string {
	return vm.buildArgs(vm)
}

// Run executes the VM
func (vm *S390XVM) Run() error {
	return vm.RunVM(vm)
}

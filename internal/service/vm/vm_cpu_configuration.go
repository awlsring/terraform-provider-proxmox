package vm

import "github.com/awlsring/proxmox-go/proxmox"

type VirtualMachineCpu struct {
	Architecture string
	Cores        int
	Sockets      int
	EmulatedType *string
	CpuUnits     *int64
}

func DetermineCPUConfiguration(sum proxmox.VirtualMachineConfigurationSummary) VirtualMachineCpu {
	cpu := VirtualMachineCpu{}

	if sum.HasArch() {
		cpu.Architecture = string(*sum.Arch)
	} else {
		cpu.Architecture = "x86_64"
	}

	if sum.HasCores() {
		cpu.Cores = int(*sum.Cores)
	} else {
		cpu.Cores = 1
	}

	if sum.HasSockets() {
		cpu.Sockets = int(*sum.Sockets)
	} else {
		cpu.Sockets = 1
	}

	if sum.HasCpu() {
		cpu.EmulatedType = sum.Cpu
	}

	if sum.HasCpuunits() {
		units := int64(*sum.Cpuunits)
		cpu.CpuUnits = &units
	}

	return cpu
}

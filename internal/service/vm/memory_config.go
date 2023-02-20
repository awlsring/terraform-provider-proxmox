package vm

import "github.com/awlsring/proxmox-go/proxmox"

type VirtualMachineMemory struct {
	Dedicated int64
	Shared    *int64
	Floating  *int64
}

func DetermineMemoryConfiguration(sum proxmox.VirtualMachineConfigurationSummary) VirtualMachineMemory {
	mem := VirtualMachineMemory{}

	if sum.HasMemory() {
		mem.Dedicated = int64(*sum.Memory)
	}

	if sum.HasShares() {
		shared := int64(*sum.Shares)
		mem.Shared = &shared
	}

	if sum.HasBallon() {
		floating := int64(*sum.Ballon)
		mem.Floating = &floating
	}

	return mem
}

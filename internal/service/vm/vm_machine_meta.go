package vm

import "github.com/awlsring/proxmox-go/proxmox"

func DetermineOsType(sum proxmox.VirtualMachineConfigurationSummary) *proxmox.VirtualMachineOperatingSystem {
	if sum.HasOstype() {
		return sum.Ostype
	}

	d := proxmox.VIRTUALMACHINEOPERATINGSYSTEM_OTHER

	return &d
}

func DetermineMachineType(sum proxmox.VirtualMachineConfigurationSummary) *string {
	if sum.HasMachine() {
		return sum.Machine
	}

	return nil
}

func DetermineBios(b *proxmox.VirtualMachineBios) proxmox.VirtualMachineBios {
	if b == nil {
		return proxmox.VIRTUALMACHINEBIOS_SEABIOS
	}
	return *b
}

func DetermineKeyboardLayout(k *proxmox.VirtualMachineKeyboard) *proxmox.VirtualMachineKeyboard {
	if k == nil {
		return nil
	}
	return k
}

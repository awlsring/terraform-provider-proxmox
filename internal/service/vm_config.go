package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
)

type VirtualMachine struct {
	Node              string
	VmId              int
	Tags              []string
	Name              *string
	Description       *string
	Agent             *vm.VirtualMachineAgent
	Bios              proxmox.VirtualMachineBios
	CPU               vm.VirtualMachineCpu
	Disks             []vm.VirtualMachineDisk
	NetworkInterfaces []vm.VirtualMachineNetworkInterface
	Memory            vm.VirtualMachineMemory
	CloudInit         *vm.VirtualMachineCloudInit
	OsType            *proxmox.VirtualMachineOperatingSystem
	MachineType       *string
	KVMArguments      *string
	StartOnBoot       bool
	KeyboardLayout    *proxmox.VirtualMachineKeyboard
}

func (c *Proxmox) DescribeVirtualMachine(ctx context.Context, node string, vmid int) (*VirtualMachine, error) {
	configSummary, err := c.GetVirtualMachineConfiguration(ctx, node, vmid)
	if err != nil {
		return nil, err
	}

	config := &VirtualMachine{
		Node:           node,
		VmId:           vmid,
		Description:    configSummary.Description,
		Agent:          vm.DetermineAgentConfig(configSummary.Agent),
		Bios:           vm.DetermineBios(configSummary.Bios),
		CPU:            vm.DetermineCPUConfiguration(*configSummary),
		Memory:         vm.DetermineMemoryConfiguration(*configSummary),
		CloudInit:      vm.DetermineCloudInitConfiguration(*configSummary),
		OsType:         vm.DetermineOsType(*configSummary),
		MachineType:    vm.DetermineMachineType(*configSummary),
		KVMArguments:   configSummary.Args,
		KeyboardLayout: vm.DetermineKeyboardLayout(configSummary.Keyboard),
		Tags:           StringSemiColonPtrListToSlice(configSummary.Tags),
		Name:           configSummary.Name,
		StartOnBoot:    BooleanIntegerConversion(configSummary.Onboot),
	}

	diskConfig, err := vm.DetermineDiskConfiguration(configSummary)
	if err != nil {
		return nil, err
	}
	config.Disks = diskConfig

	networkConfig, err := vm.DetermineNetworkDevicesFromConfig(configSummary)
	if err != nil {
		return nil, err
	}
	config.NetworkInterfaces = networkConfig

	return config, nil
}

package types

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	qt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualMachineDataSourceModel struct {
	ID                types.Int64                               `tfsdk:"id"`
	Node              types.String                              `tfsdk:"node"`
	Name              types.String                              `tfsdk:"name"`
	Description       types.String                              `tfsdk:"description"`
	Tags              types.Set                                 `tfsdk:"tags"`
	Agent             *qt.VirtualMachineAgentModel              `tfsdk:"agent"`
	BIOS              types.String                              `tfsdk:"bios"`
	CPU               qt.VirtualMachineCpuModel                 `tfsdk:"cpu"`
	Disks             qt.VirtualMachineDiskSetValue             `tfsdk:"disks"`
	PCIDevices        qt.VirtualMachinePCIDeviceSetValue        `tfsdk:"pci_devices"`
	NetworkInterfaces qt.VirtualMachineNetworkInterfaceSetValue `tfsdk:"network_interfaces"`
	Memory            qt.VirtualMachineMemoryModel              `tfsdk:"memory"`
	MachineType       types.String                              `tfsdk:"machine_type"`
	KVMArguments      types.String                              `tfsdk:"kvm_arguments"`
	KeyboardLayout    types.String                              `tfsdk:"keyboard_layout"`
	CloudInit         *qt.VirtualMachineCloudInitModel          `tfsdk:"cloud_init"`
	Type              types.String                              `tfsdk:"type"`
	ResourcePool      types.String                              `tfsdk:"resource_pool"`
	StartOnNodeBoot   types.Bool                                `tfsdk:"start_on_node_boot"`
}

func VMToModel(ctx context.Context, v *service.VirtualMachine) *VirtualMachineDataSourceModel {
	m := VirtualMachineDataSourceModel{
		ID:                types.Int64Value(int64(v.VmId)),
		Node:              types.StringValue(v.Node),
		Tags:              utils.UnpackSetType(v.Tags),
		BIOS:              types.StringValue(string(v.Bios)),
		CPU:               VMCPUToModel(&v.CPU),
		Memory:            VMMemoryToModel(&v.Memory),
		Disks:             qt.VirtualMachineDiskToSetValue(ctx, v.Disks),
		NetworkInterfaces: qt.VirtualMachineNetworkInterfaceToSetValue(ctx, v.NetworkInterfaces),
		PCIDevices:        qt.VirtualMachinePCIDeviceToSetValue(ctx, v.PCIDevices),
		CloudInit:         qt.CloudInitToModel(ctx, v.CloudInit),
		StartOnNodeBoot:   types.BoolValue(v.StartOnBoot),
	}

	if v.Description != nil {
		m.Description = types.StringValue(*v.Description)
	}

	if v.Name != nil {
		m.Name = types.StringValue(*v.Name)
	}

	if v.OsType != nil {
		m.Type = types.StringValue(string(*v.OsType))
	}

	if v.MachineType != nil {
		m.MachineType = types.StringValue(string(*v.MachineType))
	}

	if v.KeyboardLayout != nil {
		kl := string(*v.KeyboardLayout)
		m.KeyboardLayout = types.StringValue(kl)
	}

	if v.Agent != nil {
		a := VMAgentToModel(v.Agent)
		m.Agent = &a
	}

	return &m
}

func VMAgentToModel(agent *vm.VirtualMachineAgent) qt.VirtualMachineAgentModel {
	m := qt.VirtualMachineAgentModel{
		Enabled:   types.BoolValue(agent.Enabled),
		UseFSTrim: types.BoolValue(agent.FsTrim),
	}
	if agent.Type != nil {
		m.Type = types.StringValue(string(*agent.Type))
	}
	return m
}

func VMCPUToModel(cpu *vm.VirtualMachineCpu) qt.VirtualMachineCpuModel {
	m := qt.VirtualMachineCpuModel{
		Architecture: types.StringValue(string(cpu.Architecture)),
		Cores:        types.Int64Value(int64(cpu.Cores)),
		Sockets:      types.Int64Value(int64(cpu.Sockets)),
	}
	if cpu.EmulatedType != nil {
		m.EmulatedType = types.StringValue(string(*cpu.EmulatedType))
	}
	if cpu.CpuUnits != nil {
		m.CPUUnits = types.Int64Value(int64(*cpu.CpuUnits))
	}

	return m
}

func VMMemoryToModel(memory *vm.VirtualMachineMemory) qt.VirtualMachineMemoryModel {
	m := qt.VirtualMachineMemoryModel{
		Dedicated: types.Int64Value(int64(memory.Dedicated)),
	}

	if memory.Floating != nil {
		m.Floating = types.Int64Value(int64(*memory.Floating))
	}

	if memory.Shared != nil {
		m.Shared = types.Int64Value(int64(*memory.Shared))
	}

	return m
}

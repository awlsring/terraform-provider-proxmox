package types

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	qt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type VirtualMachineTerraformTimeouts struct {
	Create     types.Int64 `tfsdk:"create"`
	Delete     types.Int64 `tfsdk:"delete"`
	Stop       types.Int64 `tfsdk:"stop"`
	Start      types.Int64 `tfsdk:"start"`
	Reboot     types.Int64 `tfsdk:"reboot"`
	Shutdown   types.Int64 `tfsdk:"shutdown"`
	Clone      types.Int64 `tfsdk:"clone"`
	Configure  types.Int64 `tfsdk:"configure"`
	ResizeDisk types.Int64 `tfsdk:"resize_disk"`
}

type VirtualMachineCloneModel struct {
	Storage   types.String `tfsdk:"storage"`
	Source    types.Int64  `tfsdk:"source"`
	FullClone types.Bool   `tfsdk:"full_clone"`
}

type VirtualMachineIsoModel struct {
	Storage *types.String `tfsdk:"storage"`
	Image   *types.String `tfsdk:"image"`
}

type T struct {
	VirtualMachineDataSourceModel
	Timeouts *VirtualMachineTerraformTimeouts `tfsdk:"timeouts"`
}

type VirtualMachineResourceModel struct {
	ID                        types.Int64                               `tfsdk:"id"`
	Node                      types.String                              `tfsdk:"node"`
	Name                      types.String                              `tfsdk:"name"`
	Description               types.String                              `tfsdk:"description"`
	Tags                      types.Set                                 `tfsdk:"tags"`
	Clone                     *VirtualMachineCloneModel                 `tfsdk:"clone"`
	ISO                       *VirtualMachineIsoModel                   `tfsdk:"iso"`
	Agent                     *qt.VirtualMachineAgentModel              `tfsdk:"agent"`
	BIOS                      types.String                              `tfsdk:"bios"`
	CPU                       qt.VirtualMachineCpuModel                 `tfsdk:"cpu"`
	Disks                     qt.VirtualMachineDiskSetValue             `tfsdk:"disks"`
	ComputedDisks             qt.VirtualMachineDiskSetValue             `tfsdk:"computed_disks"`
	PCIDevices                qt.VirtualMachinePCIDeviceSetValue        `tfsdk:"pci_devices"`
	ComputedPCIDevices        qt.VirtualMachinePCIDeviceSetValue        `tfsdk:"computed_pci_devices"`
	NetworkInterfaces         qt.VirtualMachineNetworkInterfaceSetValue `tfsdk:"network_interfaces"`
	ComputedNetworkInterfaces qt.VirtualMachineNetworkInterfaceSetValue `tfsdk:"computed_network_interfaces"`
	Memory                    qt.VirtualMachineMemoryModel              `tfsdk:"memory"`
	MachineType               types.String                              `tfsdk:"machine_type"`
	KVMArguments              types.String                              `tfsdk:"kvm_arguments"`
	KeyboardLayout            types.String                              `tfsdk:"keyboard_layout"`
	CloudInit                 *qt.VirtualMachineCloudInitModel          `tfsdk:"cloud_init"`
	Type                      types.String                              `tfsdk:"type"`
	ResourcePool              types.String                              `tfsdk:"resource_pool"`
	StartOnCreate             types.Bool                                `tfsdk:"start_on_create"`
	StartOnNodeBoot           types.Bool                                `tfsdk:"start_on_node_boot"`
	Timeouts                  *VirtualMachineTerraformTimeouts          `tfsdk:"timeouts"`
}

func VMToResourceModel(ctx context.Context, v *service.VirtualMachine, state *VirtualMachineResourceModel) *VirtualMachineResourceModel {
	definedDisks, computedDisks := sortComputedAndDefinedDisks(ctx, v.Disks, &state.Disks)
	definedNics, computedNics := sortComputedAndDefinedNics(ctx, v.NetworkInterfaces, &state.NetworkInterfaces)
	definedPCI, computedPCI := sortComputedAndDefinedPCIDevices(ctx, v.PCIDevices, &state.PCIDevices)

	base := VMToModel(ctx, v)
	m := &VirtualMachineResourceModel{
		ID:                        base.ID,
		Node:                      base.Node,
		Name:                      base.Name,
		Description:               base.Description,
		Tags:                      base.Tags,
		Agent:                     base.Agent,
		BIOS:                      base.BIOS,
		CPU:                       base.CPU,
		Disks:                     qt.VirtualMachineDiskToSetValue(ctx, definedDisks),
		ComputedDisks:             qt.VirtualMachineDiskToSetValue(ctx, computedDisks),
		PCIDevices:                qt.VirtualMachinePCIDeviceToSetValue(ctx, definedPCI),
		ComputedPCIDevices:        qt.VirtualMachinePCIDeviceToSetValue(ctx, computedPCI),
		NetworkInterfaces:         qt.VirtualMachineNetworkInterfaceToSetValue(ctx, definedNics),
		ComputedNetworkInterfaces: qt.VirtualMachineNetworkInterfaceToSetValue(ctx, computedNics),
		Memory:                    base.Memory,
		MachineType:               base.MachineType,
		KVMArguments:              base.KVMArguments,
		KeyboardLayout:            base.KeyboardLayout,
		CloudInit:                 base.CloudInit,
		Type:                      base.Type,
		ResourcePool:              base.ResourcePool,
		StartOnNodeBoot:           base.StartOnNodeBoot,
	}

	// carry over statemetadata
	m.Clone = state.Clone
	m.ISO = state.ISO
	m.Timeouts = state.Timeouts
	m.StartOnCreate = state.StartOnCreate

	return m
}

func sortComputedAndDefinedDisks(ctx context.Context, disks []vm.VirtualMachineDisk, state *qt.VirtualMachineDiskSetValue) ([]vm.VirtualMachineDisk, []vm.VirtualMachineDisk) {
	definedDisksIface := []string{}
	for _, disk := range state.Disks {
		iface := fmt.Sprintf("%s%d", disk.InterfaceType.ValueString(), disk.Position.ValueInt64())
		definedDisksIface = append(definedDisksIface, iface)
	}
	tflog.Debug(ctx, fmt.Sprintf("Defined disks %v", definedDisksIface))

	generatedDisks := []vm.VirtualMachineDisk{}
	definedDisks := []vm.VirtualMachineDisk{}
	for _, disk := range disks {
		iface := fmt.Sprintf("%s%d", disk.InterfaceType, disk.Position)
		tflog.Debug(ctx, fmt.Sprintf("On iface %v", iface))
		if utils.ListContains(definedDisksIface, iface) {
			tflog.Debug(ctx, fmt.Sprintf("iface %v is defined", iface))
			definedDisks = append(definedDisks, disk)
		} else {
			tflog.Debug(ctx, fmt.Sprintf("iface %v is generated", iface))
			generatedDisks = append(generatedDisks, disk)
		}
	}

	return definedDisks, generatedDisks
}

func sortComputedAndDefinedNics(ctx context.Context, nics []vm.VirtualMachineNetworkInterface, state *qt.VirtualMachineNetworkInterfaceSetValue) ([]vm.VirtualMachineNetworkInterface, []vm.VirtualMachineNetworkInterface) {
	definedNicsIface := []string{}
	for _, nic := range state.Nics {
		iface := fmt.Sprintf("%s%d", nic.Model.ValueString(), nic.Position.ValueInt64())
		definedNicsIface = append(definedNicsIface, iface)

	}
	tflog.Debug(ctx, fmt.Sprintf("Defined nics %v", definedNicsIface))

	generatedNics := []vm.VirtualMachineNetworkInterface{}
	definedNics := []vm.VirtualMachineNetworkInterface{}
	for _, nic := range nics {
		iface := fmt.Sprintf("%s%d", nic.Model, nic.Position)
		tflog.Debug(ctx, fmt.Sprintf("On iface %v", iface))
		if utils.ListContains(definedNicsIface, iface) {
			tflog.Debug(ctx, fmt.Sprintf("iface %v is defined", iface))
			definedNics = append(definedNics, nic)
		} else {
			tflog.Debug(ctx, fmt.Sprintf("iface %v is generated", iface))
			generatedNics = append(generatedNics, nic)
		}
	}

	return definedNics, generatedNics
}

func sortComputedAndDefinedPCIDevices(ctx context.Context, devices []vm.VirtualMachinePCIDevice, state *qt.VirtualMachinePCIDeviceSetValue) ([]vm.VirtualMachinePCIDevice, []vm.VirtualMachinePCIDevice) {
	definedDevicesIface := []string{}
	for _, device := range state.PCIDevices {
		iface := fmt.Sprintf("%s-%s", device.Name.ValueString(), device.ID.ValueString())
		definedDevicesIface = append(definedDevicesIface, iface)
	}
	tflog.Debug(ctx, fmt.Sprintf("Defined pci devices %v", definedDevicesIface))

	generatedDevices := []vm.VirtualMachinePCIDevice{}
	definedDevices := []vm.VirtualMachinePCIDevice{}
	for _, device := range devices {
		iface := fmt.Sprintf("%s-%s", device.Name, device.ID)
		if utils.ListContains(definedDevicesIface, iface) {
			definedDevices = append(definedDevices, device)
		} else {
			generatedDevices = append(generatedDevices, device)
		}
	}

	return definedDevices, generatedDevices
}

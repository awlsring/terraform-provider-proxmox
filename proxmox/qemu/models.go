package qemu

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	t "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type VirtualMachineResourceModel struct {
	ID                types.Int64                               `tfsdk:"id"`
	Node              types.String                              `tfsdk:"node"`
	Name              types.String                              `tfsdk:"name"`
	Description       types.String                              `tfsdk:"description"`
	Tags              types.List                                `tfsdk:"tags"`
	Clone             *VirtualMachineCloneOptions               `tfsdk:"clone"`
	ISO               *VirtualMachineIsoOptions                 `tfsdk:"iso"`
	Agent             *VirtualMachineAgentOptions               `tfsdk:"agent"`
	BIOS              types.String                              `tfsdk:"bios"`
	CPU               VirtualMachineCpuOptions                  `tfsdk:"cpu"`
	Disks             t.VirtualMachineDiskSetValue              `tfsdk:"disks"`
	ComputedDisks     t.VirtualMachineDiskSetValue              `tfsdk:"computed_disks"`
	PCIDevices        []VirtualMachinePciDeviceOptions          `tfsdk:"pci_devices"`
	NetworkInterfaces t.VirtualMachineNetworkInterfaceListValue `tfsdk:"network_interfaces"`
	Memory            VirtualMachineMemoryOptions               `tfsdk:"memory"`
	MachineType       types.String                              `tfsdk:"machine_type"`
	KVMArguments      types.String                              `tfsdk:"kvm_arguments"`
	KeyboardLayout    types.String                              `tfsdk:"keyboard_layout"`
	CloudInit         *VirtualMachineCloudInitOptions           `tfsdk:"cloud_init"`
	Type              types.String                              `tfsdk:"type"`
	ResourcePool      types.String                              `tfsdk:"resource_pool"`
	StartOnCreate     types.Bool                                `tfsdk:"start_on_create"`
	StartOnNodeBoot   types.Bool                                `tfsdk:"start_on_node_boot"`
	Timeouts          *VirtualMachineTerraformTimeouts          `tfsdk:"timeouts"`
}

func VMToModel(ctx context.Context, v *service.VirtualMachine, state *VirtualMachineResourceModel) VirtualMachineResourceModel {
	definedDisksIface := []string{}
	for _, disk := range state.Disks.Disks {
		iface := fmt.Sprintf("%s%d", disk.InterfaceType.ValueString(), disk.Position.ValueInt64())
		definedDisksIface = append(definedDisksIface, iface)

	}
	tflog.Debug(ctx, fmt.Sprintf("Defined disks %v", definedDisksIface))

	generatedDisks := []vm.VirtualMachineDisk{}
	definedDisks := []vm.VirtualMachineDisk{}
	for _, disk := range v.Disks {
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

	m := VirtualMachineResourceModel{
		ID:                types.Int64Value(int64(v.VmId)),
		Node:              types.StringValue(v.Node),
		Tags:              utils.UnpackListType(v.Tags),
		BIOS:              types.StringValue(string(v.Bios)),
		CPU:               VMCPUToModel(&v.CPU),
		Memory:            VMMemoryToModel(&v.Memory),
		Disks:             t.VirtualMachineDiskToSetValue(ctx, definedDisks),
		ComputedDisks:     t.VirtualMachineDiskToSetValue(ctx, generatedDisks),
		NetworkInterfaces: t.VirtualMachineNetworkInterfaceToListValue(ctx, v.NetworkInterfaces),
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

	return m
}

type VirtualMachineCloneOptions struct {
	Storage   types.String `tfsdk:"storage"`
	Source    types.Int64  `tfsdk:"source"`
	FullClone types.Bool   `tfsdk:"full_clone"`
}

type VirtualMachineIsoOptions struct {
	Storage *types.String `tfsdk:"storage"`
	Image   *types.String `tfsdk:"image"`
}

type VirtualMachineAgentOptions struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	UseFSTrim types.Bool   `tfsdk:"use_fstrim"`
	Type      types.String `tfsdk:"type"`
}

func VMAgentToModel(agent *vm.VirtualMachineAgent) VirtualMachineAgentOptions {
	m := VirtualMachineAgentOptions{
		Enabled:   types.BoolValue(agent.Enabled),
		UseFSTrim: types.BoolValue(agent.FsTrim),
	}
	if agent.Type != nil {
		m.Type = types.StringValue(string(*agent.Type))
	}
	return m
}

type VirtualMachineCpuOptions struct {
	Architecture types.String `tfsdk:"architecture"`
	Cores        types.Int64  `tfsdk:"cores"`
	Sockets      types.Int64  `tfsdk:"sockets"`
	EmulatedType types.String `tfsdk:"emulated_type"`
	CPUUnits     types.Int64  `tfsdk:"cpu_units"`
}

func VMCPUToModel(cpu *vm.VirtualMachineCpu) VirtualMachineCpuOptions {
	m := VirtualMachineCpuOptions{
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

type VirtualMachineDiskSpeedLimits struct {
	Read           types.Int64 `tfsdk:"read"`
	ReadBurstable  types.Int64 `tfsdk:"read_burstable"`
	Write          types.Int64 `tfsdk:"write"`
	WriteBurstable types.Int64 `tfsdk:"write_burstable"`
}

type VirtualMachinePciDeviceOptions struct {
	DeviceName types.String `tfsdk:"device_name"`
	DeviceID   types.String `tfsdk:"device_id"`
	PCIe       types.Bool   `tfsdk:"pcie"`
	Mdev       types.String `tfsdk:"mdev"`
}

type VirtualMachineMemoryOptions struct {
	Dedicated types.Int64 `tfsdk:"dedicated"`
	Floating  types.Int64 `tfsdk:"floating"`
	Shared    types.Int64 `tfsdk:"shared"`
}

func VMMemoryToModel(memory *vm.VirtualMachineMemory) VirtualMachineMemoryOptions {
	m := VirtualMachineMemoryOptions{
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

type VirtualMachineCloudInitOptions struct {
	User *VirtualMachineCloudInitUserOptions `tfsdk:"user"`
	IP   []VirtualMachineCloudInitIpOptions  `tfsdk:"ip"`
	DNS  *VirtualMachineCloudInitDnsOptions  `tfsdk:"dns"`
}

type VirtualMachineCloudInitUserOptions struct {
	Name       types.String   `tfsdk:"name"`
	Password   types.String   `tfsdk:"password"`
	PublicKeys []types.String `tfsdk:"public_keys"`
}

func VMCloudInitToModel(cloudInit *vm.VirtualMachineCloudInit) VirtualMachineCloudInitOptions {
	m := VirtualMachineCloudInitOptions{}

	if cloudInit.User != nil {
		m.User = &VirtualMachineCloudInitUserOptions{}
		if cloudInit.User.Name != nil {
			m.User.Name = types.StringValue(*cloudInit.User.Name)
		}
		if cloudInit.User.Password != nil {
			m.User.Password = types.StringValue(*cloudInit.User.Password)
		}
		if cloudInit.User.PublicKeys != nil {
			m.User.PublicKeys = []types.String{}
			for _, key := range cloudInit.User.PublicKeys {
				m.User.PublicKeys = append(m.User.PublicKeys, types.StringValue(key))
			}
		}
	}

	m.IP = []VirtualMachineCloudInitIpOptions{}
	for _, ip := range cloudInit.Ip {
		v4 := &VirtualMachineCloudInitIpConfigOptions{
			DHCP: types.BoolValue(ip.V4.DHCP),
		}
		if ip.V4.Address != nil {
			v4.Address = types.StringValue(*ip.V4.Address)
		}
		if ip.V4.Gateway != nil {
			v4.Gateway = types.StringValue(*ip.V4.Gateway)
		}
		if ip.V4.Netmask != nil {
			v4.Netmask = types.StringValue(*ip.V4.Netmask)
		}

		v6 := &VirtualMachineCloudInitIpConfigOptions{
			DHCP: types.BoolValue(ip.V6.DHCP),
		}
		if ip.V6.Address != nil {
			v6.Address = types.StringValue(*ip.V6.Address)
		}
		if ip.V6.Gateway != nil {
			v6.Gateway = types.StringValue(*ip.V6.Gateway)
		}
		if ip.V6.Netmask != nil {
			v6.Netmask = types.StringValue(*ip.V6.Netmask)
		}
		cfg := VirtualMachineCloudInitIpOptions{
			V4: v4,
			V6: v6,
		}
		m.IP = append(m.IP, cfg)
	}

	if cloudInit.Dns != nil {
		m.DNS = &VirtualMachineCloudInitDnsOptions{}
		if cloudInit.Dns.Nameserver != nil {
			m.DNS.Nameserver = types.StringValue(*cloudInit.Dns.Nameserver)
		}
		if cloudInit.Dns.Domain != nil {
			m.DNS.Domain = types.StringValue(*cloudInit.Dns.Domain)
		}
	}

	return m
}

type VirtualMachineCloudInitIpOptions struct {
	V4 *VirtualMachineCloudInitIpConfigOptions `tfsdk:"v4"`
	V6 *VirtualMachineCloudInitIpConfigOptions `tfsdk:"v6"`
}

type VirtualMachineCloudInitIpConfigOptions struct {
	DHCP    types.Bool   `tfsdk:"dhcp"`
	Address types.String `tfsdk:"address"`
	Netmask types.String `tfsdk:"netmask"`
	Gateway types.String `tfsdk:"gateway"`
}

type VirtualMachineCloudInitDnsOptions struct {
	Nameserver types.String `tfsdk:"nameserver"`
	Domain     types.String `tfsdk:"domain"`
}

type VirtualMachineTerraformTimeouts struct {
	Create   types.Int64 `tfsdk:"create"`
	Delete   types.Int64 `tfsdk:"delete"`
	Stop     types.Int64 `tfsdk:"stop"`
	Start    types.Int64 `tfsdk:"start"`
	Reboot   types.Int64 `tfsdk:"reboot"`
	Shutdown types.Int64 `tfsdk:"shutdown"`
	Clone    types.Int64 `tfsdk:"clone"`
	MoveDisk types.Int64 `tfsdk:"move_disk"`
}

type VirtualMachineModel struct {
	ID         types.Number            `tfsdk:"id"`
	Node       types.String            `tfsdk:"node"`
	Name       types.String            `tfsdk:"name"`
	Cores      types.Number            `tfsdk:"cores"`
	Memory     types.Int64             `tfsdk:"memory"`
	Agent      types.Bool              `tfsdk:"agent"`
	Tags       []types.String          `tfsdk:"tags"`
	Disks      []VirtualDiskModel      `tfsdk:"disks"`
	Interfaces []VirtualInterfaceModel `tfsdk:"network_interfaces"`
}

type VirtualDiskModel struct {
	Storage  types.String `tfsdk:"storage"`
	Size     types.Int64  `tfsdk:"size"`
	Type     types.String `tfsdk:"type"`
	Position types.String `tfsdk:"position"`
	Discard  types.Bool   `tfsdk:"discard"`
}

type VirtualInterfaceModel struct {
	Bridge     types.String `tfsdk:"bridge"`
	Vlan       types.Number `tfsdk:"vlan"`
	Model      types.String `tfsdk:"model"`
	MacAddress types.String `tfsdk:"mac_address"`
	Position   types.String `tfsdk:"position"`
	Firewall   types.Bool   `tfsdk:"firewall"`
}

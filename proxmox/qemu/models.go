package qemu

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualMachineResourceModel struct {
	ID                types.Int64                             `tfsdk:"id"`
	Node              types.String                            `tfsdk:"node"`
	Name              types.String                            `tfsdk:"name"`
	Description       types.String                            `tfsdk:"description"`
	Tags              []types.String                          `tfsdk:"tags"`
	Clone             *VirtualMachineCloneOptions             `tfsdk:"clone"`
	ISO               *VirtualMachineIsoOptions               `tfsdk:"iso"`
	Agent             *VirtualMachineAgentOptions             `tfsdk:"agent"`
	BIOS              types.String                            `tfsdk:"bios"`
	CPU               *VirtualMachineCpuOptions               `tfsdk:"cpu"`
	Disks             []VirtualMachineDiskOptions             `tfsdk:"disks"`
	PCIDevices        []VirtualMachinePciDeviceOptions        `tfsdk:"pci_devices"`
	NetworkInterfaces []VirtualMachineNetworkInterfaceOptions `tfsdk:"network_interfaces"`
	Memory            *VirtualMachineMemoryOptions            `tfsdk:"memory"`
	MachineType       types.String                            `tfsdk:"machine_type"`
	KVMArguments      types.String                            `tfsdk:"kvm_arguments"`
	KeyboardLayout    types.String                            `tfsdk:"keyboard_layout"`
	CloudInit         *VirtualMachineCloudInitOptions         `tfsdk:"cloud_init"`
	Type              types.String                            `tfsdk:"type"`
	ResourcePool      types.String                            `tfsdk:"resource_pool"`
	StartOnCreate     types.Bool                              `tfsdk:"start_on_create"`
	StartOnNodeBoot   types.Bool                              `tfsdk:"start_on_node_boot"`
	Timeouts          *VirtualMachineTerraformTimeouts        `tfsdk:"timeouts"`
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

type VirtualMachineCpuOptions struct {
	Architecture types.String `tfsdk:"architecture"`
	Cores        types.Int64  `tfsdk:"cores"`
	Sockets      types.Int64  `tfsdk:"sockets"`
	EmulatedType types.String `tfsdk:"emulated_type"`
	CPUUnits     types.Int64  `tfsdk:"cpu_units"`
}

type VirtualMachineDiskOptions struct {
	Storage       types.String                   `tfsdk:"storage"`
	FileFormat    types.String                   `tfsdk:"file_format"`
	Size          types.Int64                    `tfsdk:"size"`
	UseIOThread   types.Bool                     `tfsdk:"use_iothread"`
	SpeedLimits   *VirtualMachineDiskSpeedLimits `tfsdk:"speed_limits"`
	InterfaceType types.String                   `tfsdk:"interface_type"`
	SSDEmulation  types.Bool                     `tfsdk:"ssd_emulation"`
	Position      types.Int64                    `tfsdk:"position"`
	Discard       types.Bool                     `tfsdk:"discard"`
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

type VirtualMachineNetworkInterfaceOptions struct {
	Bridge     types.String `tfsdk:"bridge"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	MacAddress types.String `tfsdk:"mac_address"`
	Model      types.String `tfsdk:"model"`
	Vlan       types.Number `tfsdk:"vlan"`
	RateLimit  types.Int64  `tfsdk:"rate_limit"`
	MTU        types.Int64  `tfsdk:"mtu"`
}

type VirtualMachineMemoryOptions struct {
	Dedicated types.Int64 `tfsdk:"dedicated"`
	Floating  types.Int64 `tfsdk:"floating"`
	Shared    types.Int64 `tfsdk:"shared"`
}

type VirtualMachineCloudInitOptions struct {
	User *VirtualMachineCloudInitUserOptions `tfsdk:"user"`
	IP   *VirtualMachineCloudInitIpOptions   `tfsdk:"ip"`
	DNS  *VirtualMachineCloudInitDnsOptions  `tfsdk:"dns"`
}

type VirtualMachineCloudInitUserOptions struct {
	Name       types.String   `tfsdk:"name"`
	Password   types.String   `tfsdk:"password"`
	PublicKeys []types.String `tfsdk:"public_keys"`
}

type VirtualMachineCloudInitIpOptions struct {
	V4 *VirtualMachineCloudInitIpConfigOptions `tfsdk:"v4"`
	V6 *VirtualMachineCloudInitIpConfigOptions `tfsdk:"v6"`
}

type VirtualMachineCloudInitIpConfigOptions struct {
	DHCP    types.Bool   `tfsdk:"dhcp"`
	Address types.String `tfsdk:"address"`
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

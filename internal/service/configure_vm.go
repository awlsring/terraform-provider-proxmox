package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
)

type ConfigureVirtualMachineInput struct {
	Node              string
	VmId              int
	Name              *string
	Tags              []string
	Agent             *ConfigureVirtualMachineAgentOptions
	Bios              *proxmox.VirtualMachineBios
	CPU               *ConfigureVirtualMachineCpuOptions
	Disk              []ConfigureVirtualMachineDiskOptions
	PCIDevices        []ConfigureVirtualPciDeviceOptions
	NetworkInterfaces []ConfigureVirtualMachineNetworkInterfaceOptions
	Memory            *ConfigureVirtualMachineMemoryOptions
	CloudInit         *ConfigureVirtualMachineCloudInitOptions
	OsType            *proxmox.VirtualMachineOperatingSystem
	StartOnBoot       bool
	MachineType       *string
	KVMArguments      *string
	KeyboardLayout    *proxmox.VirtualMachineKeyboard
}

type ConfigureVirtualMachineAgentOptions struct {
	Enabled bool
	FsTrim  bool
	Type    *string
}

type ConfigureVirtualMachineCpuOptions struct {
	Architecture *string
	Cores        *int
	Sockets      *int
	EmulatedType *string
	CpuUnits     *int64
}

type ConfigureVirtualMachineDiskOptions struct {
	Storage       string
	FileFormat    *string
	Size          int
	UseIOThreads  bool
	Position      int
	InterfaceType string
	SpeedLimits   *ConfigureVirtualMachineDiskSpeedLimitsOptions
	SSDEmulation  bool
	Discard       bool
}

type ConfigureVirtualMachineDiskSpeedLimitsOptions struct {
	Read           *int64
	ReadBurstable  *int64
	Write          *int64
	WriteBurstable *int64
}

type ConfigureVirtualPciDeviceOptions struct {
	DeviceName *string
	DeviceId   *string
	PCIe       bool
	Mdev       *string
}

type ConfigureVirtualMachineNetworkInterfaceOptions struct {
	Bridge    string
	Enabled   bool
	Firewall  bool
	MAC       string
	Model     string
	RateLimit *int64
	VLAN      *int
	MTU       *int64
	Position  int
}

type ConfigureVirtualMachineMemoryOptions struct {
	Dedicated *int64
	Shared    *int64
	Floating  *int64
}

type ConfigureVirtualMachineCloudInitOptions struct {
	User *ConfigureVirtualMachineCloudInitUserOptions
	Ip   *ConfigureVirtualMachineCloudInitIpOptions
	Dns  *ConfigureVirtualMachineCloudInitDnsOptions
}

type ConfigureVirtualMachineCloudInitUserOptions struct {
	Name       *string
	Password   *string
	PublicKeys []string
}

type ConfigureVirtualMachineCloudInitIpOptions struct {
	V4 *ConfigureVirtualMachineCloudInitIpConfigOptions
	V6 *ConfigureVirtualMachineCloudInitIpConfigOptions
}

type ConfigureVirtualMachineCloudInitIpConfigOptions struct {
	DHCP    bool
	Address *string
	Gateway *string
}

type ConfigureVirtualMachineCloudInitDnsOptions struct {
	Nameserver *string
	Domain     *string
}

func FormAgentString(agent bool, fstrim bool, t *string) *string {
	agentStr := ""
	if agent {
		agentStr = "1"
	}
	if fstrim {
		agentStr = agentStr + ",fstrim_cloned_disks=1"
	}
	if t != nil {
		agentStr = agentStr + ",type=" + *t
	}
	if agentStr == "" {
		return nil
	}

	return &agentStr
}

func FormDiskString(opts ConfigureVirtualMachineDiskOptions) *string {
	diskstr := opts.Storage + ":" + strconv.Itoa(opts.Size)
	if opts.Discard {
		diskstr = diskstr + ",discard=on"
	}
	if opts.SSDEmulation {
		diskstr = diskstr + ",ssd=on"
	}
	if opts.UseIOThreads {
		diskstr = diskstr + ",iothread=1"
	}
	if opts.FileFormat != nil {
		diskstr = diskstr + ",format=" + *opts.FileFormat
	}
	if opts.SpeedLimits != nil {
		if opts.SpeedLimits.Read != nil {
			diskstr = diskstr + fmt.Sprintf(",mbps_rd=%v", *opts.SpeedLimits.Read)
		}
		if opts.SpeedLimits.Write != nil {
			diskstr = diskstr + fmt.Sprintf(",mbps_wr=%v", *opts.SpeedLimits.Write)
		}
		if opts.SpeedLimits.WriteBurstable != nil {
			diskstr = diskstr + fmt.Sprintf(",mbps_wr_max=%v", *opts.SpeedLimits.WriteBurstable)
		}
		if opts.SpeedLimits.ReadBurstable != nil {
			diskstr = diskstr + fmt.Sprintf(",mbps_rd_max=%v", *opts.SpeedLimits.ReadBurstable)
		}
	}

	return &diskstr
}

func FormNetworkInterfaceString(opts ConfigureVirtualMachineNetworkInterfaceOptions) *string {
	firewallOn := "0"
	isEnabled := "0"

	nicStr := opts.Model + "=" + opts.MAC + ",bridge=" + opts.Bridge
	if opts.VLAN != nil {
		nicStr = nicStr + ",tag=" + strconv.Itoa(*opts.VLAN)
	}
	if opts.Firewall {
		firewallOn = "1"
	}
	if opts.Enabled {
		isEnabled = "1"
	}
	nicStr = nicStr + ",firewall=" + firewallOn + ",link_down=" + isEnabled

	if opts.MTU != nil {
		nicStr = nicStr + fmt.Sprintf(",mtu=%v", *opts.MTU)
	}

	if opts.RateLimit != nil {
		nicStr = nicStr + fmt.Sprintf(",rate=%v", *opts.MTU)
	}

	return &nicStr
}

func FormPCIDeviceString() *string {
	return nil
}

func (c *Proxmox) ConfigureVirtualMachine(ctx context.Context, input *ConfigureVirtualMachineInput) error {
	vmId := strconv.Itoa(input.VmId)

	content := proxmox.ApplyVirtualMachineConfigurationSyncRequestContent{
		Bios:     input.Bios,
		Name:     input.Name,
		Ostype:   input.OsType,
		Machine:  input.MachineType,
		Args:     input.KVMArguments,
		Keyboard: input.KeyboardLayout,
		Tags:     SliceToStringCommaListPtr(input.Tags),
	}
	if input.Agent != nil {
		content.Agent = FormAgentString(input.Agent.Enabled, input.Agent.FsTrim, input.Agent.Type)
	}
	if input.CPU != nil {
		if input.CPU.Architecture != nil {
			arch := proxmox.VirtualMachineArchitecture(*input.CPU.Architecture)
			content.Arch = &arch
		}
		content.Cores = PtrIntToPtrFloat(input.CPU.Cores)
		content.Sockets = PtrIntToPtrFloat(input.CPU.Sockets)
		content.Cpu = input.CPU.EmulatedType
		content.Cpuunits = PtrInt64ToPtrFloat(input.CPU.CpuUnits)
	}
	if input.Memory != nil {
		content.Memory = PtrInt64ToPtrFloat(input.Memory.Dedicated)
		content.Ballon = PtrInt64ToPtrFloat(input.Memory.Floating)
		content.Shares = PtrInt64ToPtrFloat(input.Memory.Shared)
	}
	if input.CloudInit != nil {
		content.Ciuser = input.CloudInit.User.Name
		content.Cipassword = input.CloudInit.User.Password
		content.Sshkeys = StringSliceToLinedStringPtr(input.CloudInit.User.PublicKeys)
	}
	if input.StartOnBoot {
		onboot := float32(1)
		content.Onboot = &onboot
	}

	for _, d := range input.Disk {
		config := FormDiskString(d)
		vm.AllocateDiskConfig(d.InterfaceType, d.Position, config, &content)
	}

	for _, p := range input.PCIDevices {
		fmt.Println(p)
	}

	for _, n := range input.NetworkInterfaces {
		config := FormNetworkInterfaceString(n)
		vm.AllocateNetworkInterfaceConfig(n.Position, config, &content)
	}

	request := c.client.ApplyVirtualMachineConfigurationSync(ctx, input.Node, vmId)
	request = request.ApplyVirtualMachineConfigurationSyncRequestContent(content)

	h, err := c.client.ApplyVirtualMachineConfigurationSyncExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	if input.CloudInit != nil {
		ciRequest := c.client.RegenerateVirtualMachineCloudInit(ctx, input.Node, vmId)
		h, err = c.client.RegenerateVirtualMachineCloudInitExecute(ciRequest)
		if err != nil {
			return errors.ApiError(h, err)
		}
	}

	return nil
}

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
	Node              string                                           `json:"node"`
	VmId              int                                              `json:"vmId"`
	Name              *string                                          `json:"name,omitempty"`
	Tags              []string                                         `json:"tags,omitempty"`
	Delete            []string                                         `json:"delete,omitempty"`
	Description       *string                                          `json:"description,omitempty"`
	Agent             *ConfigureVirtualMachineAgentOptions             `json:"agent,omitempty"`
	Bios              *proxmox.VirtualMachineBios                      `json:"bios,omitempty"`
	CPU               *ConfigureVirtualMachineCpuOptions               `json:"cpu,omitempty"`
	Disks             []ConfigureVirtualMachineDiskOptions             `json:"disks,omitempty"`
	PCIDevices        []ConfigureVirtualPciDeviceOptions               `json:"pciDevices,omitempty"`
	NetworkInterfaces []ConfigureVirtualMachineNetworkInterfaceOptions `json:"networkInterfaces,omitempty"`
	Memory            *ConfigureVirtualMachineMemoryOptions            `json:"memory,omitempty"`
	CloudInit         *ConfigureVirtualMachineCloudInitOptions         `json:"cloudInit,omitempty"`
	OsType            *proxmox.VirtualMachineOperatingSystem           `json:"osType,omitempty"`
	StartOnBoot       bool                                             `json:"startOnBoot,omitempty"`
	MachineType       *string                                          `json:"machineType,omitempty"`
	KVMArguments      *string                                          `json:"kvmArguments,omitempty"`
	KeyboardLayout    *proxmox.VirtualMachineKeyboard                  `json:"keyboardLayout,omitempty"`
}

type ConfigureVirtualMachineAgentOptions struct {
	Enabled bool    `json:"enabled"`
	FsTrim  bool    `json:"fsTrim"`
	Type    *string `json:"type,omitempty"`
}

type ConfigureVirtualMachineCpuOptions struct {
	Architecture *string `json:"architecture,omitempty"`
	Cores        *int    `json:"cores,omitempty"`
	Sockets      *int    `json:"sockets,omitempty"`
	EmulatedType *string `json:"emulatedType,omitempty"`
	CpuUnits     *int64  `json:"cpuUnits,omitempty"`
}

type ConfigureVirtualMachineDiskOptions struct {
	Storage       string                                         `json:"storage"`
	FileFormat    *string                                        `json:"fileFormat,omitempty"`
	Size          int                                            `json:"size"`
	UseIOThreads  bool                                           `json:"useIOThreads"`
	Position      int                                            `json:"position"`
	InterfaceType string                                         `json:"interfaceType"`
	SpeedLimits   *ConfigureVirtualMachineDiskSpeedLimitsOptions `json:"speedLimits,omitempty"`
	SSDEmulation  bool                                           `json:"ssdEmulation"`
	Discard       bool                                           `json:"discard"`
}

type ConfigureVirtualMachineDiskSpeedLimitsOptions struct {
	Read           *int64 `json:"read,omitempty"`
	ReadBurstable  *int64 `json:"readBurstable,omitempty"`
	Write          *int64 `json:"write,omitempty"`
	WriteBurstable *int64 `json:"writeBurstable,omitempty"`
}

type ConfigureVirtualPciDeviceOptions struct {
	DeviceName *string `json:"deviceName,omitempty"`
	DeviceId   *string `json:"deviceId,omitempty"`
	PCIe       bool    `json:"pcie,omitempty"`
	Mdev       *string `json:"mdev,omitempty"`
}

type ConfigureVirtualMachineNetworkInterfaceOptions struct {
	Bridge    string `json:"bridge"`
	Enabled   bool   `json:"enabled"`
	Firewall  bool   `json:"firewall"`
	MAC       string `json:"mac"`
	Model     string `json:"model"`
	RateLimit *int64 `json:"rateLimit,omitempty"`
	VLAN      *int   `json:"vlan,omitempty"`
	MTU       *int64 `json:"mtu,omitempty"`
	Position  int    `json:"position"`
}

type ConfigureVirtualMachineMemoryOptions struct {
	Dedicated *int64 `json:"dedicated,omitempty"`
	Shared    *int64 `json:"shared,omitempty"`
	Floating  *int64 `json:"floating,omitempty"`
}

type ConfigureVirtualMachineCloudInitOptions struct {
	User *ConfigureVirtualMachineCloudInitUserOptions `json:"user,omitempty"`
	Ip   []ConfigureVirtualMachineCloudInitIpOptions  `json:"ip"`
	Dns  *ConfigureVirtualMachineCloudInitDnsOptions  `json:"dns,omitempty"`
}

type ConfigureVirtualMachineCloudInitUserOptions struct {
	Name       *string  `json:"name,omitempty"`
	Password   *string  `json:"password,omitempty"`
	PublicKeys []string `json:"publicKeys"`
}

type ConfigureVirtualMachineCloudInitIpOptions struct {
	Position int                                              `json:"position"`
	V4       *ConfigureVirtualMachineCloudInitIpConfigOptions `json:"v4,omitempty"`
	V6       *ConfigureVirtualMachineCloudInitIpConfigOptions `json:"v6,omitempty"`
}

type ConfigureVirtualMachineCloudInitIpConfigOptions struct {
	DHCP    bool    `json:"dhcp"`
	Address *string `json:"address,omitempty"`
	Gateway *string `json:"gateway,omitempty"`
	Netmask *string `json:"netmask,omitempty"`
}

type ConfigureVirtualMachineCloudInitDnsOptions struct {
	Nameserver *string `json:"nameserver,omitempty"`
	Domain     *string `json:"domain,omitempty"`
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
	isEnabled := "1"

	nicStr := opts.Model + "=" + opts.MAC + ",bridge=" + opts.Bridge
	if opts.VLAN != nil {
		nicStr = nicStr + ",tag=" + strconv.Itoa(*opts.VLAN)
	}
	if opts.Firewall {
		firewallOn = "1"
	}
	if opts.Enabled {
		isEnabled = "0"
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

func FormIpConfigString(v4 *ConfigureVirtualMachineCloudInitIpConfigOptions, v6 *ConfigureVirtualMachineCloudInitIpConfigOptions) *string {
	ipStr := ""
	if v4 != nil {
		if v4.Address != nil && v4.Netmask != nil {
			ipStr = ipStr + "ip=" + fmt.Sprintf("%s/%s", *v4.Address, *v4.Netmask)
		} else if v4.DHCP {
			ipStr = ipStr + "ip=dhcp"
		}
		if v4.Gateway != nil {
			ipStr = ipStr + ",gw=" + *v4.Gateway
		}
	}
	if v6 != nil {
		if ipStr != "" {
			ipStr = ipStr + ","
		}
		if v6.Address != nil && v6.Netmask != nil {
			ipStr = ipStr + "ip6=" + fmt.Sprintf("%s/%s", *v6.Address, *v6.Netmask)
		} else if v6.DHCP {
			ipStr = ipStr + "ip6=dhcp"
		}
		if v6.Gateway != nil {
			ipStr = ipStr + ",gw6=" + *v6.Gateway
		}
	}
	if ipStr == "" {
		return nil
	}

	return &ipStr
}

func FormPCIDeviceString() *string {
	return nil
}

func (c *Proxmox) ConfigureVirtualMachine(ctx context.Context, input *ConfigureVirtualMachineInput) error {
	vmId := strconv.Itoa(input.VmId)

	content := proxmox.ApplyVirtualMachineConfigurationSyncRequestContent{
		Bios:        input.Bios,
		Name:        input.Name,
		Description: input.Description,
		Ostype:      input.OsType,
		Machine:     input.MachineType,
		Args:        input.KVMArguments,
		Keyboard:    input.KeyboardLayout,
		Tags:        SliceToStringCommaListPtr(input.Tags),
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
		if input.CloudInit.User != nil {
			content.Ciuser = input.CloudInit.User.Name
			content.Cipassword = input.CloudInit.User.Password
			content.Sshkeys = EncodeStringList(input.CloudInit.User.PublicKeys)
		}
		if input.CloudInit.Dns != nil {
			content.Searchdomain = input.CloudInit.Dns.Domain
			content.Nameserver = input.CloudInit.Dns.Nameserver
		}
		for _, n := range input.CloudInit.Ip {
			config := FormIpConfigString(n.V4, n.V6)
			fmt.Println("config str: ", *config)
			vm.AllocateCiNetConfig(n.Position, config, &content)
		}
	}
	if input.StartOnBoot {
		onboot := float32(1)
		content.Onboot = &onboot
	}

	if len(input.Delete) > 0 {
		content.Delete = SliceToStringCommaListPtr(input.Delete)
	}

	for _, d := range input.Disks {
		config := FormDiskString(d)
		vm.AllocateDiskConfig(d.InterfaceType, d.Position, config, &content)
	}

	for _, p := range input.PCIDevices {
		fmt.Println(p)
	}

	for _, n := range input.NetworkInterfaces {
		if n.MAC == "" {
			n.MAC = vm.GenerateMAC()
		}
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

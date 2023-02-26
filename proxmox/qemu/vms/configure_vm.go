package vms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	t "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *virtualMachineResource) modifyResourcePool(ctx context.Context, id int, was types.String, is types.String) error {
	tflog.Debug(ctx, "modifyResourcePool virtual machine method")
	tflog.Debug(ctx, fmt.Sprintf("was: %s", was.ValueString()))
	tflog.Debug(ctx, fmt.Sprintf("is: %s", is.ValueString()))

	if was.IsNull() && is.IsNull() {
		tflog.Debug(ctx, "modifyResourcePool was and is null, skipping")
		return nil
	}

	if was.IsNull() && !is.IsNull() {
		tflog.Debug(ctx, "modifyResourcePool was null, adding to pool")
		return r.client.AddVirtualMachineToResourcePool(ctx, id, is.ValueString())
	}

	if !was.IsNull() && is.IsNull() {
		tflog.Debug(ctx, "modifyResourcePool is null, removing from pool")
		return r.client.RemoveVirtualMachineFromResourcePool(ctx, id, was.ValueString())
	}

	if was.ValueString() != is.ValueString() {
		tflog.Debug(ctx, "modifyResourcePool was and is not null, moving between pools")
		err := r.client.RemoveVirtualMachineFromResourcePool(ctx, id, was.ValueString())
		if err != nil {
			return err
		}
		return r.client.AddVirtualMachineToResourcePool(ctx, id, is.ValueString())
	}

	return nil
}

func (r *virtualMachineResource) configureVm(ctx context.Context, plan *qemu.VirtualMachineResourceModel) error {
	tflog.Debug(ctx, "configure virtual machine method")

	node := plan.Node.ValueString()
	vmId := int(plan.ID.ValueInt64())

	request := service.ConfigureVirtualMachineInput{
		Node:              node,
		VmId:              vmId,
		Name:              utils.OptionalToPointerString(plan.Name.ValueString()),
		Description:       utils.OptionalToPointerString(plan.Description.ValueString()),
		Tags:              utils.ListTypeToStringSlice(plan.Tags),
		Agent:             FormAgentConfig(plan.Agent),
		Bios:              FormBIOSConfig(plan.BIOS),
		CPU:               FormCPUConfig(&plan.CPU),
		Disks:             FormDiskConfig(ctx, plan.Disks),
		NetworkInterfaces: FormNetworkInterfaceConfig(plan.NetworkInterfaces),
		Memory:            FormMemoryConfig(&plan.Memory),
		CloudInit:         FormCloudInitConfig(ctx, plan.CloudInit),
		OsType:            FormOSTypeConfig(plan.Type),
		MachineType:       utils.OptionalToPointerString(plan.MachineType.ValueString()),
		KVMArguments:      utils.OptionalToPointerString(plan.KVMArguments.ValueString()),
		KeyboardLayout:    FormKeyboardConfig(plan.KeyboardLayout),
		StartOnBoot:       plan.StartOnNodeBoot.ValueBool(),
	}

	info, _ := json.Marshal(request)
	tflog.Debug(ctx, "configure virtual machine request: "+string(info))

	err := r.client.ConfigureVirtualMachine(ctx, &request)
	if err != nil {
		tflog.Error(ctx, "configure recieved error: "+err.Error())
		return err
	}
	tflog.Debug(ctx, "configure virtual machine complete")

	tflog.Debug(ctx, "waiting for lock")
	r.waitForLock(ctx, node, vmId)
	tflog.Debug(ctx, "lock released")
	return nil
}

func FormNetworkInterfaceConfig(nic t.VirtualMachineNetworkInterfaceSetValue) []service.ConfigureVirtualMachineNetworkInterfaceOptions {
	n := make([]service.ConfigureVirtualMachineNetworkInterfaceOptions, len(nic.Nics))
	for i, v := range nic.Nics {
		nconfig := service.ConfigureVirtualMachineNetworkInterfaceOptions{
			Bridge:    v.Bridge.ValueString(),
			Enabled:   v.Enabled.ValueBool(),
			Firewall:  v.UseFirewall.ValueBool(),
			MAC:       v.MacAddress.ValueString(),
			Model:     v.Model.ValueString(),
			RateLimit: utils.OptionaToPointerInt64(v.RateLimit.ValueInt64()),
			VLAN:      utils.OptionaInt64ToPointerInt(v.Vlan.ValueInt64()),
			MTU:       utils.OptionaToPointerInt64(v.MTU.ValueInt64()),
			Position:  int(v.Position.ValueInt64()),
		}
		n[i] = nconfig
	}
	return n
}

func FormDiskConfig(ctx context.Context, disks t.VirtualMachineDiskSetValue) []service.ConfigureVirtualMachineDiskOptions {
	tflog.Debug(ctx, "Entered form disk config")
	d := make([]service.ConfigureVirtualMachineDiskOptions, len(disks.Disks))
	tflog.Debug(ctx, fmt.Sprintf("Disks to configure: %v", len(disks.Disks)))
	for i, v := range disks.Disks {
		dconfig := service.ConfigureVirtualMachineDiskOptions{
			Storage:       v.Storage.ValueString(),
			FileFormat:    utils.OptionalToPointerString(v.FileFormat.ValueString()),
			Size:          int(v.Size.ValueInt64()),
			UseIOThreads:  v.UseIOThread.ValueBool(),
			Position:      int(v.Position.ValueInt64()),
			InterfaceType: v.InterfaceType.ValueString(),
			SSDEmulation:  v.SSDEmulation.ValueBool(),
			Discard:       v.Discard.ValueBool(),
		}

		if v.SpeedLimits != nil {
			s := service.ConfigureVirtualMachineDiskSpeedLimitsOptions{}
			s.Read = utils.OptionaToPointerInt64(v.SpeedLimits.Read.ValueInt64())
			s.Write = utils.OptionaToPointerInt64(v.SpeedLimits.Write.ValueInt64())
			s.ReadBurstable = utils.OptionaToPointerInt64(v.SpeedLimits.ReadBurstable.ValueInt64())
			s.WriteBurstable = utils.OptionaToPointerInt64(v.SpeedLimits.WriteBurstable.ValueInt64())
			dconfig.SpeedLimits = &s
		}

		tflog.Debug(ctx, fmt.Sprintf("Disk %v: %v", i, dconfig))

		d[i] = dconfig
	}
	return d
}

func FormBIOSConfig(bios basetypes.StringValue) *proxmox.VirtualMachineBios {
	b := proxmox.VirtualMachineBios(bios.ValueString())
	return &b
}

func FormOSTypeConfig(osType basetypes.StringValue) *proxmox.VirtualMachineOperatingSystem {
	o := proxmox.VirtualMachineOperatingSystem(osType.ValueString())
	return &o
}

func FormKeyboardConfig(key basetypes.StringValue) *proxmox.VirtualMachineKeyboard {
	k := proxmox.VirtualMachineKeyboard(key.ValueString())
	return &k
}

func FormAgentConfig(agent *qemu.VirtualMachineAgentOptions) *service.ConfigureVirtualMachineAgentOptions {
	if agent == nil {
		return nil
	}
	a := service.ConfigureVirtualMachineAgentOptions{
		Enabled: agent.Enabled.ValueBool(),
		FsTrim:  agent.UseFSTrim.ValueBool(),
		Type:    utils.OptionalToPointerString(agent.Type.ValueString()),
	}
	return &a
}

func FormCPUConfig(cpu *qemu.VirtualMachineCpuOptions) *service.ConfigureVirtualMachineCpuOptions {
	if cpu == nil {
		return nil
	}
	c := service.ConfigureVirtualMachineCpuOptions{
		Architecture: utils.OptionalToPointerString(cpu.Architecture.ValueString()),
		Cores:        utils.OptionaInt64ToPointerInt(cpu.Cores.ValueInt64()),
		Sockets:      utils.OptionaInt64ToPointerInt(cpu.Sockets.ValueInt64()),
		EmulatedType: utils.OptionalToPointerString(cpu.EmulatedType.ValueString()),
		CpuUnits:     utils.OptionaToPointerInt64(cpu.CPUUnits.ValueInt64()),
	}
	return &c
}

func FormMemoryConfig(mem *qemu.VirtualMachineMemoryOptions) *service.ConfigureVirtualMachineMemoryOptions {
	if mem == nil {
		return nil
	}

	m := service.ConfigureVirtualMachineMemoryOptions{
		Dedicated: utils.OptionaToPointerInt64(mem.Dedicated.ValueInt64()),
		Shared:    utils.OptionaToPointerInt64(mem.Shared.ValueInt64()),
		Floating:  utils.OptionaToPointerInt64(mem.Floating.ValueInt64()),
	}

	return &m
}

func FormCloudInitConfig(ctx context.Context, ci *t.VirtualMachineCloudInitModel) *service.ConfigureVirtualMachineCloudInitOptions {
	tflog.Debug(ctx, fmt.Sprintf("Cloud Init: %v", ci))
	if ci != nil {
		c := service.ConfigureVirtualMachineCloudInitOptions{}
		if ci.User != nil {
			tflog.Debug(ctx, fmt.Sprintf("Configuring user: %v", ci.User))
			user := service.ConfigureVirtualMachineCloudInitUserOptions{}
			if !ci.User.Name.IsNull() || ci.User.Name.IsUnknown() {
				user.Name = utils.OptionalToPointerString(ci.User.Name.ValueString())
			}
			if !ci.User.Password.IsNull() || ci.User.Password.IsUnknown() {
				user.Password = utils.OptionalToPointerString(ci.User.Password.ValueString())
			}
			user.PublicKeys = utils.ListTypeToStringSlice(ci.User.PublicKeys)
			tflog.Debug(ctx, fmt.Sprintf("Determined User: %v", user))
			c.User = &user
		}

		tflog.Debug(ctx, fmt.Sprintf("IP Config amount: %v", len(ci.IP.Configs)))
		configs := []service.ConfigureVirtualMachineCloudInitIpOptions{}
		for _, ipc := range ci.IP.Configs {
			tflog.Debug(ctx, fmt.Sprintf("IP Config: %v", ipc))
			ip := service.ConfigureVirtualMachineCloudInitIpOptions{}
			if ipc.V4 != nil {
				ip.V4 = FormCloudInitIpConfig(ctx, ipc.V4)
			}
			if ipc.V6 != nil {
				ip.V6 = FormCloudInitIpConfig(ctx, ipc.V6)
			}
			tflog.Debug(ctx, fmt.Sprintf("Determined IP Config: %v", ip))
			configs = append(configs, ip)
		}
		c.Ip = configs

		if ci.DNS != nil {
			tflog.Debug(ctx, fmt.Sprintf("Configuring DNS: %v", ci.DNS))
			dns := service.ConfigureVirtualMachineCloudInitDnsOptions{}
			if !ci.DNS.Nameserver.IsNull() || !ci.DNS.Nameserver.IsUnknown() {
				dns.Nameserver = utils.OptionalToPointerString(ci.DNS.Nameserver.ValueString())
			}
			if !ci.DNS.Domain.IsNull() || !ci.DNS.Domain.IsUnknown() {
				dns.Domain = utils.OptionalToPointerString(ci.DNS.Domain.ValueString())
			}
			tflog.Debug(ctx, fmt.Sprintf("Determined DNS: %v", dns))
			c.Dns = &dns
		}

		return &c
	}
	tflog.Debug(ctx, "Cloud Init is nil")
	return nil
}

func FormCloudInitIpConfig(ctx context.Context, ipConfig *t.VirtualMachineCloudInitIpConfigModel) *service.ConfigureVirtualMachineCloudInitIpConfigOptions {
	config := service.ConfigureVirtualMachineCloudInitIpConfigOptions{}

	tflog.Debug(ctx, fmt.Sprintf("IP Config address: %v", ipConfig.Address.ValueString()))
	tflog.Debug(ctx, fmt.Sprintf("IP Config netmask: %v", ipConfig.Netmask.ValueString()))

	if !ipConfig.DHCP.IsNull() || !ipConfig.DHCP.IsUnknown() {
		config.DHCP = ipConfig.DHCP.ValueBool()
	}
	if !ipConfig.Gateway.IsNull() || !ipConfig.Gateway.IsUnknown() {
		config.Gateway = utils.OptionalToPointerString(ipConfig.Gateway.ValueString())
	}
	if (!ipConfig.Address.IsNull() || !ipConfig.Address.IsUnknown()) && (!ipConfig.Netmask.IsNull() || !ipConfig.Netmask.IsUnknown()) {
		config.Address = utils.OptionalToPointerString(ipConfig.Address.ValueString())
		config.Netmask = utils.OptionalToPointerString(ipConfig.Netmask.ValueString())
	}
	return &config
}

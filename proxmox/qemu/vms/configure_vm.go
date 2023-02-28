package vms

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	t "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/r3labs/diff/v3"
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

func (r *virtualMachineResource) determineVmConfigurations(ctx context.Context, state *qemu.VirtualMachineResourceModel, plan *qemu.VirtualMachineResourceModel) error {
	tflog.Debug(ctx, "determine virtual machine configurations method")
	configureRequest, resizeRequests, err := formConfigureRequests(ctx, state, plan)
	if err != nil {
		return err
	}
	node := plan.Node.ValueString()
	vmId := int(plan.ID.ValueInt64())

	info, _ := json.Marshal(configureRequest)
	tflog.Debug(ctx, "configure virtual machine request: "+string(info))
	info, _ = json.Marshal(resizeRequests)
	tflog.Debug(ctx, "resize disk virtual machine request: "+string(info))

	if configureRequest != nil {
		err = r.configureVm(ctx, node, vmId, configureRequest)
		if err != nil {
			return err
		}
	}

	for _, resizeRequest := range resizeRequests {
		err = r.resizeDisk(ctx, node, vmId, &resizeRequest)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *virtualMachineResource) configureVm(ctx context.Context, node string, vmId int, request *service.ConfigureVirtualMachineInput) error {
	tflog.Debug(ctx, "configure virtual machine method")

	err := r.client.ConfigureVirtualMachine(ctx, request)
	if err != nil {
		tflog.Error(ctx, "configure recieved error: "+err.Error())
		return err
	}
	tflog.Debug(ctx, "configure virtual machine complete")

	tflog.Debug(ctx, "waiting for lock")
	r.waitForLock(ctx, node, vmId, r.timeouts.Configure)
	tflog.Debug(ctx, "lock released")
	return nil
}

func (r *virtualMachineResource) resizeDisk(ctx context.Context, node string, vmId int, request *service.ResizeVirtualMachineDiskInput) error {
	err := r.client.ResizeVirtualMachineDisk(ctx, request)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, "waiting for lock")
	r.waitForLock(ctx, node, vmId, r.timeouts.Configure)
	tflog.Debug(ctx, "lock released")

	return nil
}

func formConfigureRequests(ctx context.Context, state *qemu.VirtualMachineResourceModel, plan *qemu.VirtualMachineResourceModel) (*service.ConfigureVirtualMachineInput, []service.ResizeVirtualMachineDiskInput, error) {
	node := plan.Node.ValueString()
	vmId := int(plan.ID.ValueInt64())

	pre := state
	if state == nil {
		pre = &qemu.VirtualMachineResourceModel{}
	}

	diff, err := diff.Diff(*pre, *plan)
	if err != nil {
		return nil, nil, err
	}
	j, _ := json.Marshal(diff)
	tflog.Debug(ctx, fmt.Sprintf("diff: %+v", string(j)))

	r := service.ConfigureVirtualMachineInput{
		Node: node,
		VmId: vmId,
	}

	resizeRequests := []service.ResizeVirtualMachineDiskInput{}

	do := false
	for _, d := range diff {
		if len(d.Path) == 0 {
			continue
		}
		field := d.Path[0]
		switch d.Type {
		case "update":
			if d.From != d.To {
				if field == "Disks" {
					if len(d.Path) <= 4 {
						continue
					}
					if d.Path[3] == "Size" {
						i, err := strconv.Atoi(d.Path[2])
						if err != nil {
							return nil, nil, err
						}
						resizeRequests = append(resizeRequests, formResizeRequest(ctx, i, plan))
					}
				} else {
					do = true
					setField(ctx, field, plan, &r)
				}
			}
		case "create":
			do = true
			setField(ctx, field, plan, &r)
		case "delete":
			// create some conditional branch to formulate a removal
			do = true
			setField(ctx, field, plan, &r)
		}
	}

	if do {
		return &r, resizeRequests, nil
	}
	return nil, resizeRequests, nil
}

func setField(ctx context.Context, f string, m *qemu.VirtualMachineResourceModel, r *service.ConfigureVirtualMachineInput) {
	switch f {
	case "Node":
		r.Node = m.Node.ValueString()
	case "VmId":
		r.VmId = int(m.ID.ValueInt64())
	case "Name":
		r.Name = utils.OptionalToPointerString(m.Name.ValueString())
	case "Description":
		r.Description = utils.OptionalToPointerString(m.Description.ValueString())
	case "Tags":
		r.Tags = utils.SetTypeToStringSlice(m.Tags)
	case "Agent":
		r.Agent = FormAgentConfig(m.Agent)
	case "Bios":
		r.Bios = FormBIOSConfig(m.BIOS)
	case "CPU":
		r.CPU = FormCPUConfig(&m.CPU)
	case "Disks":
		r.Disks = FormDiskConfig(ctx, m.Disks)
	case "NetworkInterfaces":
		r.NetworkInterfaces = FormNetworkInterfaceConfig(m.NetworkInterfaces)
	case "Memory":
		r.Memory = FormMemoryConfig(&m.Memory)
	case "CloudInit":
		r.CloudInit = FormCloudInitConfig(ctx, m.CloudInit)
	case "OsType":
		r.OsType = FormOSTypeConfig(m.Type)
	case "MachineType":
		r.MachineType = utils.OptionalToPointerString(m.MachineType.ValueString())
	case "KVMArguments":
		r.KVMArguments = utils.OptionalToPointerString(m.KVMArguments.ValueString())
	case "KeyboardLayout":
		r.KeyboardLayout = FormKeyboardConfig(m.KeyboardLayout)
	case "StartOnNodeBoot":
		r.StartOnBoot = m.StartOnNodeBoot.ValueBool()
	}
}

func formResizeRequest(ctx context.Context, elm int, m *qemu.VirtualMachineResourceModel) service.ResizeVirtualMachineDiskInput {
	d := m.Disks.Disks[elm]
	disk := fmt.Sprintf("%v%v", d.InterfaceType.ValueString(), d.Position.ValueInt64())
	return service.ResizeVirtualMachineDiskInput{
		Node: m.Node.ValueString(),
		VmId: int(m.ID.ValueInt64()),
		Disk: disk,
		Size: utils.GbToBytes(d.Size.ValueInt64()),
	}
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

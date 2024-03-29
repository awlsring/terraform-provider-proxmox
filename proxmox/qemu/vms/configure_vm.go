package vms

import (
	"context"
	"fmt"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	ct "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	vt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/vms/types"
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

func (r *virtualMachineResource) determineVmConfigurations(ctx context.Context, state *vt.VirtualMachineResourceModel, plan *vt.VirtualMachineResourceModel) error {
	tflog.Debug(ctx, "determine virtual machine configurations method")
	updates, deletes := formConfigureRequests(ctx, state, plan)

	resizes, err := r.formResizeRequests(ctx, plan.Node.ValueString(), int(plan.ID.ValueInt64()), plan)
	if err != nil {
		return err
	}

	node := plan.Node.ValueString()
	vmId := int(plan.ID.ValueInt64())

	tflog.Debug(ctx, "configure virtual machine updates request: "+utils.MarshalSafe(updates))
	tflog.Debug(ctx, "configure virtual machine deletes request: "+utils.MarshalSafe(deletes))
	tflog.Debug(ctx, "resize disk virtual machine request: "+utils.MarshalSafe(resizes))

	if deletes != nil {
		tflog.Debug(ctx, "deletes not nil, running configure vm with delete method")
		err := r.configureVm(ctx, node, vmId, deletes)
		if err != nil {
			return err
		}
	}

	if updates != nil {
		tflog.Debug(ctx, "updates not nil, running configure vm with updated properties")
		err := r.configureVm(ctx, node, vmId, updates)
		if err != nil {
			return err
		}
	}

	for _, resizeRequest := range resizes {
		tflog.Debug(ctx, fmt.Sprintf("running resize disk method on %s", resizeRequest.Disk))
		err := r.resizeDisk(ctx, node, vmId, &resizeRequest)
		if err != nil {
			return err
		}
	}

	err = r.removeUnusedDisks(ctx, node, vmId)
	if err != nil {
		return err
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
	err = r.waitForLock(ctx, node, vmId, r.timeouts.Configure)
	if err != nil {
		return err
	}
	tflog.Debug(ctx, "lock released")
	return nil
}

func (r *virtualMachineResource) resizeDisk(ctx context.Context, node string, vmId int, request *service.ResizeVirtualMachineDiskInput) error {
	err := r.client.ResizeVirtualMachineDisk(ctx, request)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, "waiting for lock")
	err = r.waitForLock(ctx, node, vmId, r.timeouts.Configure)
	if err != nil {
		return err
	}
	tflog.Debug(ctx, "lock released")

	return nil
}

func formConfigureRequests(
	ctx context.Context,
	state *vt.VirtualMachineResourceModel,
	plan *vt.VirtualMachineResourceModel,
) (
	*service.ConfigureVirtualMachineInput,
	*service.ConfigureVirtualMachineInput,
) {
	node := plan.Node.ValueString()
	vmId := int(plan.ID.ValueInt64())

	updates := determineUpdates(ctx, node, vmId, state, plan)
	deletes := determineDeletes(ctx, node, vmId, state, plan)

	return updates, deletes
}

func (r *virtualMachineResource) formResizeRequests(ctx context.Context, node string, vmId int, plan *vt.VirtualMachineResourceModel) ([]service.ResizeVirtualMachineDiskInput, error) {
	current, err := r.client.DescribeVirtualMachine(ctx, node, vmId)
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "determine disk resizes method")
	var resizes []service.ResizeVirtualMachineDiskInput

	for _, planDisk := range plan.Disks.Disks {
		for _, currentDisk := range current.Disks {
			if int(planDisk.Position.ValueInt64()) == currentDisk.Position && planDisk.InterfaceType.ValueString() == currentDisk.InterfaceType && planDisk.Storage.ValueString() == currentDisk.Storage {
				if planDisk.Size.ValueInt64() != currentDisk.Size {
					disk := fmt.Sprintf("%s%v", planDisk.InterfaceType.ValueString(), planDisk.Position.ValueInt64())
					resizes = append(resizes, service.ResizeVirtualMachineDiskInput{
						Node: node,
						VmId: vmId,
						Disk: disk,
						Size: utils.GbToBytes(planDisk.Size.ValueInt64()),
					})
				}
			}
		}
	}

	return resizes, nil
}

func (r *virtualMachineResource) removeUnusedDisks(ctx context.Context, node string, vmId int) error {
	current, err := r.client.DescribeVirtualMachine(ctx, node, vmId)
	if err != nil {
		return err
	}

	deletes := []string{}
	for _, disk := range current.Disks {
		if disk.InterfaceType == "unused" {
			deletes = append(deletes, fmt.Sprintf("%s%v", disk.InterfaceType, disk.Position))
		}
	}

	if len(deletes) != 0 {
		tflog.Debug(ctx, fmt.Sprintf("removing unused disks %v", deletes))
		err = r.configureVm(ctx, node, vmId, &service.ConfigureVirtualMachineInput{
			Node:   node,
			VmId:   vmId,
			Delete: deletes,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func determineUpdates(ctx context.Context, node string, vmId int, state *vt.VirtualMachineResourceModel, plan *vt.VirtualMachineResourceModel) *service.ConfigureVirtualMachineInput {
	tflog.Debug(ctx, "determine configuration updates method")
	request := &service.ConfigureVirtualMachineInput{
		Node:              node,
		VmId:              vmId,
		Memory:            FormMemoryConfig(&plan.Memory),
		CPU:               FormCPUConfig(&plan.CPU),
		NetworkInterfaces: FormNetworkInterfaceConfig(plan.NetworkInterfaces.Nics),
	}

	old := state
	if state == nil {
		old = &vt.VirtualMachineResourceModel{}
	}

	if old.Name.ValueString() != plan.Name.ValueString() {
		request.Name = utils.OptionalToPointerString(plan.Name.ValueString())
	}

	if old.Description.ValueString() != plan.Description.ValueString() {
		request.Description = utils.OptionalToPointerString(plan.Description.ValueString())
	}

	if !plan.Tags.IsNull() {
		request.Tags = utils.SetTypeToStringSlice(plan.Tags)
	}

	newDisks, existingDisks := determineAddedDisks(ctx, old.Disks.Disks, plan.Disks.Disks)
	if len(newDisks) > 0 {
		request.Disks = FormDiskConfig(ctx, newDisks, true)
	}
	request.Disks = append(request.Disks, FormDiskConfig(ctx, existingDisks, false)...)

	if plan.Agent != nil {
		request.Agent = FormAgentConfig(plan.Agent)
	}

	if !plan.BIOS.IsNull() {
		request.Bios = FormBIOSConfig(plan.BIOS)
	}

	if plan.CloudInit != nil {
		request.CloudInit = FormCloudInitConfig(ctx, plan.CloudInit)
	}

	if !plan.Type.IsNull() {
		request.OsType = FormOSTypeConfig(plan.Type)
	}

	if !plan.MachineType.IsNull() {
		request.MachineType = utils.OptionalToPointerString(plan.MachineType.ValueString())
	}

	if !plan.KVMArguments.IsNull() {
		request.KVMArguments = utils.OptionalToPointerString(plan.KVMArguments.ValueString())
	}

	if !plan.KeyboardLayout.IsNull() {
		kl := plan.KeyboardLayout.ValueString()
		k := proxmox.VirtualMachineKeyboard(kl)
		request.KeyboardLayout = &k
	}

	if plan.StartOnNodeBoot.ValueBool() {
		request.StartOnBoot = plan.StartOnNodeBoot.ValueBool()
	}

	return request
}

func determineDeletes(ctx context.Context, node string, vmId int, state *vt.VirtualMachineResourceModel, plan *vt.VirtualMachineResourceModel) *service.ConfigureVirtualMachineInput {
	tflog.Debug(ctx, "determine configuration deletes method")
	fieldsToDelete := []string{}

	old := state
	if state == nil {
		old = &vt.VirtualMachineResourceModel{}
	}

	if !old.Name.IsNull() && (plan.Name.IsNull() || plan.Name.IsUnknown()) {
		tflog.Debug(ctx, "name is null, will delete")
		fieldsToDelete = append(fieldsToDelete, "name")
	}

	if !old.Description.IsNull() && plan.Description.IsNull() {
		tflog.Debug(ctx, "description is null, will delete")
		fieldsToDelete = append(fieldsToDelete, "description")
	}

	if !old.Tags.IsNull() && plan.Tags.IsNull() {
		tflog.Debug(ctx, "tags are null, will delete")
		fieldsToDelete = append(fieldsToDelete, "tags")
	}

	removedDisks := determineRemovedDisks(ctx, old.Disks.Disks, plan.Disks.Disks)
	if len(removedDisks) > 0 {
		fieldsToDelete = append(fieldsToDelete, removedDisks...)
	}

	removedNics := determineRemovedNetworkInterfaces(ctx, old.NetworkInterfaces.Nics, plan.NetworkInterfaces.Nics)
	if len(removedNics) > 0 {
		fieldsToDelete = append(fieldsToDelete, removedNics...)
	}

	oldCfgs := []ct.VirtualMachineCloudInitIpModel{}
	if old.CloudInit != nil {
		if !plan.CloudInit.IP.IsNull() {
			oldCfgs = old.CloudInit.IP.Configs
		}
	}

	planCfgs := []ct.VirtualMachineCloudInitIpModel{}
	if plan.CloudInit != nil {
		if !plan.CloudInit.IP.IsNull() {
			planCfgs = plan.CloudInit.IP.Configs
		}
	}

	removedIpCfg := determineRemovedIpConfig(ctx, oldCfgs, planCfgs)
	if len(removedIpCfg) > 0 {
		fieldsToDelete = append(fieldsToDelete, removedIpCfg...)
	}

	if len(fieldsToDelete) != 0 {
		return &service.ConfigureVirtualMachineInput{
			Node:   node,
			VmId:   vmId,
			Delete: fieldsToDelete,
		}
	}
	return nil
}

func flattenNetworkInterfaces(ctx context.Context, state []ct.VirtualMachineNetworkInterfaceModel, plan []ct.VirtualMachineNetworkInterfaceModel) ([]string, []string) {
	stateNics := []string{}
	planNics := []string{}

	for _, nic := range state {
		stateNics = append(stateNics, fmt.Sprintf("net%v", nic.Position.ValueInt64()))
	}

	for _, nic := range plan {
		planNics = append(planNics, fmt.Sprintf("net%v", nic.Position.ValueInt64()))
	}

	return stateNics, planNics
}

func determineRemovedIpConfig(ctx context.Context, state []ct.VirtualMachineCloudInitIpModel, plan []ct.VirtualMachineCloudInitIpModel) []string {
	stateCfg, planCfg := flattenIpConfigs(ctx, state, plan)

	removeCfg := []string{}
	for _, nic := range stateCfg {
		if !utils.ListContains(planCfg, nic) {
			removeCfg = append(removeCfg, nic)
		}
	}

	return removeCfg
}

func flattenIpConfigs(ctx context.Context, state []ct.VirtualMachineCloudInitIpModel, plan []ct.VirtualMachineCloudInitIpModel) ([]string, []string) {
	stateConfigs := []string{}
	planConfigs := []string{}

	for _, cfg := range state {
		stateConfigs = append(stateConfigs, fmt.Sprintf("ipconfig%v", cfg.Position.ValueInt64()))
	}

	for _, cfg := range plan {
		planConfigs = append(planConfigs, fmt.Sprintf("ipconfig%v", cfg.Position.ValueInt64()))
	}

	return stateConfigs, planConfigs
}

func determineRemovedNetworkInterfaces(ctx context.Context, state []ct.VirtualMachineNetworkInterfaceModel, plan []ct.VirtualMachineNetworkInterfaceModel) []string {
	stateNics, planNics := flattenNetworkInterfaces(ctx, state, plan)

	removeNics := []string{}
	for _, nic := range stateNics {
		if !utils.ListContains(planNics, nic) {
			removeNics = append(removeNics, nic)
		}
	}

	return removeNics
}

func determineRemovedDisks(ctx context.Context, state []ct.VirtualMachineDiskModel, plan []ct.VirtualMachineDiskModel) []string {
	stateDisks, planDisks := flattenDisks(ctx, state, plan)
	tflog.Debug(ctx, fmt.Sprintf("stateDisks: %v", stateDisks))
	tflog.Debug(ctx, fmt.Sprintf("planDisks: %v", planDisks))

	removeDisks := []string{}
	for _, disk := range stateDisks {
		if !utils.ListContains(planDisks, disk) {
			tflog.Debug(ctx, fmt.Sprintf("disk to remove: %s", disk))
			d := strings.Split(disk, ":")[0]
			removeDisks = append(removeDisks, d)
		}
	}

	return removeDisks
}

func flattenDisks(ctx context.Context, state []ct.VirtualMachineDiskModel, plan []ct.VirtualMachineDiskModel) ([]string, []string) {
	stateDisks := []string{}
	planDisks := []string{}

	for _, disk := range state {
		stateDisks = append(stateDisks, fmt.Sprintf("%s%v:%s", disk.InterfaceType.ValueString(), disk.Position.ValueInt64(), disk.Storage.ValueString()))
	}

	for _, disk := range plan {
		planDisks = append(planDisks, fmt.Sprintf("%s%v:%s", disk.InterfaceType.ValueString(), disk.Position.ValueInt64(), disk.Storage.ValueString()))
	}

	return stateDisks, planDisks
}

func determineAddedDisks(ctx context.Context, state []ct.VirtualMachineDiskModel, plan []ct.VirtualMachineDiskModel) ([]ct.VirtualMachineDiskModel, []ct.VirtualMachineDiskModel) {
	stateDisks, planDisks := flattenDisks(ctx, state, plan)

	addedDisks := []ct.VirtualMachineDiskModel{}
	existingDisks := []ct.VirtualMachineDiskModel{}
	for _, disk := range planDisks {
		if !utils.ListContains(stateDisks, disk) {
			addedDisks = append(addedDisks, formDiskModel(disk, plan)...)
		} else {
			diskModels := formDiskModel(disk, plan)

			for _, diskModel := range diskModels {
				for _, s := range state {
					d1 := fmt.Sprintf("%s%v", diskModel.InterfaceType.ValueString(), diskModel.Position.ValueInt64())
					d2 := fmt.Sprintf("%s%v", s.InterfaceType.ValueString(), s.Position.ValueInt64())
					if d1 == d2 {
						tflog.Info(ctx, fmt.Sprintf("size of plan disk %s is %v", d1, diskModel.Size.ValueInt64()))
						tflog.Info(ctx, fmt.Sprintf("size of state disk %s is %v", d2, s.Size.ValueInt64()))
						diskModel.Size = s.Size
						diskModel.Name = s.Name
						existingDisks = append(existingDisks, diskModel)
					}
				}
			}
		}
	}

	return addedDisks, existingDisks
}

func formDiskModel(disk string, models []ct.VirtualMachineDiskModel) []ct.VirtualMachineDiskModel {
	disks := []ct.VirtualMachineDiskModel{}
	d := strings.Split(disk, ":")[0]
	for _, pd := range models {
		if fmt.Sprintf("%s%v", pd.InterfaceType.ValueString(), pd.Position.ValueInt64()) == d {
			disks = append(disks, pd)
		}
	}
	return disks
}

func FormNetworkInterfaceConfig(nics []ct.VirtualMachineNetworkInterfaceModel) []service.ConfigureVirtualMachineNetworkInterfaceOptions {
	n := make([]service.ConfigureVirtualMachineNetworkInterfaceOptions, len(nics))
	for i, v := range nics {
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

func FormDiskConfig(ctx context.Context, disks []ct.VirtualMachineDiskModel, new bool) []service.ConfigureVirtualMachineDiskOptions {
	tflog.Debug(ctx, "Entered form disk config")
	d := make([]service.ConfigureVirtualMachineDiskOptions, len(disks))
	tflog.Debug(ctx, fmt.Sprintf("Disks to configure: %v", len(disks)))
	for i, v := range disks {
		dconfig := service.ConfigureVirtualMachineDiskOptions{
			Storage:       v.Storage.ValueString(),
			FileFormat:    utils.OptionalToPointerString(v.FileFormat.ValueString()),
			UseIOThreads:  v.UseIOThread.ValueBool(),
			Size:          int(v.Size.ValueInt64()),
			Position:      int(v.Position.ValueInt64()),
			InterfaceType: v.InterfaceType.ValueString(),
			SSDEmulation:  v.SSDEmulation.ValueBool(),
			Discard:       v.Discard.ValueBool(),
			New:           new,
		}

		if !v.Name.IsNull() || !v.Name.IsUnknown() {
			name := v.Name.ValueString()
			dconfig.Name = &name
		}

		if v.SpeedLimits != nil {
			s := service.ConfigureVirtualMachineDiskSpeedLimitsOptions{}
			s.Read = utils.OptionaToPointerInt64(v.SpeedLimits.Read.ValueInt64())
			s.Write = utils.OptionaToPointerInt64(v.SpeedLimits.Write.ValueInt64())
			s.ReadBurstable = utils.OptionaToPointerInt64(v.SpeedLimits.ReadBurstable.ValueInt64())
			s.WriteBurstable = utils.OptionaToPointerInt64(v.SpeedLimits.WriteBurstable.ValueInt64())
			dconfig.SpeedLimits = &s
		}

		tflog.Info(ctx, fmt.Sprintf("Disk %v: %v", i, dconfig))

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

func FormAgentConfig(agent *ct.VirtualMachineAgentModel) *service.ConfigureVirtualMachineAgentOptions {
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

func FormCPUConfig(cpu *ct.VirtualMachineCpuModel) *service.ConfigureVirtualMachineCpuOptions {
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

func FormMemoryConfig(mem *ct.VirtualMachineMemoryModel) *service.ConfigureVirtualMachineMemoryOptions {
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

func FormCloudInitConfig(ctx context.Context, ci *ct.VirtualMachineCloudInitModel) *service.ConfigureVirtualMachineCloudInitOptions {
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
			user.PublicKeys = utils.SetTypeToStringSlice(ci.User.PublicKeys)
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

func FormCloudInitIpConfig(ctx context.Context, ipConfig *ct.VirtualMachineCloudInitIpConfigModel) *service.ConfigureVirtualMachineCloudInitIpConfigOptions {
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

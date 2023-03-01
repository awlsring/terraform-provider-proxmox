package vms

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func disksAreSame(sd types.VirtualMachineDiskModel, pd types.VirtualMachineDiskModel) bool {
	if sd.Storage.ValueString() != pd.Storage.ValueString() {
		return false
	}
	if sd.Position.ValueInt64() != pd.Position.ValueInt64() {
		return false
	}
	if sd.InterfaceType.ValueString() != pd.InterfaceType.ValueString() {
		return false
	}
	return true
}

func isDiskInList(disk types.VirtualMachineDiskModel, list []types.VirtualMachineDiskModel) bool {
	for _, d := range list {
		if disksAreSame(disk, d) {
			return true
		}
	}
	return false
}

func changeValidatorDiskSize(_ context.Context, state *qemu.VirtualMachineResourceModel, plan *qemu.VirtualMachineResourceModel, resp *resource.ModifyPlanResponse) {
	for i, disk := range plan.Disks.Disks {
		previous := state.Disks.Disks[i]
		if !disksAreSame(previous, disk) {
			continue
		}
		if disk.Size.ValueInt64() < previous.Size.ValueInt64() {
			resp.Diagnostics.AddError("Disk size cannot be reduced", fmt.Sprintf("Disk %s%v size cannot be reduced from %d to %d", disk.InterfaceType.ValueString(), disk.Position.ValueInt64(), previous.Size.ValueInt64(), disk.Size.ValueInt64()))
			return
		}
	}
}

func changeValidatorDiskStorage(_ context.Context, state *qemu.VirtualMachineResourceModel, plan *qemu.VirtualMachineResourceModel, resp *resource.ModifyPlanResponse) {
	for i, disk := range plan.Disks.Disks {
		previous := state.Disks.Disks[i]
		if disksAreSame(previous, disk) {
			continue
		}
		diskName := fmt.Sprintf("%s%v", disk.InterfaceType.ValueString(), disk.Position.ValueInt64())
		if disk.Storage.ValueString() != previous.Storage.ValueString() {
			resp.Diagnostics.AddWarning(fmt.Sprintf("Changing disk storage for %s", diskName), fmt.Sprintf("Detected storage changed from %s to %s for disk %s. This will result in a new disk being created.", previous.Storage.ValueString(), disk.Storage.ValueString(), diskName))
			return
		}
	}
}

func changeValidatorDiskRemoved(_ context.Context, state *qemu.VirtualMachineResourceModel, plan *qemu.VirtualMachineResourceModel, resp *resource.ModifyPlanResponse) {
	removedDisks := []types.VirtualMachineDiskModel{}
	for _, disk := range state.Disks.Disks {
		if !isDiskInList(disk, plan.Disks.Disks) {
			removedDisks = append(removedDisks, disk)
		}
	}
	removedDisksName := []string{}
	for _, disk := range removedDisks {
		removedDisksName = append(removedDisksName, fmt.Sprintf("%s%v", disk.InterfaceType.ValueString(), disk.Position.ValueInt64()))
	}
	if len(removedDisks) > 0 {
		resp.Diagnostics.AddWarning("Detected disk(s) removal", fmt.Sprintf("Detected removal of disk(s) %v. This will result in the disk(s) being deleted.", removedDisksName))
	}
}

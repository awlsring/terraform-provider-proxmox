package vms

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func changeValidatorDiskSize(_ context.Context, state *qemu.VirtualMachineResourceModel, plan *qemu.VirtualMachineResourceModel, resp *resource.ModifyPlanResponse) {
	for i, disk := range plan.Disks.Disks {
		previous := state.Disks.Disks[i]
		if disk.Size.ValueInt64() < previous.Size.ValueInt64() {
			resp.Diagnostics.AddError("Disk size cannot be reduced", fmt.Sprintf("Disk %s%v size cannot be reduced from %d to %d", disk.InterfaceType.ValueString(), disk.Position.ValueInt64(), previous.Size.ValueInt64(), disk.Size.ValueInt64()))
			return
		}
	}
}

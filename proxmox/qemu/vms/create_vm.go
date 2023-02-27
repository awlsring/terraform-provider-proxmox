package vms

import (
	"context"
	"fmt"
	"time"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *virtualMachineResource) routeCreateVm(ctx context.Context, plan *qemu.VirtualMachineResourceModel) error {
	tflog.Debug(ctx, "route virtual machine creation method")
	switch true {
	case plan.Clone != nil:
		return r.clone(ctx, plan)
	case plan.ISO != nil:
		return r.iso(ctx, plan)
	default:
		tflog.Debug(ctx, "No valid init options provided")
		return fmt.Errorf("no valid init options provided")
	}
}

func (r *virtualMachineResource) clone(ctx context.Context, plan *qemu.VirtualMachineResourceModel) error {
	tflog.Debug(ctx, "clone virtual machine creation method")

	node := plan.Node.ValueString()
	vmId := int(plan.ID.ValueInt64())

	err := r.client.CloneVirtualMachine(ctx, &service.CloneVirtualMachineInput{
		Node:         node,
		VmId:         vmId,
		Source:       int(plan.Clone.Source.ValueInt64()),
		FullClone:    plan.Clone.FullClone.ValueBool(),
		Storage:      utils.OptionalToPointerString(plan.Clone.Storage.ValueString()),
		Description:  utils.OptionalToPointerString(plan.Description.ValueString()),
		Name:         utils.OptionalToPointerString(plan.Name.ValueString()),
		ResourcePool: utils.OptionalToPointerString(plan.ResourcePool.ValueString()),
	})
	if err != nil {
		tflog.Error(ctx, "clone recieved error: "+err.Error())
		return err
	}

	// wait till clone is complete
	r.waitForLock(ctx, node, vmId, r.timeouts.Clone)

	tflog.Debug(ctx, "clone virtual machine complete")
	return nil
}

func (r *virtualMachineResource) iso(ctx context.Context, plan *qemu.VirtualMachineResourceModel) error {
	tflog.Debug(ctx, "iso virtual machine creation method")

	err := r.client.CreateVirtualMachineIso(ctx, &service.CreateVirtualMachineIsoInput{
		Node:         plan.Node.ValueString(),
		VmId:         int(plan.ID.ValueInt64()),
		IsoStorage:   plan.ISO.Storage.ValueString(),
		IsoImage:     plan.ISO.Image.ValueString(),
		Name:         utils.OptionalToPointerString(plan.Name.ValueString()),
		Description:  utils.OptionalToPointerString(plan.Description.ValueString()),
		ResourcePool: utils.OptionalToPointerString(plan.ResourcePool.ValueString()),
	})
	if err != nil {
		return nil
	}

	return nil
}

func (r *virtualMachineResource) waitForLock(ctx context.Context, node string, vmId int, timeout int64) error {
	tflog.Debug(ctx, "waiting lock to release...")
	deadline := setDeadline(timeout)
	for {
		status, err := r.client.GetVirtualMachineStatus(ctx, node, vmId)
		if err != nil {
			tflog.Error(ctx, "error: "+err.Error())
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for lock to release")
			}
		} else if !status.HasLock() {
			break
		}
		tflog.Debug(ctx, "lock is still active, waiting 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	return nil
}

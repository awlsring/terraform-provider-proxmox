package vms

import (
	"context"
	"time"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *virtualMachineResource) waitForStateChange(ctx context.Context, node string, vmId int, endState proxmox.VirtualMachineStatus) error {
	tflog.Debug(ctx, "waiting for state change...")
	retries := 0
	limit := 10
	for {
		status, err := r.client.GetVirtualMachineStatus(ctx, node, vmId)
		if err != nil {
			return err
		}
		if status.Status == endState {
			break
		}
		if retries <= limit {
			retries++
		}
		tflog.Debug(ctx, "state is still "+string(status.Status)+", waiting 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	tflog.Debug(ctx, "state changed to "+string(endState))
	return nil
}

func (r *virtualMachineResource) startVm(ctx context.Context, node string, id int) error {
	tflog.Debug(ctx, "Starting virtual machine")
	err := r.client.StartVirtualMachine(ctx, node, id)
	if err != nil {
		return err
	}

	err = r.waitForStateChange(ctx, node, id, proxmox.VIRTUALMACHINESTATUS_RUNNING)
	if err != nil {
		return err
	}

	return nil
}

func (r *virtualMachineResource) stopVm(ctx context.Context, node string, id int) error {
	tflog.Debug(ctx, "Stopping virtual machine")
	err := r.client.StopVirtualMachine(ctx, node, id)
	if err != nil {
		return err
	}

	err = r.waitForStateChange(ctx, node, id, proxmox.VIRTUALMACHINESTATUS_STOPPED)
	if err != nil {
		return err
	}

	return nil
}

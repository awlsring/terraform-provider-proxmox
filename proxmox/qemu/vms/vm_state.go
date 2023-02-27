package vms

import (
	"context"
	"fmt"
	"time"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/r3labs/diff/v3"
)

var stateSensitiveProperties = []string{
	"bios",
	"cpu",
	"memory",
	"disks",
	"network_interfaces",
	"pci_devices",
	"cloud_init",
	"machine_type",
	"kvm_agruments",
}

func (r *virtualMachineResource) stopIfSensitivePropertyChanged(ctx context.Context, state *qemu.VirtualMachineResourceModel, plan *qemu.VirtualMachineResourceModel) (bool, error) {
	node := state.Node.ValueString()
	vmId := int(state.ID.ValueInt64())

	running := false
	status, err := r.client.GetVirtualMachineStatus(ctx, node, vmId)
	if err != nil {
		return false, err
	}
	if status.Status == proxmox.VIRTUALMACHINESTATUS_RUNNING {
		running = true
	}

	diff, err := diff.Diff(state, plan)
	if err != nil {
		tflog.Debug(ctx, "Error determining running state, defaulting to running")
		return false, nil
	}

	for _, d := range diff {
		field := d.Path[0]
		if utils.ListContains(stateSensitiveProperties, field) {
			tflog.Debug(ctx, fmt.Sprintf("Property '%v' is state sensitive, stopping VM", field))
			if running {
				err = r.stopVm(ctx, node, vmId)
				if err != nil {
					return false, err
				}
				return true, nil
			} else {
				return false, nil
			}
		}

	}
	return false, nil
}

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

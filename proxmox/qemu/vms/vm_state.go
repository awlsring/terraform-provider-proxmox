package vms

import (
	"context"
	"fmt"
	"time"

	"github.com/awlsring/proxmox-go/proxmox"
	vt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/vms/types"

	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/r3labs/diff/v3"
)

var stateSensitiveProperties = []string{
	"BIOS",
	"CPU",
	"Memory",
	"Disks",
	"NetworkInterfaces",
	"PCIDevices",
	"CloudInit",
	"MachineType",
	"KVMArguments",
}

func isSensitivePropertyChanged(ctx context.Context, state *vt.VirtualMachineResourceModel, plan *vt.VirtualMachineResourceModel) (bool, error) {
	diff, err := diff.Diff(state, plan)
	if err != nil {
		tflog.Debug(ctx, "Error determining running state, defaulting to running")
		return false, nil
	}

	for _, d := range diff {
		field := d.Path[0]
		tflog.Debug(ctx, fmt.Sprintf("Checking property '%v'", field))
		if utils.ListContains(stateSensitiveProperties, field) {
			tflog.Debug(ctx, fmt.Sprintf("Property '%v' is state sensitive", field))
			return true, nil
		}
	}
	return false, nil
}

func (r *virtualMachineResource) stopIfSensitivePropertyChanged(ctx context.Context, state *vt.VirtualMachineResourceModel, plan *vt.VirtualMachineResourceModel) (bool, error) {
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

	isSensitive, err := isSensitivePropertyChanged(ctx, state, plan)
	if err != nil {
		return false, err
	}

	if isSensitive {
		if running {
			tflog.Debug(ctx, "Property is state sensitive, stopping VM")
			err = r.stopVm(ctx, node, vmId)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (r *virtualMachineResource) waitForStateChange(ctx context.Context, node string, vmId int, timeout int64, endState proxmox.VirtualMachineStatus) error {
	tflog.Debug(ctx, "waiting for state change...")
	deadline := setDeadline(timeout)
	for {
		status, err := r.client.GetVirtualMachineStatus(ctx, node, vmId)
		if err != nil {
			return err
		}
		if status.Status == endState {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for state change")
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

	err = r.waitForStateChange(ctx, node, id, r.timeouts.Start, proxmox.VIRTUALMACHINESTATUS_RUNNING)
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

	err = r.waitForStateChange(ctx, node, id, r.timeouts.Stop, proxmox.VIRTUALMACHINESTATUS_STOPPED)
	if err != nil {
		return err
	}

	return nil
}

func (r *virtualMachineResource) deleteVm(ctx context.Context, node string, id int) error {
	tflog.Debug(ctx, "Deleting virtual machine")
	err := r.client.DeleteVirtualMachine(ctx, node, id)
	if err != nil {
		return err
	}
	deadline := setDeadline(r.timeouts.Delete)
	for {
		_, err := r.client.GetVirtualMachineStatus(ctx, node, id)
		if err != nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for state change")
		}
		tflog.Debug(ctx, "VM still exists, waiting 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	return nil
}

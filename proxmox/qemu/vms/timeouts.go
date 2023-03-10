package vms

import (
	"context"
	"time"

	vt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/vms/types"
)

type VirtualMachineTimeouts struct {
	Create     int64
	Delete     int64
	Stop       int64
	Start      int64
	Reboot     int64
	Shutdown   int64
	Clone      int64
	Configure  int64
	ResizeDisk int64
}

var timeoutDefaults = VirtualMachineTimeouts{
	Create:     600,
	Delete:     600,
	Stop:       600,
	Start:      600,
	Reboot:     600,
	Shutdown:   600,
	Clone:      600,
	Configure:  600,
	ResizeDisk: 600,
}

func loadTimeouts(ctx context.Context, timeouts *vt.VirtualMachineTerraformTimeouts) *VirtualMachineTimeouts {
	t := timeoutDefaults
	if timeouts == nil {
		return &t
	}

	if !timeouts.Create.IsNull() || !timeouts.Create.IsUnknown() {
		create := int64(timeouts.Create.ValueInt64())
		t.Create = create
	}

	if !timeouts.Delete.IsNull() || !timeouts.Delete.IsUnknown() {
		delete := int64(timeouts.Delete.ValueInt64())
		t.Delete = delete
	}

	if !timeouts.Stop.IsNull() || !timeouts.Stop.IsUnknown() {
		stop := int64(timeouts.Stop.ValueInt64())
		t.Stop = stop
	}

	if !timeouts.Start.IsNull() || !timeouts.Start.IsUnknown() {
		start := int64(timeouts.Start.ValueInt64())
		t.Start = start
	}

	if !timeouts.Reboot.IsNull() || !timeouts.Reboot.IsUnknown() {
		reboot := int64(timeouts.Reboot.ValueInt64())
		t.Reboot = reboot
	}

	if !timeouts.Shutdown.IsNull() || !timeouts.Shutdown.IsUnknown() {
		shutdown := int64(timeouts.Shutdown.ValueInt64())
		t.Shutdown = shutdown
	}

	if !timeouts.Clone.IsNull() || !timeouts.Clone.IsUnknown() {
		clone := int64(timeouts.Clone.ValueInt64())
		t.Clone = clone
	}

	if !timeouts.Configure.IsNull() || !timeouts.Configure.IsUnknown() {
		clone := int64(timeouts.Configure.ValueInt64())
		t.Clone = clone
	}

	if !timeouts.ResizeDisk.IsNull() || !timeouts.ResizeDisk.IsUnknown() {
		clone := int64(timeouts.ResizeDisk.ValueInt64())
		t.Clone = clone
	}

	return &t
}

func setDeadline(timeout int64) time.Time {
	now := time.Now().Unix() + timeout
	return time.Unix(now, 0)
}

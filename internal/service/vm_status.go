package service

import (
	"context"
	"strconv"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

func (c *Proxmox) GetVirtualMachineStatus(ctx context.Context, node string, vmid int) (*proxmox.VirtualMachineStatusSummary, error) {
	vmId := strconv.Itoa(vmid)
	request := c.client.GetVirtualMachineStatus(ctx, node, vmId)
	r, h, err := c.client.GetVirtualMachineStatusExecute(request)
	if err != nil {
		return nil, errors.ApiError(h, err)
	}

	return &r.Data, nil
}

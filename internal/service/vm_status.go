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

func (c *Proxmox) StartVirtualMachine(ctx context.Context, node string, vmid int) error {
	vmId := strconv.Itoa(vmid)
	request := c.client.StartVirtualMachine(ctx, node, vmId)
	_, h, err := c.client.StartVirtualMachineExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

func (c *Proxmox) StopVirtualMachine(ctx context.Context, node string, vmid int) error {
	vmId := strconv.Itoa(vmid)
	request := c.client.StopVirtualMachine(ctx, node, vmId)
	_, h, err := c.client.StopVirtualMachineExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

package service

import (
	"context"
	"strconv"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
)

type ResizeVirtualMachineDiskInput struct {
	Node string
	VmId int
	Disk string
	Size int64
}

func (c *Proxmox) ResizeVirtualMachineDisk(ctx context.Context, input *ResizeVirtualMachineDiskInput) error {
	vmId := strconv.Itoa(input.VmId)
	size := vm.BytesToStr(input.Size)
	request := c.client.ResizeVirtualMachineDisk(ctx, input.Node, vmId)
	request = request.Disk(proxmox.VirtualMachineDiskTarget(input.Disk))
	request = request.Size(size)
	h, err := c.client.ResizeVirtualMachineDiskExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

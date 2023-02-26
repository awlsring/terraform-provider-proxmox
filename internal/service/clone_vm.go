package service

import (
	"context"
	"strconv"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type CloneVirtualMachineInput struct {
	Node         string
	VmId         int
	Source       int
	FullClone    bool
	Storage      *string
	Description  *string
	Name         *string
	ResourcePool *string
}

func (c *Proxmox) CloneVirtualMachine(ctx context.Context, input *CloneVirtualMachineInput) error {
	vmId := float32(input.VmId)
	sourceId := strconv.Itoa(input.Source)
	content := proxmox.CloneVirtualMachineRequestContent{
		Newid:       vmId,
		Full:        &input.FullClone,
		Pool:        input.ResourcePool,
		Storage:     input.Storage,
		Description: input.Description,
		Name:        input.Name,
	}
	request := c.client.CloneVirtualMachine(ctx, input.Node, sourceId)
	request = request.CloneVirtualMachineRequestContent(content)
	_, h, err := c.client.CloneVirtualMachineExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

type CreateVirtualMachineIsoInput struct {
	Node         string
	VmId         int
	IsoStorage   string
	IsoImage     string
	Name         *string
	Description  *string
	ResourcePool *string
}

func formIsoString(storage string, image string) *string {
	isoStr := storage + ":iso/" + image + ",media=cdrom"
	return &isoStr
}

func (c *Proxmox) CreateVirtualMachineIso(ctx context.Context, input *CreateVirtualMachineIsoInput) error {
	vmId := strconv.Itoa(input.VmId)
	content := proxmox.CreateVirtualMachineRequestContent{
		Vmid: vmId,
		Name: input.Name,
		Ide2: formIsoString(input.IsoStorage, input.IsoImage),
		Pool: input.ResourcePool,
	}
	request := c.client.CreateVirtualMachine(ctx, input.Node)
	request = request.CreateVirtualMachineRequestContent(content)
	_, h, err := c.client.CreateVirtualMachineExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

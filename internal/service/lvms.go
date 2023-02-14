package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type LVM struct {
	Name   string
	Node   string
	Size   int64
	Device string
}

func (c *Proxmox) ListLVMs(ctx context.Context, node string) ([]LVM, error) {
	request := c.client.ListLVMs(ctx, node)
	resp, h, err := c.client.ListLVMsExecute(request)
	if err != nil {
		return nil, errors.ApiError(h, err)
	}

	lvms := []LVM{}
	for _, c := range resp.Data.Children {
		var d string
		if c.HasChildren() {
			for _, ch := range c.Children {
				if isDiskChild(ch) {
					d = *ch.Name
				}
			}
		}

		lvm := LVM{
			Name:   PtrStringToString(c.Name),
			Node:   node,
			Size:   PtrFloatToInt64(c.Size),
			Device: d,
		}
		lvms = append(lvms, lvm)
	}

	return lvms, nil
}

func isDiskChild(c proxmox.LVMChild) bool {
	if c.Name == nil {
		return false
	}
	if strings.Contains(*c.Name, "/dev/") {
		return true
	}
	return false
}

func (c *Proxmox) GetLVM(ctx context.Context, node string, pool string) (*LVM, error) {
	pools, err := c.ListLVMs(ctx, node)
	if err != nil {
		return nil, err
	}

	for _, p := range pools {
		if p.Name == pool {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("LVM thinpool not found")
}

type CreateLVMInput struct {
	Node   string
	Name   string
	Device string
}

func (c *Proxmox) CreateLVM(ctx context.Context, input *CreateLVMInput) error {
	request := c.client.CreateLVM(ctx, input.Node)
	addStorage := float32(0)
	request = request.CreateLVMRequestContent(proxmox.CreateLVMRequestContent{
		Device:     input.Device,
		Name:       input.Name,
		AddStorage: &addStorage,
	})
	_, h, err := c.client.CreateLVMExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

func (c *Proxmox) DeleteLVM(ctx context.Context, node string, pool string) error {
	request := c.client.DeleteLVM(ctx, node, pool)
	request = request.CleanupDisks(1)
	_, h, err := c.client.DeleteLVMExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

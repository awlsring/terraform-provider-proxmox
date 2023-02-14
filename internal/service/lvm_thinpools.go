package service

import (
	"context"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type LVMThinpool struct {
	Name         string
	Node         string
	Size         int64
	MetadataSize int64
	VolumeGroup  string
	Device       string
}

func (c *Proxmox) ListLVMThinpools(ctx context.Context, node string) ([]LVMThinpool, error) {
	request := c.client.ListLVMThins(ctx, node)
	pools, h, err := c.client.ListLVMThinsExecute(request)
	if err != nil {
		return nil, errors.ApiError(h, err)
	}

	lvms, err := c.ListLVMs(ctx, node)
	if err != nil {
		return nil, err
	}

	mapping := map[string]LVM{}
	for _, l := range lvms {
		mapping[l.Name] = l
	}

	thins := []LVMThinpool{}
	for _, p := range pools.Data {

		var vg LVM
		if v, ok := mapping[p.Vg]; ok {
			vg = v
		}

		thin := LVMThinpool{
			Name:         p.Lv,
			Node:         node,
			Size:         int64(p.LvSize),
			MetadataSize: int64(p.MetadataSize),
			VolumeGroup:  p.Vg,
			Device:       vg.Device,
		}
		thins = append(thins, thin)
	}

	return thins, nil
}

func (c *Proxmox) GetLVMThinpool(ctx context.Context, node string, pool string) (*LVMThinpool, error) {
	pools, err := c.ListLVMThinpools(ctx, node)
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

type CreateLVMThinpoolInput struct {
	Node   string
	Name   string
	Device string
}

func (c *Proxmox) CreateLVMThinpool(ctx context.Context, input *CreateLVMThinpoolInput) error {
	request := c.client.CreateLVMThin(ctx, input.Node)
	addStorage := float32(0)
	request = request.CreateLVMThinRequestContent(proxmox.CreateLVMThinRequestContent{
		Device:     input.Device,
		Name:       input.Name,
		AddStorage: &addStorage,
	})
	_, h, err := c.client.CreateLVMThinExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

func (c *Proxmox) DeleteLVMThinpool(ctx context.Context, node string, pool string, vg string) error {
	request := c.client.DeleteLVMThin(ctx, node, pool)
	request = request.VolumeGroup(vg)
	request = request.CleanupDisks(1)
	_, h, err := c.client.DeleteLVMThinExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

package service

import (
	"context"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

func (c *Proxmox) ListZFSPools(ctx context.Context, node string) ([]proxmox.ZFSPoolSummary, error) {
	request := c.client.ListZFSPools(ctx, node)
	pools, h, err := c.client.ListZFSPoolsExecute(request)
	if err != nil {
		return nil, errors.ApiError(h, err)
	}
	return pools.Data, nil
}

func (c *Proxmox) GetZFSPool(ctx context.Context, node string, pool string) (proxmox.ZFSPoolSummary, error) {
	pools, err := c.ListZFSPools(ctx, node)
	if err != nil {
		return proxmox.ZFSPoolSummary{}, err
	}
	for _, p := range pools {
		if p.Name == pool {
			return p, nil
		}
	}
	return proxmox.ZFSPoolSummary{}, fmt.Errorf("ZFS pool %s not found", pool)
}

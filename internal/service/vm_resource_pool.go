package service

import (
	"context"
	"strconv"

	"github.com/awlsring/proxmox-go/proxmox"
)

func (c *Proxmox) DetermineVirtualMachineResourcePool(ctx context.Context, vmId int) (bool, string, error) {
	pools, err := c.DescribePools(ctx)
	if err != nil {
		return false, "", err
	}

	for _, pool := range pools {
		for _, member := range pool.Members {
			if member.Type != proxmox.POOLMEMBERTYPE_QEMU {
				continue
			}
			if member.Id == strconv.Itoa(vmId) {
				return true, pool.Id, nil
			}
		}
	}
	return false, "", nil
}

func (c *Proxmox) AddVirtualMachineToResourcePool(ctx context.Context, vmId int, pool string) error {
	err := c.UpdatePool(ctx, &UpdatePoolInput{
		PoolId: pool,
		Vms:    []int{vmId},
	})
	if err != nil {
		return err
	}
	return nil
}
func (c *Proxmox) RemoveVirtualMachineFromResourcePool(ctx context.Context, vmId int, pool string) error {
	err := c.UpdatePool(ctx, &UpdatePoolInput{
		PoolId: pool,
		Delete: true,
		Vms:    []int{vmId},
	})
	if err != nil {
		return err
	}
	return nil
}

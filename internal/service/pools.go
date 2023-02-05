package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
)

type Pool struct {
	Id string
	Comment string
	Members []PoolMember
}

type PoolMember struct {
	Id string
	Type proxmox.PoolMemberType
}

func (c *Proxmox) DescribePools(ctx context.Context) ([]Pool, error) {
	pool, err := c.ListPools(ctx)
	if err != nil {
		return nil, err
	}

	poolsSummaries := []Pool{}
	for _, poolSummary := range pool {
		
		poolConfiguration, err := c.GetPool(ctx, poolSummary.Poolid)
		if err != nil {
			return nil, err
		}

		members := []PoolMember{}

		for _, member := range poolConfiguration.Members {
			members = append(members, PoolMember{
				Id: PtrStringToString(member.Id),
				Type: *member.Type,
			})
		}

		poolsSummaries = append(poolsSummaries, Pool{
			Id: poolSummary.Poolid,
			Comment: PtrStringToString(poolSummary.Comment),
			Members: members,
		})
	}

	return poolsSummaries, nil
}

func (c *Proxmox) ListPools(ctx context.Context) ([]proxmox.PoolSummary, error) {
	request := c.client.ListPools(ctx)
	resp, _, err := c.client.ListPoolsExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Proxmox) GetPool(ctx context.Context, poolId string) (*proxmox.PoolConfigurationSummary, error) {
	request := c.client.GetPool(ctx, poolId)
	resp, _, err := c.client.GetPoolExecute(request)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}
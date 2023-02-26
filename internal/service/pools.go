package service

import (
	"context"
	"strconv"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type Pool struct {
	Id      string
	Comment string
	Members []PoolMember
}

type PoolMember struct {
	Id   string
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
			var id string
			switch member.Type {
			case proxmox.POOLMEMBERTYPE_QEMU:
				id = strconv.Itoa(int(*member.Vmid))
			case proxmox.POOLMEMBERTYPE_STORAGE:
				id = *member.Storage
			}

			members = append(members, PoolMember{
				Id:   id,
				Type: member.Type,
			})
		}

		poolsSummaries = append(poolsSummaries, Pool{
			Id:      poolSummary.Poolid,
			Comment: PtrStringToString(poolSummary.Comment),
			Members: members,
		})
	}

	return poolsSummaries, nil
}

func (c *Proxmox) ListPools(ctx context.Context) ([]proxmox.PoolSummary, error) {
	request := c.client.ListPools(ctx)
	resp, h, err := c.client.ListPoolsExecute(request)
	if err != nil {
		return nil, errors.ApiError(h, err)
	}

	return resp.Data, nil
}

func (c *Proxmox) GetPool(ctx context.Context, poolId string) (*proxmox.PoolConfigurationSummary, error) {
	request := c.client.GetPool(ctx, poolId)
	resp, h, err := c.client.GetPoolExecute(request)
	if err != nil {
		return nil, errors.ApiError(h, err)
	}

	return &resp.Data, nil
}

type CreatePoolInput struct {
	PoolId  string
	Comment *string
}

func (c *Proxmox) CreatePool(ctx context.Context, input *CreatePoolInput) error {
	request := c.client.CreatePool(ctx)
	request = request.CreatePoolRequestContent(
		proxmox.CreatePoolRequestContent{
			Poolid:  input.PoolId,
			Comment: input.Comment,
		},
	)
	h, err := c.client.CreatePoolExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

func (c *Proxmox) DeletePool(ctx context.Context, poolId string) error {
	request := c.client.DeletePool(ctx, poolId)
	h, err := c.client.DeletePoolExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

type UpdatePoolInput struct {
	PoolId  string
	Comment *string
	Delete  bool
	Storage []string
	Vms     []int
}

func (c *Proxmox) UpdatePool(ctx context.Context, input *UpdatePoolInput) error {
	request := c.client.ModifyPool(ctx, input.PoolId)

	content := proxmox.ModifyPoolRequestContent{
		Comment: input.Comment,
	}
	if input.Delete {
		content.Delete = &input.Delete
	}
	if len(input.Storage) != 0 {
		conv := SliceToStringCommaList(input.Storage)
		content.Storage = &conv
	}
	if len(input.Vms) != 0 {
		ids := []string{}
		for _, vm := range input.Vms {
			ids = append(ids, strconv.Itoa(vm))
		}
		conv := SliceToStringCommaList(ids)
		content.Vms = &conv
	}

	request = request.ModifyPoolRequestContent(content)

	h, err := c.client.ModifyPoolExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

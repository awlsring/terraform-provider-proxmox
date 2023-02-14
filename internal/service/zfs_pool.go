package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type ZFSPool struct {
	Name      string
	Node      string
	Size      int64
	Health    string
	Disks     []string
	RaidLevel proxmox.ZFSRaidLevel
}

func (c *Proxmox) ListZFSPools(ctx context.Context, node string) ([]proxmox.ZFSPoolSummary, error) {
	request := c.client.ListZFSPools(ctx, node)
	pools, h, err := c.client.ListZFSPoolsExecute(request)
	if err != nil {
		return nil, errors.ApiError(h, err)
	}
	return pools.Data, nil
}

func (c *Proxmox) GetZFSPoolStatus(ctx context.Context, node string, pool string) (proxmox.ZFSPoolStatusSummary, error) {
	request := c.client.GetZFSPoolStatus(ctx, node, pool)
	resp, h, err := c.client.GetZFSPoolStatusExecute(request)
	if err != nil {
		return proxmox.ZFSPoolStatusSummary{}, errors.ApiError(h, err)
	}
	return resp.Data, nil
}

// Describes a ZFS pool and the disks that makes it up
// This chains several API calls togther to build a ZFSPool struct
//
// Process flow is:
//  1. Get a list of all ZFS pools
//  2. Get the status of each pool
//  3. Get a list of all disks on node
//  4. Match the disks to the pool
//  5. Build a ZFSPool struct and append to a list
//  6. Return a list of ZFSPool structs
func (c *Proxmox) DescribeZFSPools(ctx context.Context, node string) ([]ZFSPool, error) {
	zfsPools := []ZFSPool{}

	pools, err := c.ListZFSPools(ctx, node)
	if err != nil {
		return nil, err
	}

	for _, p := range pools {
		status, err := c.GetZFSPoolStatus(ctx, node, p.Name)
		if err != nil {
			return nil, err
		}

		links := []string{}
		for _, d := range status.Children {
			links = append(links, recursiveNameCheck(d.Children)...)
		}

		disks, err := c.ListDisks(ctx, node)
		if err != nil {
			return nil, err
		}

		poolDisks := []string{}
		for _, d := range disks {
			for _, l := range links {
				if strings.Contains(l, d.IDLink) {
					poolDisks = append(poolDisks, d.Device)
				}
			}
		}

		zfsPools = append(zfsPools, ZFSPool{
			Name:      p.Name,
			Node:      node,
			Size:      int64(p.Size),
			Health:    p.Health,
			Disks:     poolDisks,
			RaidLevel: determineRaidLevel(poolDisks),
		})
	}
	return zfsPools, nil
}

// An unefficient method of getting a single ZFS pool
func (c *Proxmox) DescribeZFSPool(ctx context.Context, node string, pool string) (*ZFSPool, error) {
	zfsPools, err := c.DescribeZFSPools(ctx, node)
	if err != nil {
		return nil, err
	}

	for _, p := range zfsPools {
		if p.Name == pool {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("ZFS pool not found")
}

type CreateZFSPoolInput struct {
	Node      string
	Name      string
	Disks     []string
	RaidLevel proxmox.ZFSRaidLevel
}

func (c *Proxmox) CreateZFSPool(ctx context.Context, input *CreateZFSPoolInput) error {
	request := c.client.CreateZFSPool(ctx, input.Node)
	addStorage := float32(0)
	request = request.CreateZFSPoolRequestContent(proxmox.CreateZFSPoolRequestContent{
		Devices:    SliceToStringCommaList(input.Disks),
		Name:       input.Name,
		Raidlevel:  input.RaidLevel,
		AddStorage: &addStorage,
	})
	_, h, err := c.client.CreateZFSPoolExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

func (c *Proxmox) DeleteZFSPool(ctx context.Context, node string, pool string) error {
	request := c.client.DeleteZFSPool(ctx, node, pool)
	request = request.CleanupDisks(1)
	_, h, err := c.client.DeleteZFSPoolExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

func recursiveNameCheck(children []proxmox.ZFSPoolStatusChild) []string {
	names := []string{}
	for _, child := range children {
		if strings.Contains(child.Name, "/dev/disk/by-id/") {
			names = append(names, child.Name)
		}
		if child.HasChildren() {
			names = append(names, recursiveNameCheck(child.Children)...)
		}
	}
	return names
}

func determineRaidLevel(disks []string) proxmox.ZFSRaidLevel {
	switch len(disks) {
	case 1:
		return proxmox.ZFSRAIDLEVEL_SINGLE
	case 2:
		return proxmox.ZFSRAIDLEVEL_MIRROR
	case 3:
		return proxmox.ZFSRAIDLEVEL_RAIDZ
	case 4:
		// could also be Raid10?
		return proxmox.ZFSRAIDLEVEL_RAIDZ2
	case 5:
		return proxmox.ZFSRAIDLEVEL_RAIDZ3
	default:
		return proxmox.ZFSRAIDLEVEL_SINGLE
	}
}

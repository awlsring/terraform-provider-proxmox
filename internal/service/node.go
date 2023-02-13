package service

import (
	"context"
	"errors"

	"github.com/awlsring/proxmox-go/proxmox"
)

type Node struct {
	Id                string
	Node              string
	Cores             int
	SslFingerprint    string
	Memory            int64
	DiskSpace         int64
	Disks             []Disk
	NetworkInterfaces []NetworkInterface
}

type Disk struct {
	Device string
	IDLink string
	Size   int64
	Model  string
	Serial string
	Vendor string
	Type   proxmox.DiskType
}

type NetworkInterface struct {
	Name string
}

func (c *Proxmox) DescribeNode(ctx context.Context, node string) (Node, error) {
	nodeSummary, err := c.GetNode(ctx, node)
	if err != nil {
		return Node{}, err
	}

	disks, err := c.ListDisks(ctx, node)
	if err != nil {
		return Node{}, err
	}

	networkInterfaces, err := c.ListNetworkInterfaces(ctx, node)
	if err != nil {
		return Node{}, err
	}

	return Node{
		Id:                *nodeSummary.Id,
		Node:              nodeSummary.Node,
		Cores:             int(*nodeSummary.Maxcpu),
		SslFingerprint:    *nodeSummary.SslFingerprint,
		Memory:            int64(*nodeSummary.Maxmem),
		DiskSpace:         int64(*nodeSummary.Maxdisk),
		Disks:             disks,
		NetworkInterfaces: networkInterfaces,
	}, nil
}

func (c *Proxmox) GetNode(ctx context.Context, node string) (*proxmox.NodeSummary, error) {
	nodes, err := c.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	for _, n := range nodes {
		if n.Node == node {
			return &n, nil
		}
	}
	return nil, errors.New("node not found")
}

func (c *Proxmox) ListNodes(ctx context.Context) ([]proxmox.NodeSummary, error) {
	request := c.client.ListNodes(ctx)
	resp, _, err := c.client.ListNodesExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Proxmox) listNodesNames(ctx context.Context) ([]string, error) {
	request := c.client.ListNodes(ctx)
	resp, _, err := c.client.ListNodesExecute(request)
	if err != nil {
		return nil, err
	}

	nodes := make([]string, len(resp.Data))
	for i, node := range resp.Data {
		nodes[i] = node.Node
	}
	return nodes, nil
}

func (c *Proxmox) ListDisks(ctx context.Context, node string) ([]Disk, error) {
	request := c.client.ListDisks(ctx, node)
	request = request.IncludePartitions(0)
	resp, _, err := c.client.ListDisksExecute(request)
	if err != nil {
		return nil, err
	}

	disks := make([]Disk, len(resp.Data))
	for i, disk := range resp.Data {
		disks[i] = Disk{
			Device: disk.Devpath,
			IDLink: PtrStringToString(disk.ByIdLink),
			Size:   int64(disk.Size),
			Model:  *disk.Model,
			Serial: *disk.Serial,
			Vendor: *disk.Vendor,
			Type:   *disk.Type,
		}
	}

	return disks, nil
}

func (c *Proxmox) ListNetworkInterfaces(ctx context.Context, node string) ([]NetworkInterface, error) {
	request := c.client.ListNetworkInterfaces(ctx, node)
	request = request.Type_(proxmox.NETWORKINTERFACETYPE_ETH)
	resp, _, err := c.client.ListNetworkInterfacesExecute(request)
	if err != nil {
		return nil, err
	}

	networkInterfaces := make([]NetworkInterface, len(resp.Data))
	for i, networkInterface := range resp.Data {

		networkInterfaces[i] = NetworkInterface{
			Name: networkInterface.Iface,
		}
	}

	return networkInterfaces, nil
}

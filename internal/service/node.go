package service

import (
	"context"
	"errors"

	"github.com/awlsring/proxmox-go/proxmox"
)

func (c *Proxmox) GetNode(ctx context.Context, node string) (*proxmox.NodeSummary, error){
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
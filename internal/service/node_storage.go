package service

import (
	"context"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
)

func (c *Proxmox) GetNodeStorage(ctx context.Context, node string, storage string) (*proxmox.NodeStorageSummary, error) {
	request := c.client.ListNodeStorage(ctx, node)
	request = request.Storage(storage)
	resp, _, err := c.client.ListNodeStorageExecute(request)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) != 1 {
		return nil, fmt.Errorf("expected 1 storage, got %d", len(resp.Data))
	}

	return &resp.Data[0], nil
}

func (c *Proxmox) ListNodeStorage(ctx context.Context, node string) ([]proxmox.NodeStorageSummary, error) {
	request := c.client.ListNodeStorage(ctx, node)
	request.Enabled(1)
	resp, _, err := c.client.ListNodeStorageExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

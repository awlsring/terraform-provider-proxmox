package service

import (
	"context"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
)

type ZFSNodeStorage struct {
	Id           string
	Node         string
	Storage      string
	ContentTypes []string
	Size         int64
	ZFSPool      string
	Mount        string
}

func (c *Proxmox) DescribeZFSNodeStorage(ctx context.Context, node string) ([]*ZFSNodeStorage, error) {
	nodeStorage, err := c.ListNodeStorage(ctx, node)
	if err != nil {
		return nil, err
	}

	storageList := []*ZFSNodeStorage{}
	for _, s := range nodeStorage {
		if s.Type != proxmox.STORAGETYPE_ZFSPOOL || s.Enabled == nil {
			continue
		}
		if *s.Enabled != 1 {
			continue
		}

		s := &ZFSNodeStorage{
			Id:           s.Storage,
			Node:         node,
			Storage:      s.Storage,
			Size:         PtrFloatToInt64(s.Total),
			ContentTypes: StringCommaListToSlice(s.Content),
		}
		storageList = append(storageList, s)
	}

	for _, s := range storageList {
		storage, err := c.GetZFSStorageClass(ctx, s.Storage)
		if err != nil {
			return nil, err
		}
		s.ZFSPool = PtrStringToString(&storage.ZFSPool)
		s.Mount = storage.Mount
	}

	return storageList, nil
}

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

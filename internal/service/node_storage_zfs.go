package service

import (
	"context"

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

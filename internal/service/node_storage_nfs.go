package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
)

type NFSNodeStorage struct {
	Id           string
	Node         string
	Storage      string
	ContentTypes []string
	Size         int64
	Server       string
	Mount        string
	Export       string
}

func (c *Proxmox) DescribeNFSNodeStorage(ctx context.Context, node string) ([]*NFSNodeStorage, error) {
	nodeStorage, err := c.ListNodeStorage(ctx, node)
	if err != nil {
		return nil, err
	}

	storageList := []*NFSNodeStorage{}
	for _, s := range nodeStorage {
		if s.Type != proxmox.STORAGETYPE_NFS || s.Enabled == nil {
			continue
		}
		if *s.Enabled != 1 {
			continue
		}

		s := &NFSNodeStorage{
			Id:           s.Storage,
			Node:         node,
			Storage:      s.Storage,
			Size:         PtrFloatToInt64(s.Total),
			ContentTypes: StringCommaListToSlice(s.Content),
		}
		storageList = append(storageList, s)
	}

	for _, s := range storageList {
		storage, err := c.GetNFSStorageClass(ctx, s.Storage)
		if err != nil {
			return nil, err
		}
		s.Server = storage.Server
		s.Export = storage.Export
		s.Mount = storage.Mount
	}

	return storageList, nil
}

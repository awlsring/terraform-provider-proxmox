package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
)

type LVMNodeStorage struct {
	Id           string
	Node         string
	Storage      string
	ContentTypes []string
	Size         int64
	VolumeGroup  string
}

func (c *Proxmox) DescribeLVMNodeStorage(ctx context.Context, node string) ([]*LVMNodeStorage, error) {
	nodeStorage, err := c.ListNodeStorage(ctx, node)
	if err != nil {
		return nil, err
	}

	storageList := []*LVMNodeStorage{}
	for _, s := range nodeStorage {
		if s.Type != proxmox.STORAGETYPE_LVM || s.Enabled == nil {
			continue
		}
		if *s.Enabled != 1 {
			continue
		}

		s := &LVMNodeStorage{
			Id:           s.Storage,
			Node:         node,
			Storage:      s.Storage,
			Size:         PtrFloatToInt64(s.Total),
			ContentTypes: StringCommaListToSlice(s.Content),
		}
		storageList = append(storageList, s)
	}

	for _, s := range storageList {
		storage, err := c.GetLVMStorageClass(ctx, s.Storage)
		if err != nil {
			return nil, err
		}
		s.VolumeGroup = storage.VolumeGroup
	}

	return storageList, nil
}

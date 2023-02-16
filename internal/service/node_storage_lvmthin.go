package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
)

type LVMThinNodeStorage struct {
	Id           string
	Node         string
	Storage      string
	ContentTypes []string
	Size         int64
	VolumeGroup  string
	Thinpool     string
}

func (c *Proxmox) DescribeLVMThinNodeStorage(ctx context.Context, node string) ([]*LVMThinNodeStorage, error) {
	nodeStorage, err := c.ListNodeStorage(ctx, node)
	if err != nil {
		return nil, err
	}

	storageList := []*LVMThinNodeStorage{}
	for _, s := range nodeStorage {
		if s.Type != proxmox.STORAGETYPE_LVMTHIN || s.Enabled == nil {
			continue
		}
		if *s.Enabled != 1 {
			continue
		}

		s := &LVMThinNodeStorage{
			Id:           s.Storage,
			Node:         node,
			Storage:      s.Storage,
			Size:         PtrFloatToInt64(s.Total),
			ContentTypes: StringCommaListToSlice(s.Content),
		}
		storageList = append(storageList, s)
	}

	for _, s := range storageList {
		storage, err := c.GetLVMThinStorageClass(ctx, s.Storage)
		if err != nil {
			return nil, err
		}
		s.VolumeGroup = storage.VolumeGroup
		s.Thinpool = storage.Thinpool
	}

	return storageList, nil
}

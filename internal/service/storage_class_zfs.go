package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
)

type ZFSStorageClass struct {
	Id      string
	Nodes   []string
	Content []string
	ZFSPool string
	Mount   string
}

func (c *Proxmox) ListZFSStorageClasses(ctx context.Context) ([]ZFSStorageClass, error) {
	storage, err := c.listStorageOfType(ctx, proxmox.STORAGETYPE_ZFSPOOL)
	if err != nil {
		return nil, err
	}

	allNodes := []string{}
	storageList := []ZFSStorageClass{}
	for _, s := range storage {
		var nodes []string
		if s.Nodes != nil {
			nodes = StringCommaPtrListToSlice(s.Nodes)
		} else {
			if len(allNodes) == 0 {
				allNodes, err = c.listNodesNames(ctx)
				if err != nil {
					return nil, err
				}
			}
			nodes = allNodes
		}

		s := ZFSStorageClass{
			Id:      s.Storage,
			Nodes:   nodes,
			Content: StringCommaPtrListToSlice(s.Content),
			ZFSPool: PtrStringToString(s.Pool),
			Mount:   *s.Mountpoint,
		}
		storageList = append(storageList, s)
	}

	return storageList, nil
}

func (c *Proxmox) GetZFSStorageClass(ctx context.Context, name string) (*ZFSStorageClass, error) {
	storage, err := c.GetStorageClass(ctx, name)
	if err != nil {
		return nil, err
	}

	var nodes []string
	if storage.Nodes != nil {
		nodes = StringCommaPtrListToSlice(storage.Nodes)
	} else {
		nodes, err = c.listNodesNames(ctx)
		if err != nil {
			return nil, err
		}
	}

	s := ZFSStorageClass{
		Id:      storage.Storage,
		Nodes:   nodes,
		Content: StringCommaPtrListToSlice(storage.Content),
		ZFSPool: PtrStringToString(storage.Pool),
		Mount:   *storage.Mountpoint,
	}

	return &s, nil
}

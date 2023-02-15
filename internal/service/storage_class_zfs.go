package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
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
				allNodes, err = c.ListNodesNames(ctx)
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
		nodes, err = c.ListNodesNames(ctx)
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

type CreateZFSStorageClassInput struct {
	Id           string
	Nodes        []string
	ContentTypes []string
	Pool         string
}

func (c *Proxmox) CreateZFSStorageClass(ctx context.Context, input *CreateZFSStorageClassInput) error {
	request := c.client.CreateStorage(ctx)
	content := proxmox.CreateStorageRequestContent{
		Storage: input.Id,
		Type:    proxmox.STORAGETYPE_ZFSPOOL,
		Content: SliceToStringCommaListPtr(input.ContentTypes),
		Pool:    &input.Pool,
	}
	request = request.CreateStorageRequestContent(content)

	_, h, err := c.client.CreateStorageExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

func (c *Proxmox) DeleteZFSSStorageClass(ctx context.Context, name string) error {
	request := c.client.DeleteStorage(ctx, name)
	h, err := c.client.DeleteStorageExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

type ModifyZFSStorageClassInput struct {
	Nodes        []string
	ContentTypes []string
}

func (c *Proxmox) ModifyZFSStorageClass(ctx context.Context, name string, input *ModifyZFSStorageClassInput) error {
	request := c.client.ModifyStorage(ctx, name)
	content := proxmox.ModifyStorageRequestContent{
		Nodes:   SliceToStringCommaListPtr(input.Nodes),
		Content: SliceToStringCommaListPtr(input.ContentTypes),
	}
	request = request.ModifyStorageRequestContent(content)

	_, h, err := c.client.ModifyStorageExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

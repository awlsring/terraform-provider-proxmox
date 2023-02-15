package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type NFSStorageClass struct {
	Id      string
	Server  string
	Nodes   []string
	Content []string
	Mount   string
	Export  string
}

func (c *Proxmox) ListNFSStorageClasses(ctx context.Context) ([]NFSStorageClass, error) {
	storage, err := c.listStorageOfType(ctx, proxmox.STORAGETYPE_NFS)
	if err != nil {
		return nil, err
	}

	allNodes := []string{}
	storageList := []NFSStorageClass{}
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

		s := NFSStorageClass{
			Id:      s.Storage,
			Server:  *s.Server,
			Nodes:   nodes,
			Content: StringCommaPtrListToSlice(s.Content),
			Mount:   *s.Path,
			Export:  *s.Export,
		}
		storageList = append(storageList, s)
	}

	return storageList, nil
}

func (c *Proxmox) GetNFSStorageClass(ctx context.Context, name string) (*NFSStorageClass, error) {
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

	s := NFSStorageClass{
		Id:      storage.Storage,
		Server:  *storage.Server,
		Nodes:   nodes,
		Content: StringCommaPtrListToSlice(storage.Content),
		Mount:   *storage.Path,
		Export:  *storage.Export,
	}

	return &s, nil
}

type CreateNFSStorageClassInput struct {
	Id           string
	Server       string
	Export       string
	Nodes        []string
	ContentTypes []string
}

func (c *Proxmox) CreateNFSStorageClass(ctx context.Context, input *CreateNFSStorageClassInput) error {
	request := c.client.CreateStorage(ctx)
	content := proxmox.CreateStorageRequestContent{
		Storage: input.Id,
		Server:  &input.Server,
		Type:    proxmox.STORAGETYPE_NFS,
		Export:  &input.Export,
		Content: SliceToStringCommaListPtr(input.ContentTypes),
		Nodes:   SliceToStringCommaListPtr(input.Nodes),
	}
	request = request.CreateStorageRequestContent(content)

	_, h, err := c.client.CreateStorageExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

func (c *Proxmox) DeleteNFSSStorageClass(ctx context.Context, name string) error {
	request := c.client.DeleteStorage(ctx, name)
	h, err := c.client.DeleteStorageExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

type ModifyNFSStorageClassInput struct {
	Nodes        []string
	ContentTypes []string
}

func (c *Proxmox) ModifyNFSStorageClass(ctx context.Context, name string, input *ModifyNFSStorageClassInput) error {
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

package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type LVMThinStorageClass struct {
	Id          string
	VolumeGroup string
	Thinpool    string
	Nodes       []string
	Content     []string
}

func (c *Proxmox) ListLVMThinStorageClasses(ctx context.Context) ([]LVMThinStorageClass, error) {
	storage, err := c.listStorageOfType(ctx, proxmox.STORAGETYPE_LVMTHIN)
	if err != nil {
		return nil, err
	}

	allNodes := []string{}
	storageList := []LVMThinStorageClass{}
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

		s := LVMThinStorageClass{
			Id:          s.Storage,
			VolumeGroup: *s.Vgname,
			Thinpool:    *s.Thinpool,
			Nodes:       nodes,
			Content:     StringCommaPtrListToSlice(s.Content),
		}
		storageList = append(storageList, s)
	}

	return storageList, nil
}

func (c *Proxmox) GetLVMThinStorageClass(ctx context.Context, name string) (*LVMThinStorageClass, error) {
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

	s := LVMThinStorageClass{
		Id:          storage.Storage,
		VolumeGroup: *storage.Vgname,
		Thinpool:    *storage.Thinpool,
		Nodes:       nodes,
		Content:     StringCommaPtrListToSlice(storage.Content),
	}

	return &s, nil
}

type CreateLVMThinStorageClassInput struct {
	Id           string
	VolumeGroup  string
	Thinpool     string
	Nodes        []string
	ContentTypes []string
}

func (c *Proxmox) CreateLVMThinStorageClass(ctx context.Context, input *CreateLVMThinStorageClassInput) error {
	request := c.client.CreateStorage(ctx)
	content := proxmox.CreateStorageRequestContent{
		Storage:  input.Id,
		Vgname:   &input.VolumeGroup,
		Type:     proxmox.STORAGETYPE_LVMTHIN,
		Thinpool: &input.Thinpool,
		Content:  SliceToStringCommaListPtr(input.ContentTypes),
		Nodes:    SliceToStringCommaListPtr(input.Nodes),
	}
	request = request.CreateStorageRequestContent(content)

	_, h, err := c.client.CreateStorageExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

func (c *Proxmox) DeleteLVMThinSStorageClass(ctx context.Context, name string) error {
	request := c.client.DeleteStorage(ctx, name)
	h, err := c.client.DeleteStorageExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

type ModifyLVMThinStorageClassInput struct {
	Nodes        []string
	ContentTypes []string
}

func (c *Proxmox) ModifyLVMThinStorageClass(ctx context.Context, name string, input *ModifyLVMThinStorageClassInput) error {
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

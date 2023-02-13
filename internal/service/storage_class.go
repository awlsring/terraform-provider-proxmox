package service

import (
	"context"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
)

type Storage struct {
	Id          string
	SharedNodes []string
	Shared      bool
	Local       bool
	Size        int64
	Source      string
	Content     []string
	Type        proxmox.StorageType
}

// deprecated
func (c *Proxmox) DescribeStorage(ctx context.Context) ([]Storage, error) {
	storage, err := c.ListStorage(ctx)
	if err != nil {
		return nil, err
	}

	storageList := []Storage{}
	for _, ss := range storage {
		nodes := StringCommaPtrListToSlice(ss.Nodes)
		if len(nodes) != 0 {
			storageSummary, err := c.GetStorage(ctx, nodes[0], ss.Storage)
			if err != nil {
				return nil, err
			}
			source, err := determineStorageSource(ss, storageSummary.Type)
			if err != nil {
				return nil, err
			}
			s := Storage{
				Id:          fmt.Sprintf(storageSummary.Storage),
				SharedNodes: nodes,
				Source:      source,
				Shared:      BooleanIntegerConversion(storageSummary.Shared),
				Local:       false,
				Content:     StringCommaListToSlice(storageSummary.Content),
				Type:        storageSummary.Type,
				Size:        PtrFloatToInt64(storageSummary.Total),
			}
			storageList = append(storageList, s)
		}
	}

	return storageList, nil
}

func determineStorageSource(s proxmox.StorageSummary, t proxmox.StorageType) (string, error) {
	switch t {
	case proxmox.STORAGETYPE_ZFSPOOL:

		return PtrStringToString(s.Pool), nil
	default:
		return "", fmt.Errorf("unknown storage type: %s", t)
	}
}

// deprecated
func (c *Proxmox) DescribeLocalStorage(ctx context.Context) ([]Storage, error) {
	nodes, err := c.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	nodeList := []string{}
	for _, node := range nodes {
		nodeList = append(nodeList, node.Node)
	}

	storageSummaries := []Storage{}
	for _, node := range nodes {
		local, err := c.GetStorage(ctx, node.Node, "local")
		if err != nil {
			return nil, err
		}

		ls := Storage{
			Id:      fmt.Sprintf("%s/%s", node.Node, local.Storage),
			Shared:  BooleanIntegerConversion(local.Shared),
			Source:  node.Node,
			Local:   true,
			Content: StringCommaListToSlice(local.Content),
			Type:    local.Type,
			Size:    PtrFloatToInt64(local.Total),
		}
		if ls.Shared {
			ls.SharedNodes = nodeList
		} else {
			ls.SharedNodes = []string{node.Node}
		}

		localLvm, err := c.GetStorage(ctx, node.Node, "local-lvm")
		if err != nil {
			return nil, err
		}

		lls := Storage{
			Id:      fmt.Sprintf("%s/%s", node.Node, localLvm.Storage),
			Shared:  BooleanIntegerConversion(localLvm.Shared),
			Source:  node.Node,
			Local:   true,
			Content: StringCommaListToSlice(localLvm.Content),
			Type:    localLvm.Type,
			Size:    PtrFloatToInt64(localLvm.Total),
		}
		if lls.Shared {
			lls.SharedNodes = nodeList
		} else {
			lls.SharedNodes = []string{node.Node}
		}

		storageSummaries = append(storageSummaries, ls, lls)
	}

	return storageSummaries, nil
}

func (c *Proxmox) GetStorageClass(ctx context.Context, storage string) (*proxmox.StorageSummary, error) {
	request := c.client.GetStorage(ctx, storage)
	resp, _, err := c.client.GetStorageExecute(request)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

// deprecated
func (c *Proxmox) GetStorage(ctx context.Context, node string, storage string) (*proxmox.NodeStorageSummary, error) {
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

func (c *Proxmox) ListStorage(ctx context.Context) ([]proxmox.StorageSummary, error) {
	request := c.client.ListStorage(ctx)
	resp, _, err := c.client.ListStorageExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Proxmox) listStorageOfType(ctx context.Context, filter proxmox.StorageType) ([]proxmox.StorageSummary, error) {
	request := c.client.ListStorage(ctx)
	request = request.Type_(filter)
	resp, _, err := c.client.ListStorageExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Proxmox) ListNodeStorage(ctx context.Context, node string) ([]proxmox.NodeStorageSummary, error) {
	request := c.client.ListNodeStorage(ctx, node)
	resp, _, err := c.client.ListNodeStorageExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

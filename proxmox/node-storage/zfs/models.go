package zfs

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nodeStorageZfsDataSourceModel struct {
	ZFS     []nodeStorageZfsModel `tfsdk:"node_storage_zfs"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

// maybe include content entries in the model
type nodeStorageZfsModel struct {
	ID           types.String   `tfsdk:"id"`
	Storage      types.String   `tfsdk:"storage"`
	Node         types.String   `tfsdk:"node"`
	ContentTypes []types.String `tfsdk:"content_types"`
	Size         types.Int64    `tfsdk:"size"`
	Pool         types.String   `tfsdk:"pool"`
	Mount        types.String   `tfsdk:"mount"`
}

func ZFSToModel(zfs *service.ZFSNodeStorage) nodeStorageZfsModel {
	m := nodeStorageZfsModel{
		ID:           types.StringValue(utils.FormId(zfs.Node, zfs.Storage)),
		Storage:      types.StringValue(zfs.Storage),
		Node:         types.StringValue(zfs.Node),
		ContentTypes: utils.UnpackList(zfs.ContentTypes),
		Size:         types.Int64Value(zfs.Size),
		Pool:         types.StringValue(zfs.ZFSPool),
		Mount:        types.StringValue(string(zfs.Mount)),
	}
	return m
}

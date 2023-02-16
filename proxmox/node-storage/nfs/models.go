package nfs

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nodeStorageNfsDataSourceModel struct {
	NFS     []nodeStorageNfsModel `tfsdk:"node_storage_nfs"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type nodeStorageNfsModel struct {
	ID           types.String   `tfsdk:"id"`
	Storage      types.String   `tfsdk:"storage"`
	Node         types.String   `tfsdk:"node"`
	ContentTypes []types.String `tfsdk:"content_types"`
	Size         types.Int64    `tfsdk:"size"`
	Mount        types.String   `tfsdk:"mount"`
	Export       types.String   `tfsdk:"export"`
	Server       types.String   `tfsdk:"server"`
}

func NFSToModel(nfs *service.NFSNodeStorage) nodeStorageNfsModel {
	m := nodeStorageNfsModel{
		ID:           types.StringValue(utils.FormId(nfs.Node, nfs.Storage)),
		Storage:      types.StringValue(nfs.Storage),
		Node:         types.StringValue(nfs.Node),
		ContentTypes: utils.UnpackList(nfs.ContentTypes),
		Size:         types.Int64Value(nfs.Size),
		Mount:        types.StringValue(string(nfs.Mount)),
		Export:       types.StringValue(string(nfs.Export)),
		Server:       types.StringValue(string(nfs.Server)),
	}
	return m
}

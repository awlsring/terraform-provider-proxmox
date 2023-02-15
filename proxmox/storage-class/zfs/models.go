package zfs

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type storageClassZfsDataSourceModel struct {
	ZFS     []storageClassZfsModel `tfsdk:"zfs_storage_classos"`
	Filters []filters.FilterModel  `tfsdk:"filters"`
}

type storageClassZfsModel struct {
	ID           types.String `tfsdk:"id"`
	Nodes        types.List   `tfsdk:"nodes"`
	ContentTypes types.List   `tfsdk:"content_types"`
	Pool         types.String `tfsdk:"pool"`
	Mount        types.String `tfsdk:"mount"`
}

func ZFSStorageClassToModel(s *service.ZFSStorageClass) storageClassZfsModel {
	m := storageClassZfsModel{
		ID:           types.StringValue(s.Id),
		Pool:         types.StringValue(s.ZFSPool),
		ContentTypes: utils.UnpackListType(s.Content),
		Nodes:        utils.UnpackListType(s.Nodes),
		Mount:        types.StringValue(s.Mount),
	}

	return m
}

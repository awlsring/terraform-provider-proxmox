package nfs

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nfsStorageClassDataSourceModel struct {
	NFS     []nfsStorageClassModel `tfsdk:"nfs_storage_classes"`
	Filters []filters.FilterModel  `tfsdk:"filters"`
}

type nfsStorageClassModel struct {
	ID           types.String `tfsdk:"id"`
	Server       types.String `tfsdk:"server"`
	Nodes        types.List   `tfsdk:"nodes"`
	ContentTypes types.List   `tfsdk:"content_types"`
	Mount        types.String `tfsdk:"mount"`
	Export       types.String `tfsdk:"export"`
}

func NFSStorageClassToModel(s *service.NFSStorageClass) nfsStorageClassModel {
	m := nfsStorageClassModel{
		ID:           types.StringValue(s.Id),
		Server:       types.StringValue(s.Server),
		Nodes:        utils.UnpackListType(s.Nodes),
		ContentTypes: utils.UnpackListType(s.Content),
		Mount:        types.StringValue(s.Mount),
		Export:       types.StringValue(s.Export),
	}

	return m
}

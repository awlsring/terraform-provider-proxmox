package lvm

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type lvmStorageClassDataSourceModel struct {
	LVM     []lvmStorageClassModel `tfsdk:"lvm_storage_classes"`
	Filters []filters.FilterModel  `tfsdk:"filters"`
}

type lvmStorageClassModel struct {
	ID           types.String `tfsdk:"id"`
	VolumeGroup  types.String `tfsdk:"volume_group"`
	Nodes        types.List   `tfsdk:"nodes"`
	ContentTypes types.List   `tfsdk:"content_types"`
}

func LVMStorageClassToModel(s *service.LVMStorageClass) lvmStorageClassModel {
	m := lvmStorageClassModel{
		ID:           types.StringValue(s.Id),
		VolumeGroup:  types.StringValue(s.VolumeGroup),
		Nodes:        utils.UnpackListType(s.Nodes),
		ContentTypes: utils.UnpackListType(s.Content),
	}

	return m
}

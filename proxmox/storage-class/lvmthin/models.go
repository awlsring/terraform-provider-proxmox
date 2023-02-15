package lvmthin

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type lvmThinStorageClassDataSourceModel struct {
	LVMThinpools []lvmThinStorageClassModel `tfsdk:"lvm_thinpool_storage_classes"`
	Filters      []filters.FilterModel      `tfsdk:"filters"`
}

type lvmThinStorageClassModel struct {
	ID           types.String `tfsdk:"id"`
	VolumeGroup  types.String `tfsdk:"volume_group"`
	Thinpool     types.String `tfsdk:"thinpool"`
	Nodes        types.List   `tfsdk:"nodes"`
	ContentTypes types.List   `tfsdk:"content_types"`
}

func LVMThinStorageClassToModel(s *service.LVMThinStorageClass) lvmThinStorageClassModel {
	m := lvmThinStorageClassModel{
		ID:           types.StringValue(s.Id),
		VolumeGroup:  types.StringValue(s.VolumeGroup),
		Thinpool:     types.StringValue(s.Thinpool),
		Nodes:        utils.UnpackListType(s.Nodes),
		ContentTypes: utils.UnpackListType(s.Content),
	}

	return m
}

package lvmthin

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type lvmThinpoolDataSourceModel struct {
	LVMThinpools []lvmThinpoolModel    `tfsdk:"lvm_thinpools"`
	Filters      []filters.FilterModel `tfsdk:"filters"`
}

type lvmThinpoolModel struct {
	ID           types.String `tfsdk:"id"`
	Node         types.String `tfsdk:"node"`
	Name         types.String `tfsdk:"name"`
	Size         types.Int64  `tfsdk:"size"`
	MetadataSize types.Int64  `tfsdk:"metadata_size"`
	VolumeGroup  types.String `tfsdk:"volume_group"`
	Device       types.String `tfsdk:"device"`
}

func LVMThinpoolToModel(l *service.LVMThinpool) lvmThinpoolModel {
	m := lvmThinpoolModel{
		ID:           types.StringValue(utils.FormId(l.Node, l.Name)),
		Node:         types.StringValue(l.Node),
		Name:         types.StringValue(l.Name),
		Size:         types.Int64Value(l.Size),
		MetadataSize: types.Int64Value(l.MetadataSize),
		VolumeGroup:  types.StringValue(utils.FormId(l.Node, l.VolumeGroup)),
		Device:       types.StringValue(l.Device),
	}

	return m
}

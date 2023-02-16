package lvmthin

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nodeStorageLVMThinDataSourceModel struct {
	LVMThins []nodeStorageLVMThinModel `tfsdk:"node_storage_lvm_thinpools"`
	Filters  []filters.FilterModel     `tfsdk:"filters"`
}

type nodeStorageLVMThinModel struct {
	ID           types.String   `tfsdk:"id"`
	Storage      types.String   `tfsdk:"storage"`
	Node         types.String   `tfsdk:"node"`
	ContentTypes []types.String `tfsdk:"content_types"`
	Size         types.Int64    `tfsdk:"size"`
	VolumeGroup  types.String   `tfsdk:"volume_group"`
	Thinpool     types.String   `tfsdk:"thinpool"`
}

func LVMThinToModel(l *service.LVMThinNodeStorage) nodeStorageLVMThinModel {
	m := nodeStorageLVMThinModel{
		ID:           types.StringValue(utils.FormId(l.Node, l.Storage)),
		Storage:      types.StringValue(l.Storage),
		Node:         types.StringValue(l.Node),
		ContentTypes: utils.UnpackList(l.ContentTypes),
		Size:         types.Int64Value(l.Size),
		VolumeGroup:  types.StringValue(l.VolumeGroup),
		Thinpool:     types.StringValue(l.Thinpool),
	}
	return m
}

package lvmthin

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nodeStorageLVMDataSourceModel struct {
	LVMs    []nodeStorageLVMModel `tfsdk:"node_storage_lvms"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type nodeStorageLVMModel struct {
	ID           types.String   `tfsdk:"id"`
	Storage      types.String   `tfsdk:"storage"`
	Node         types.String   `tfsdk:"node"`
	ContentTypes []types.String `tfsdk:"content_types"`
	Size         types.Int64    `tfsdk:"size"`
	VolumeGroup  types.String   `tfsdk:"volume_group"`
}

func LVMToModel(l *service.LVMNodeStorage) nodeStorageLVMModel {
	m := nodeStorageLVMModel{
		ID:           types.StringValue(utils.FormId(l.Node, l.Storage)),
		Storage:      types.StringValue(l.Storage),
		Node:         types.StringValue(l.Node),
		ContentTypes: utils.UnpackList(l.ContentTypes),
		Size:         types.Int64Value(l.Size),
		VolumeGroup:  types.StringValue(l.VolumeGroup),
	}
	return m
}

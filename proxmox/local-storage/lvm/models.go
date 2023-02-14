package lvm

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type lvmDataSourceModel struct {
	LVMs    []lvmModel            `tfsdk:"lvms"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type lvmModel struct {
	ID     types.String `tfsdk:"id"`
	Node   types.String `tfsdk:"node"`
	Name   types.String `tfsdk:"name"`
	Size   types.Int64  `tfsdk:"size"`
	Device types.String `tfsdk:"device"`
}

func LVMToModel(l *service.LVM) lvmModel {
	m := lvmModel{
		ID:     types.StringValue(utils.FormId(l.Node, l.Name)),
		Node:   types.StringValue(l.Node),
		Name:   types.StringValue(l.Name),
		Size:   types.Int64Value(l.Size),
		Device: types.StringValue(l.Device),
	}

	return m
}

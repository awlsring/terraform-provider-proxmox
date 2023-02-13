package zfs

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type storageClassZfsDataSourceModel struct {
	ZFS     []storageClassZfsModel `tfsdk:"storage_class_zfs"`
	Filters []filters.FilterModel  `tfsdk:"filters"`
}

type storageClassZfsModel struct {
	ID           types.String   `tfsdk:"id"`
	Nodes        []types.String `tfsdk:"nodes"`
	ContentTypes []types.String `tfsdk:"content_types"`
	Pool         types.String   `tfsdk:"pool"`
	Mount        types.String   `tfsdk:"mount"`
}

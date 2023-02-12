package zfs

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type zfsDataSourceModel struct {
	ZFSPools []zfsModel            `tfsdk:"zfs_pools"`
	Filters  []filters.FilterModel `tfsdk:"filters"`
}

type zfsModel struct {
	ID     types.String   `tfsdk:"id"`
	Node   types.String   `tfsdk:"node"`
	Name   types.String   `tfsdk:"name"`
	Size   types.Int64    `tfsdk:"size"`
	Health types.String   `tfsdk:"health"`
	Disks  []types.String `tfsdk:"disks"`
}

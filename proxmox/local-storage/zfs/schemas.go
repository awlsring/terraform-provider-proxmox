package zfs

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	ds "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var filter = filters.FilterConfig{"node"}

var dataSourceSchema = ds.Schema{
	Attributes: map[string]ds.Attribute{
		"filters": filter.Schema(),
		"zfs_pools": ds.ListNestedAttribute{
			Computed: true,
			NestedObject: ds.NestedAttributeObject{
				Attributes: map[string]ds.Attribute{
					"id": ds.StringAttribute{
						Computed:    true,
						Description: "The id of the ZFS pool. Formatted as `{node}/{name}`.",
					},
					"node": ds.StringAttribute{
						Computed:    true,
						Description: "The node the ZFS pool is on.",
					},
					"name": ds.StringAttribute{
						Computed:    true,
						Description: "The name of the ZFS pool.",
					},
					"size": ds.Int64Attribute{
						Computed:    true,
						Description: "The size of the ZFS pool in bytes.",
					},
					"health": ds.StringAttribute{
						Computed:    true,
						Description: "The health of the ZFS pool.",
					},
				},
			},
		},
	},
}

package zfs

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	ds "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// add storage class filter
var filter = filters.FilterConfig{"node"}
var dataSourceSchema = ds.Schema{
	Attributes: map[string]ds.Attribute{
		"filters": filter.Schema(),
		"node_storage_zfs": ds.ListNestedAttribute{
			Computed: true,
			NestedObject: ds.NestedAttributeObject{
				Attributes: map[string]ds.Attribute{
					"id": ds.StringAttribute{
						Computed:    true,
						Description: "The identifier of the node storage. Formatted as `{node}/{storage}`",
					},
					"storage": ds.StringAttribute{
						Computed:    true,
						Description: "The name of the storage class this implements.",
					},
					"node": ds.StringAttribute{
						Computed:    true,
						Description: "The node which the class implementation is on.",
					},
					"content_types": ds.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "The content types that can be stored on this storage class.",
					},
					"size": ds.Int64Attribute{
						Computed:    true,
						Description: "The size of the storage in bytes.",
					},
					"pool": ds.StringAttribute{
						Computed:    true,
						Description: "The ZFS pool of the storage.",
					},
					"mount": ds.StringAttribute{
						Computed:    true,
						Description: "The path the ZFS pool should be mounted at on each node.",
					},
				},
			},
		},
	},
}

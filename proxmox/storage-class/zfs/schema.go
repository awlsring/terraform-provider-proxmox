package zfs

import (
	ds "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var dataSourceSchema = ds.Schema{
	Attributes: map[string]ds.Attribute{
		"filters": filter.Schema(),
		"storage_class_zfs": ds.ListNestedAttribute{
			Computed: true,
			NestedObject: ds.NestedAttributeObject{
				Attributes: map[string]ds.Attribute{
					"id": ds.StringAttribute{
						Computed:    true,
						Description: "The identifier of the storage class.",
					},
					"nodes": ds.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "Nodes that implement this storage class.",
					},
					"content": ds.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "The content types that can be stored on this storage class.",
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

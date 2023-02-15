package zfs

import (
	ds "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rs "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var dataSourceSchema = ds.Schema{
	Attributes: map[string]ds.Attribute{
		"filters": filter.Schema(),
		"zfs_storage_classes": ds.ListNestedAttribute{
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
					"content_types": ds.ListAttribute{
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

var resourceSchema = rs.Schema{
	Attributes: map[string]rs.Attribute{
		"id": rs.StringAttribute{
			Required:    true,
			Description: "The id of the ZFS storage class.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"pool": rs.StringAttribute{
			Required:    true,
			Description: "The ZFS pool of the storage.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"mount": rs.StringAttribute{
			Computed:    true,
			Description: "The path the ZFS pool should be mounted at on each node.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"nodes": rs.ListAttribute{
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			Description: "Nodes that implement this storage class.",
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"content_types": rs.ListAttribute{
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			Description: "The content types that can be stored on this storage class.",
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
	},
}

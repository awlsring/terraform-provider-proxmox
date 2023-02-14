package lvmthin

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	ds "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rs "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var filter = filters.FilterConfig{"node"}

var dataSourceSchema = ds.Schema{
	Attributes: map[string]ds.Attribute{
		"filters": filter.Schema(),
		"lvm_thinpools": ds.ListNestedAttribute{
			Computed: true,
			NestedObject: ds.NestedAttributeObject{
				Attributes: map[string]ds.Attribute{
					"id": ds.StringAttribute{
						Computed:    true,
						Description: "The id of the LVM thinpool. Formatted as `{node}/{name}`.",
					},
					"node": ds.StringAttribute{
						Computed:    true,
						Description: "The node the LVM thinpool is on.",
					},
					"name": ds.StringAttribute{
						Computed:    true,
						Description: "The name of the LVM thinpool.",
					},
					"size": ds.Int64Attribute{
						Computed:    true,
						Description: "The size of the LVM thinpool in bytes.",
					},
					"metadata_size": ds.Int64Attribute{
						Computed:    true,
						Description: "The size of the LVM thinpool metadata lv in bytes.",
					},
					"volume_group": ds.StringAttribute{
						Computed:    true,
						Description: "The associated volume group. Formatted as `{node}/{volume_group}`",
					},
					"device": ds.StringAttribute{
						Computed:    true,
						Description: "The device used to create the LVM thinpool.`",
					},
				},
			},
		},
	},
}

var resourceSchema = rs.Schema{
	Attributes: map[string]rs.Attribute{
		"id": rs.StringAttribute{
			Computed:    true,
			Description: "The id of the LVM thinpool. Formatted as `{node}/{name}`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node": rs.StringAttribute{
			Required:    true,
			Description: "The node the LVM thinpool is on.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": rs.StringAttribute{
			Required:    true,
			Description: "The name of the LVM thinpool.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"size": rs.Int64Attribute{
			Computed:    true,
			Description: "The size of the LVM thinpool in bytes.",
		},
		"metadata_size": rs.Int64Attribute{
			Computed:    true,
			Description: "The size of the LVM thinpool metadata lv in bytes.",
		},
		"volume_group": rs.StringAttribute{
			Computed:    true,
			Description: "The associated volume group id. Formatted as `{node}/{volume_group}`",
		},
		"device": rs.StringAttribute{
			Required:    true,
			Description: "The device to create the LVM thinpool on.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	},
}

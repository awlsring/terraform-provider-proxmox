package lvm

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
		"lvms": ds.ListNestedAttribute{
			Computed: true,
			NestedObject: ds.NestedAttributeObject{
				Attributes: map[string]ds.Attribute{
					"id": ds.StringAttribute{
						Computed:    true,
						Description: "The id of the LVM. Formatted as `{node}/{name}`.",
					},
					"node": ds.StringAttribute{
						Computed:    true,
						Description: "The node the LVM is on.",
					},
					"name": ds.StringAttribute{
						Computed:    true,
						Description: "The name of the LVM.",
					},
					"size": ds.Int64Attribute{
						Computed:    true,
						Description: "The size of the LVM in bytes.",
					},
					"device": ds.StringAttribute{
						Computed:    true,
						Description: "Device the LVM is on.",
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
			Description: "The id of the LVM. Formatted as `{node}/{name}`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node": rs.StringAttribute{
			Required:    true,
			Description: "The node the LVM is on.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": rs.StringAttribute{
			Required:    true,
			Description: "The name of the LVM.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"size": rs.Int64Attribute{
			Computed:    true,
			Description: "The size of the LVM in bytes.",
		},
		"device": rs.StringAttribute{
			Required:    true,
			Description: "Device the LVM is on.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	},
}

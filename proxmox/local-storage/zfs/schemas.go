package zfs

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	ds "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rs "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
					"raid_level": ds.StringAttribute{
						Computed:    true,
						Description: "The RAID level of the ZFS pool.",
					},
					"health": ds.StringAttribute{
						Computed:    true,
						Description: "The health of the ZFS pool.",
					},
					"disks": ds.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "List of disks that make up the pool.",
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
			Description: "The id of the ZFS pool. Formatted as `{node}/{name}`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node": rs.StringAttribute{
			Required:    true,
			Description: "The node the ZFS pool is on.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": rs.StringAttribute{
			Required:    true,
			Description: "The name of the ZFS pool.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"raid_level": rs.StringAttribute{
			Required:    true,
			Description: "The RAID level of the ZFS pool.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			// Commented out options that im not sure how they work
			// TODO: See if subvalidators work to ensure disk amount match right raid level
			Validators: []validator.String{
				stringvalidator.OneOf(
					"single",
					"mirror",
					// "raid10",
					"raidz",
					"raidz2",
					"raidz3",
					// "draid",
					// "draid2",
					// "draid3",
				),
			},
		},
		"size": rs.Int64Attribute{
			Computed:    true,
			Description: "Size of the ZFS pool.",
		},
		"health": rs.StringAttribute{
			Computed:    true,
			Description: "Health of the ZFS pool.",
		},
		"disks": rs.ListAttribute{
			Required:    true,
			ElementType: types.StringType,
			Description: "List of disks that make the ZFS pool.",
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
	},
}

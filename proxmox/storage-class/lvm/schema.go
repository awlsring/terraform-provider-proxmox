package lvm

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/storage-class"
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
		"lvm_storage_classes": ds.ListNestedAttribute{
			Computed: true,
			NestedObject: ds.NestedAttributeObject{
				Attributes: map[string]ds.Attribute{
					"id": ds.StringAttribute{
						Computed:    true,
						Description: "The identifier of the storage class.",
					},
					"volume_group": ds.StringAttribute{
						Computed:    true,
						Description: "The volume group used by the thinpool.",
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
				},
			},
		},
	},
}

var resourceSchema = rs.Schema{
	Attributes: map[string]rs.Attribute{
		"id": rs.StringAttribute{
			Required:    true,
			Description: "The identifier of the storage class.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"volume_group": rs.StringAttribute{
			Required:    true,
			Description: "The associated volume group.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
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
			Validators: storage.ContentTypeValidator(
				"images",
				"rootdir",
			),
		},
	},
}

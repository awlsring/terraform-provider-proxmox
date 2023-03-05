package schemas

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/defaults"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var DiskObjectDataSourceSchema = dschema.NestedAttributeObject{
	Attributes: map[string]dschema.Attribute{
		"storage": dschema.StringAttribute{
			Computed:    true,
			Description: "The storage the disk is on.",
		},
		"file_format": dschema.StringAttribute{
			Computed:    true,
			Description: "The file format of the disk.",
		},
		"size": dschema.Int64Attribute{
			Computed:    true,
			Description: "The size of the disk in GiB.",
		},
		"use_iothread": dschema.BoolAttribute{
			Computed:    true,
			Description: "Whether to use an iothread for the disk.",
		},
		"speed_limits": dschema.SingleNestedAttribute{
			Computed:    true,
			Description: "The speed limits of the disk. If not set, no speed limitations are applied.",
			Attributes: map[string]dschema.Attribute{
				"read": dschema.Int64Attribute{
					Computed:    true,
					Description: "The read speed limit in bytes per second.",
				},
				"write": dschema.Int64Attribute{
					Computed:    true,
					Description: "The write speed limit in bytes per second.",
				},
				"write_burstable": dschema.Int64Attribute{
					Computed:    true,
					Description: "The write burstable speed limit in bytes per second.",
				},
				"read_burstable": dschema.Int64Attribute{
					Computed:    true,
					Description: "The read burstable speed limit in bytes per second.",
				},
			},
		},
		"interface_type": dschema.StringAttribute{
			Computed:    true,
			Description: "The type of the disk.",
		},
		"ssd_emulation": dschema.BoolAttribute{
			Computed:    true,
			Description: "Whether to use SSD emulation. conflicts with virtio disk type.",
		},
		"position": dschema.Int64Attribute{
			Computed:    true,
			Description: "The position of the disk. (0, 1, 2, etc.) This is combined with the `interface_type` to determine the disk name.",
		},
		"discard": dschema.BoolAttribute{
			Computed:    true,
			Description: "Whether the disk has discard enabled.",
		},
	},
}

var DiskObjectSchema = schema.NestedAttributeObject{
	PlanModifiers: []planmodifier.Object{
		objectplanmodifier.UseStateForUnknown(),
	},
	Attributes: map[string]schema.Attribute{
		"storage": schema.StringAttribute{
			Required:    true,
			Description: "The storage the disk is on.",
		},
		"file_format": schema.StringAttribute{
			Optional:    true,
			Description: "The file format of the disk.",
			Validators: []validator.String{
				stringvalidator.OneOf(
					"raw",
					"qcow2",
					"vmdk",
				),
			},
		},
		"size": schema.Int64Attribute{
			Required:    true,
			Description: "The size of the disk in GiB.",
		},
		"use_iothread": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether to use an iothread for the disk.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(false),
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"speed_limits": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "The speed limits of the disk. If not set, no speed limitations are applied.",
			Attributes: map[string]schema.Attribute{
				"read": schema.Int64Attribute{
					Optional:    true,
					Description: "The read speed limit in bytes per second.",
				},
				"write": schema.Int64Attribute{
					Optional:    true,
					Description: "The write speed limit in bytes per second.",
				},
				"write_burstable": schema.Int64Attribute{
					Optional:    true,
					Description: "The write burstable speed limit in bytes per second.",
				},
				"read_burstable": schema.Int64Attribute{
					Optional:    true,
					Description: "The read burstable speed limit in bytes per second.",
				},
			},
		},
		"interface_type": schema.StringAttribute{
			Required:    true,
			Description: "The type of the disk.",
			Validators: []validator.String{
				stringvalidator.OneOf(
					"scsi",
					"sata",
					"virtio",
				),
			},
		},
		// add conflict with virtio
		"ssd_emulation": schema.BoolAttribute{
			Computed:    true,
			Optional:    true,
			Description: "Whether to use SSD emulation. conflicts with virtio disk type.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(false),
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"position": schema.Int64Attribute{
			Required:    true,
			Description: "The position of the disk. (0, 1, 2, etc.) This is combined with the `interface_type` to determine the disk name.",
		},
		"discard": schema.BoolAttribute{
			Computed:    true,
			Optional:    true,
			Description: "Whether the disk has discard enabled.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(true),
				boolplanmodifier.UseStateForUnknown(),
			},
		},
	},
}

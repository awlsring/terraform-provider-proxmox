package schemas

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/defaults"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var PCIDeviceObjectSchema = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The device name of the PCI device.",
		},
		"id": schema.StringAttribute{
			Required:    true,
			Description: "The device ID of the PCI device.",
		},
		"pcie": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the PCI device is PCIe.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(false),
			},
		},
		"primary_gpu": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the PCI device is the primary GPU.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(false),
			},
		},
		"mdev": schema.StringAttribute{
			Optional:    true,
			Description: "The mediated device name.",
		},
		"rombar": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Make the firmware room visible to the VM.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(true),
			},
		},
		"rom_file": schema.StringAttribute{
			Optional:    true,
			Description: "The relative path to the ROM for the device.",
		},
	},
}

var PCIDeviceObjectDataSourceSchema = dschema.NestedAttributeObject{
	Attributes: map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: "The device name of the PCI device.",
		},
		"id": dschema.StringAttribute{
			Computed:    true,
			Description: "The device ID of the PCI device.",
		},
		"pcie": dschema.BoolAttribute{
			Computed:    true,
			Description: "Whether the PCI device is PCIe.",
		},
		"primary_gpu": dschema.BoolAttribute{
			Computed:    true,
			Description: "Whether the PCI device is the primary GPU.",
		},
		"mdev": dschema.StringAttribute{
			Computed:    true,
			Description: "The mediated device name.",
		},
		"rombar": dschema.BoolAttribute{
			Computed:    true,
			Description: "Make the firmware room visible to the VM.",
		},
		"rom_file": dschema.StringAttribute{
			Computed:    true,
			Description: "The relative path to the ROM for the device.",
		},
	},
}

package qemu

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/defaults"
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

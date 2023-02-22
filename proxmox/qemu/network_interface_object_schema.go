package qemu

import (
	"regexp"

	"github.com/awlsring/terraform-provider-proxmox/proxmox/defaults"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var NetworkInterfaceObjectSchema = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"bridge": schema.StringAttribute{
			Required:    true,
			Description: "The bridge the network interface is on.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("vmbr[0-9]$"), "name must follow scheme `vmbr<n>`"),
			},
		},
		"enabled": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the network interface is enabled.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(true),
			},
		},
		"use_firewall": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the firewall for the network interface is enabled.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(false),
			},
		},
		"mac_address": schema.StringAttribute{
			Computed:    true,
			Optional:    true,
			Description: "The MAC address of the network interface.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"), "must be a valid MAC address"),
			},
		},
		"model": schema.StringAttribute{
			Computed:    true,
			Optional:    true,
			Description: "The model of the network interface.",
			PlanModifiers: []planmodifier.String{
				defaults.DefaultString("virtio"),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					"virtio",
					"e1000",
					"rtl8139",
					"vmxnet3",
				),
			},
		},
		"rate_limit": schema.Int64Attribute{
			Optional:    true,
			Description: "The rate limit of the network interface in megabytes per second.",
		},
		"position": schema.Int64Attribute{
			Required:    true,
			Description: "The position of the network interface in the VM. (0, 1, 2, etc.)",
		},
		"vlan": schema.Int64Attribute{
			Optional:    true,
			Description: "The VLAN tag of the network interface.",
		},
		"mtu": schema.Int64Attribute{
			Optional:    true,
			Description: "The MTU of the network interface. Only valid for virtio.",
		},
	},
}

package schemas

import (
	"regexp"

	"github.com/awlsring/terraform-provider-proxmox/proxmox/defaults"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var NetworkInterfaceObjectSchema = schema.NestedAttributeObject{
	PlanModifiers: []planmodifier.Object{
		objectplanmodifier.UseStateForUnknown(),
	},
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
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"use_firewall": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the firewall for the network interface is enabled.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(false),
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"mac_address": schema.StringAttribute{
			Computed:    true,
			Optional:    true,
			Description: "The MAC address of the network interface.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"), "must be a valid MAC address"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"model": schema.StringAttribute{
			Computed:    true,
			Optional:    true,
			Description: "The model of the network interface.",
			PlanModifiers: []planmodifier.String{
				defaults.DefaultString("virtio"),
				stringplanmodifier.UseStateForUnknown(),
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
			Description: "The position of the network interface in the VM as an int. Used to determine the interface name (net0, net1, etc).",
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

var NetworkInterfaceObjectDataSourceSchema = dschema.NestedAttributeObject{
	Attributes: map[string]dschema.Attribute{
		"bridge": dschema.StringAttribute{
			Computed:    true,
			Description: "The bridge the network interface is on.",
		},
		"enabled": dschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the network interface is enabled.",
		},
		"use_firewall": dschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the firewall for the network interface is enabled.",
		},
		"mac_address": dschema.StringAttribute{
			Computed:    true,
			Optional:    true,
			Description: "The MAC address of the network interface.",
		},
		"model": dschema.StringAttribute{
			Computed:    true,
			Optional:    true,
			Description: "The model of the network interface.",
		},
		"rate_limit": dschema.Int64Attribute{
			Optional:    true,
			Description: "The rate limit of the network interface in megabytes per second.",
		},
		"position": dschema.Int64Attribute{
			Computed:    true,
			Description: "The position of the network interface in the VM as an int. Used to determine the interface name (net0, net1, etc).",
		},
		"vlan": dschema.Int64Attribute{
			Optional:    true,
			Description: "The VLAN tag of the network interface.",
		},
		"mtu": dschema.Int64Attribute{
			Optional:    true,
			Description: "The MTU of the network interface. Only valid for virtio.",
		},
	},
}

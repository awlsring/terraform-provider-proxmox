package bonds

import (
	"regexp"

	"github.com/awlsring/terraform-provider-proxmox/proxmox/defaults"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	ds "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rs "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var filter = filters.FilterConfig{"node"}

var dataSourceSchema = ds.Schema{
	Attributes: map[string]ds.Attribute{
		"filters": filter.Schema(),
		"network_bonds": ds.ListNestedAttribute{
			Computed: true,
			NestedObject: ds.NestedAttributeObject{
				Attributes: map[string]ds.Attribute{
					"id": ds.StringAttribute{
						Computed:    true,
						Description: "The id of the bond. Formatted as `{node}/{name}`.",
					},
					"node": ds.StringAttribute{
						Computed:    true,
						Description: "The node the bond is on.",
					},
					"name": ds.StringAttribute{
						Computed:    true,
						Description: "The name of the bond.",
					},
					"active": ds.BoolAttribute{
						Computed:    true,
						Description: "If the bond is active.",
					},
					"autostart": ds.BoolAttribute{
						Computed:    true,
						Description: "If the bond is set to autostart.",
					},
					"hash_policy": ds.StringAttribute{
						Computed:    true,
						Description: "Hash policy used on the bond.",
					},
					"bond_primary": ds.StringAttribute{
						Computed:    true,
						Description: "Primary interface on the bond.",
					},
					"comments": ds.StringAttribute{
						Computed:    true,
						Description: "Comments on the bond.",
					},
					"mode": ds.StringAttribute{
						Computed:    true,
						Description: "Mode of the bond.",
					},
					"mii_mon": ds.StringAttribute{
						Computed:    true,
						Description: "Miimon of the bond.",
					},
					"interfaces": ds.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "List of interfaces on the bond.",
					},
					"ipv4":         network.IpDataSourceSchema(network.IP_ADDRESS_TYPE_4),
					"ipv6":         network.IpDataSourceSchema(network.IP_ADDRESS_TYPE_6),
					"ipv4_gateway": network.IpGatewayDataSourceSchema(network.IP_ADDRESS_TYPE_4),
					"ipv6_gateway": network.IpGatewayDataSourceSchema(network.IP_ADDRESS_TYPE_6),
				},
			},
		},
	},
}

var resourceSchema = rs.Schema{
	Attributes: map[string]rs.Attribute{
		"id": rs.StringAttribute{
			Computed:    true,
			Description: "The id of the bond. Formatted as `{node}/{name}`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node": rs.StringAttribute{
			Required:    true,
			Description: "The node the bond is on.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": rs.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The name of the bond.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("bond[0-9]$"), "name must follow scheme `bond<n>`"),
			},
		},
		"active": rs.BoolAttribute{
			Computed:    true,
			Description: "If the bond is active.",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"autostart": rs.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "If the bond is set to autostart.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(true),
			},
		},
		"hash_policy": rs.StringAttribute{
			Optional:    true,
			Description: "Hash policy used on the bond.",
			Validators: []validator.String{
				stringvalidator.OneOf("layer2", "layer2+3", "layer3+4"),
			},
		},
		"bond_primary": rs.StringAttribute{
			Optional:    true,
			Description: "Primary interface on the bond.",
		},
		"mode": rs.StringAttribute{
			Required:    true,
			Description: "Mode of the bond.",
			Validators: []validator.String{
				stringvalidator.OneOf(
					"balance-rr",
					"active-backup",
					"balance-xor",
					"broadcast",
					"802.3ad",
					"balance-tlb",
					"balance-alb",
					"balance-slb",
					"lacp-balance-slb",
					"lacp-balance-tcp",
				),
			},
		},
		"comments": rs.StringAttribute{
			Optional:    true,
			Description: "Comment in the bond.",
		},
		"mii_mon": rs.StringAttribute{
			Computed:    true,
			Description: "Miimon of the bond.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"interfaces": rs.ListAttribute{
			Required:    true,
			ElementType: types.StringType,
			Description: "List of interfaces on the bond.",
		},
		"ipv4":         network.IpResourceSchema(network.IP_ADDRESS_TYPE_4),
		"ipv6":         network.IpResourceSchema(network.IP_ADDRESS_TYPE_6),
		"ipv4_gateway": network.IpGatewayResourceSchema(network.IP_ADDRESS_TYPE_4),
		"ipv6_gateway": network.IpGatewayResourceSchema(network.IP_ADDRESS_TYPE_6),
	},
}

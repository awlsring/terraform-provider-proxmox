package bridges

import (
	"regexp"

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
		"network_bridges": ds.ListNestedAttribute{
			Computed: true,
			NestedObject: ds.NestedAttributeObject{
				Attributes: map[string]ds.Attribute{
					"id": ds.StringAttribute{
						Computed:    true,
						Description: "The id of the bridge. Formatted as `{node}/{name}`.",
					},
					"node": ds.StringAttribute{
						Computed:    true,
						Description: "The node the bridge is on.",
					},
					"name": ds.StringAttribute{
						Computed:    true,
						Description: "The name of the bridge.",
					},
					"active": ds.BoolAttribute{
						Computed:    true,
						Description: "If the bridge is active.",
					},
					"autostart": ds.BoolAttribute{
						Computed:    true,
						Description: "If the bridge is set to autostart.",
					},
					"vlan_aware": ds.BoolAttribute{
						Computed:    true,
						Description: "If the bridge is vlan aware.",
					},
					"interfaces": ds.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "List of interfaces on the bridge.",
					},
					"comments": rs.StringAttribute{
						Optional:    true,
						Description: "Comment in the bond.",
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
			Description: "The id of the bridge. Formatted as `{node}/{name}`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node": rs.StringAttribute{
			Required:    true,
			Description: "The node the bridge is on.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": rs.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The name of the bridge. Follows the scheme `vmbr<n>`. If not set, the next available name will be used.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("vmbr[0-9]$"), "name must follow scheme `vmbr<n>`"),
			},
		},
		"active": rs.BoolAttribute{
			Computed:    true,
			Description: "If the bridge is active.",
		},
		"autostart": rs.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "If the bridge is set to autostart.",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"vlan_aware": rs.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "If the bridge is vlan aware.",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"comments": rs.StringAttribute{
			Optional:    true,
			Description: "Comment on the bridge.",
		},
		"interfaces": rs.ListAttribute{
			Required:    true,
			ElementType: types.StringType,
			Description: "List of interfaces on the bridge.",
		},
		"ipv4":         network.IpResourceSchema(network.IP_ADDRESS_TYPE_4),
		"ipv6":         network.IpResourceSchema(network.IP_ADDRESS_TYPE_6),
		"ipv4_gateway": network.IpGatewayResourceSchema(network.IP_ADDRESS_TYPE_4),
		"ipv6_gateway": network.IpGatewayResourceSchema(network.IP_ADDRESS_TYPE_6),
	},
}

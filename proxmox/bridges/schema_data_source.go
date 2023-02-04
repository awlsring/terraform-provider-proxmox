package bridges

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var bridgeDataSource = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The id of the bridge. Formatted as /{node}/{name}.",
	},
	"node": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The node the bridge is on.",
	},
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The name of the bridge.",
	},
	"active": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "If the bridge is active.",
	},
	"autostart": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "If the bridge is set to autostart.",
	},
	"vlan_aware": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "If the bridge is vlan aware.",
	},
	"interfaces": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of interfaces on the bridge.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"ipv4_address": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ipv4 address.",
	},
	"ipv4_gateway": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ipv4 gateway.",
	},
	"ipv4_netmask": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ipv4 netmask.",
	},
	"ipv6_address": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ipv6 address.",
	},
	"ipv6_gateway": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ipv6 gateway.",
	},
	"ipv6_netmask": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ipv6 netmask.",
	},
}
package bonds

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var bondDataSource = map[string]*schema.Schema{
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
	"hash_policy": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Hash policy used on the bond.",
	},
	"mode": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Mode of the bond.",
	},
	"mii_mon": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Miimon of the bond.",
	},
	"interfaces": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of interfaces on the bridge.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
}
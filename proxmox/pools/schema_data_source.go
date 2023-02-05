package pools

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var poolDataSource = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The id of the pool.",
	},
	"comment": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Notes of the pool.",
	},
	"members": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Resources that are part of the pool.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The id of the resource.",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The type of the resource.",
				},
			},
		},
	},
}
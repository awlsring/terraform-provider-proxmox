package storage

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var storageDataSource = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The storage identifier.",
	},
	"shared_nodes": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The nodes this storage is shared with.",
		Elem: &schema.Schema{Type: schema.TypeString},
	},
	"shared": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Nodes that host the storage pool.",
		Elem: &schema.Schema{Type: schema.TypeString},
	},
	"local": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "If this is local storage.",
	},
	"size": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Total size available in bytes.",
	},
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of storage.",
	},
	"content": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The content type supported by this storage.",
		Elem: &schema.Schema{Type: schema.TypeString},
	},
	"source": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The source of the space used by storage.",
	},
}
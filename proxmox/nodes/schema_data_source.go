package nodes

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var nodeDataSource = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Node id",
	},
	"node": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Node name",
	},
	"cores": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Amount of CPU cores on the machine",
	},
	"ssl_fingerprint": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The SSL fingerprint of the node",
	},
	"memory": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Amount of memory on the machine",
	},
	"total_disk_space": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Amount of disk space on the machine",
	},
	"disks": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of physical disks on the machine",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"device": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The device path.",
				},
				"size": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Disk size in bytes.",
				},
				"model": {
					Type:        schema.TypeString,
					Computed:   true,
					Description: "Disk model.",
				},
				"serial": {
					Type:        schema.TypeString,
					Computed:   true,
					Description: "Disk serial number.",
				},
				"vendor": {
					Type:        schema.TypeString,
					Computed:   true,
					Description: "Disk vendor.",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Disk type",
				},
			},
		},
	},
	"network_interfaces": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of physical network interfaces on the machine.",
		Elem: &schema.Schema{Type: schema.TypeString},
	},
}
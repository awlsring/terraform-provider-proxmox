package vms

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var virtualMachineDataSource = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "The virtual machine id.",
	},
	"node": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The virtual machine name.",
	},
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The virtual machine name.",
	},
	"cores": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "configured cpu cores on the virtual machine",
	},
	"memory": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "configured memory on the virtual machine",
	},
	"agent": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "If the QEMU guest agent is enabled",
	},
	"tags": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of tags on the virtual machine",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"disks": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of virtual disks on the machine",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"storage": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The storage device the volume is mounted on.",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The type of disk (scsi, virtio, ide, sata).",
				},
				"position": {
					Type:        schema.TypeString,
					Computed:   true,
					Description: "Connection position on the virtual machine (ex. virito0).",
				},
				"size": {
					Type:        schema.TypeInt,
					Computed:   true,
					Description: "Space allocated to the disk in bytes.",
				},
				"discard": {
					Type:        schema.TypeBool,
					Computed:   true,
					Description: "If discard in enabled.",
				},
			},
		},
	},
	"network_interfaces": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of virtual network interfaces on the machine.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"bridge": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Bridge the virtual device will run on.",
				},
				"vlan": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The vlan the nic will operate in.",
				},
				"model": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Interface model the nic emulates.",
				},
				"mac": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Mac address of the interface.",
				},
				"position": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Position of the interface on the virtual machine.",
				},
				"firewall": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "If firewall is enabled for virtual machine.",
				},
			},
			Description: "A virtual network interface on the machine.",
		},
	},
}
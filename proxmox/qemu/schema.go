package qemu

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var Schema = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.NumberAttribute{
				Computed:    true,
				Description: "The id of the resource.",
			},
			"node": schema.StringAttribute{
				Computed:    true,
				Description: "The owning node.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the template.",
			},
			"cores": schema.NumberAttribute{
				Computed:    true,
				Description: "The number of cores.",
			},
			"memory": schema.Int64Attribute{
				Computed:    true,
				Description: "The allocated of memory in bytes.",
			},
			"agent": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the the guest agent is installed.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				Description: "Tags on the resource.",
				ElementType: types.StringType,
			},
			"disks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"storage": schema.StringAttribute{
							Computed:    true,
							Description: "The storage the disk is on.",
						},
						"size": schema.Int64Attribute{
							Computed:    true,
							Description: "The size of the disk.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the disk.",
						},
						"position": schema.StringAttribute{
							Computed:    true,
							Description: "The position of the disk.",
						},
						"discard": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the disk has discard enabled.",
						},
					},
				},
			},
			"network_interfaces": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"bridge": schema.StringAttribute{
							Computed:    true,
							Description: "The bridge the network interface is on.",
						},
						"firewall": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the network interface has the firewall enabled.",
						},
						"model": schema.StringAttribute{
							Computed:    true,
							Description: "The model of the network interface.",
						},
						"mac_address": schema.StringAttribute{
							Computed:    true,
							Description: "The MAC address of the network interface.",
						},
						"vlan": schema.NumberAttribute{
							Computed:    true,
							Description: "The VLAN of the network interface.",
						},
						"position": schema.StringAttribute{
							Computed:    true,
							Description: "The position of the network interface.",
						},
					},
				},
			},
		},
	},
}

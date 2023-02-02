package proxmox

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNode() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNodeRead,
		Schema: map[string]*schema.Schema{
			"node": {
				Type:        schema.TypeString,
				Required:    true,
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The interface name.",
						},
					},
					Description: "A physical network interface on the machine.",
				},
			},
		},
	}
}

func dataSourceNodeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	nodeId := d.Get("node").(string)
	
	client := m.(*service.Proxmox)
	
	node, err := client.DescribeNode(ctx, nodeId)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeToResourceData(node, d)

	return diags
}

func nodeToResourceData(node *service.Node, d *schema.ResourceData) {
	d.SetId(node.Node)
	d.Set("cores", node.Cores)
	d.Set("sslFingerprint", node.SslFingerprint)
	d.Set("memory", node.Memory)
	d.Set("total_disk_space", node.DiskSpace)
	d.Set("disks", diskToResourceData(node.Disks))
	d.Set("network_interfaces", networkInterfaceToResourceData(node.NetworkInterfaces))
}

func diskToResourceData(disk []service.Disk) []interface{} {
	var disks []interface{}
	for _, disk := range disk {
		disks = append(disks, map[string]interface{}{
			"device": disk.Device,
			"size":   disk.Size,
			"model":  disk.Model,
			"serial": disk.Serial,
			"vendor": disk.Vendor,
			"type":   disk.Type,
		})
	}
	return disks
}

func networkInterfaceToResourceData(networkInterface []service.NetworkInterface) []interface{} {
	var networkInterfaces []interface{}
	for _, networkInterface := range networkInterface {
		networkInterfaces = append(networkInterfaces, map[string]interface{}{
			"name": networkInterface.Name,
		})
	}
	return networkInterfaces
}
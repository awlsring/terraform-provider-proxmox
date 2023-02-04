package nodes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/crypto/sha3"

	log "github.com/sirupsen/logrus"
)

func dataSourceNode() *schema.Resource {
	return &schema.Resource{
		Schema: nodeDataSource,
	}
}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNodeRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of node to find.",
						},
						"ip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "IP of a node to find.",
						},
					},
				},
			},
			"nodes": {
				Type:        schema.TypeList,
				Description: "The returned list of nodes.",
				Computed:    true,
				Elem:        dataSourceNode(),
			},
		},
	}
}

func dataSourceNodeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*service.Proxmox)

	filterId, err := makeListId(d)
	if err != nil {
		return diag.Errorf("failed to generate filter id: %s", err)
	}

	names := []string{}

	filters := d.Get("filter")
	for _, filter := range filters.([]interface{}) {
		if filter == nil {
			continue
		}
		f := filter.(map[string]interface{})
		if f["name"] == nil {
			continue
		}
		name := f["name"].(string)
		names = append(names, name)
	}
	
	fmt.Println("names: ", names)
	nodes, err := client.DescribeNodes(ctx, names)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(filterId)
	d.Set("nodes", flattenNodes(nodes))

	return diags
}

func makeListId(d *schema.ResourceData) (string, error) {
	idMap := map[string]interface{}{
		"filter":   d.Get("filter"),
	}

	result, err := json.Marshal(idMap)
	if err != nil {
		return "", err
	}

	hash := sha3.Sum512(result)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

func flattenNodes(nodes []service.Node) []map[string]interface{} {
	log.Debugf("Flattening nodes: %v", nodes)
	var nodesData []map[string]interface{}
	for _, node := range nodes {
		nodesData = append(nodesData, map[string]interface{}{
			"id": node.Id,
			"node": node.Node,
			"cores": node.Cores,
			"ssl_fingerprint": node.SslFingerprint,
			"memory": node.Memory,
			"total_disk_space": node.DiskSpace,
			"disks": flattenDisk(node.Disks),
			"network_interfaces": flattenNetworkInterfaces(node.NetworkInterfaces),
		})
	}
	return nodesData
}

func flattenDisk(disk []service.Disk) []interface{} {
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

func flattenNetworkInterfaces(networkInterface []service.NetworkInterface) []interface{} {
	var networkInterfaces []interface{}
	for _, networkInterface := range networkInterface {
		networkInterfaces = append(networkInterfaces, map[string]interface{}{
			"name": networkInterface.Name,
		})
	}
	return networkInterfaces
}
package bridges

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBridge() *schema.Resource {
	return &schema.Resource{
		Schema: bridgeDataSource,
	}
}

var filter = filters.FilterConfig{"node"}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkBridgeRead,
		Schema: map[string]*schema.Schema{
			"filter": filter.Schema(),
			"network_bridges": {
				Type:        schema.TypeList,
				Description: "The returned list of network bridges.",
				Computed:    true,
				Elem:        dataSourceBridge(),
			},
		},
	}
}

func dataSourceNetworkBridgeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*service.Proxmox)
	filterId, err := filters.MakeListId(d)
	if err != nil {
		return diag.Errorf("failed to generate filter id: %s", err)
	}
	nodes := filters.DetermineNodes(client, d)

	bridges := []service.NetworkBridge{}
	for _, node := range nodes {
		b, err := client.DescribeNetworkBridges(ctx, node)
		if err != nil {
			return diag.Errorf("failed to list bridges: %s", err)
		}
		bridges = append(bridges, b...)
	}

	d.SetId(filterId)
	d.Set("network_bridges", flattenBridges(bridges))

	return diags
}

func flattenBridges(bridges []service.NetworkBridge) []map[string]interface{} {
	var result []map[string]interface{}
	for _, bridge := range bridges {
		bridgeMap := map[string]interface{}{
			"id":   fmt.Sprintf("%s/%s", bridge.Node, bridge.Name),
			"node": bridge.Node,
			"name": bridge.Name,
			"active": bridge.Active,
			"autostart": bridge.Autostart,
			"vlan_aware": bridge.VLANAware,
			"interfaces": bridge.Interfaces,
		}

		if bridge.IPv4 != nil {
			bridgeMap["ipv4_address"] = bridge.IPv4.Address
			bridgeMap["ipv4_gateway"] = bridge.IPv4.Gateway
			bridgeMap["ipv4_netmask"] = bridge.IPv4.Netmask
		} else {
			bridgeMap["ipv4_address"] = nil
			bridgeMap["ipv4_gateway"] = nil
			bridgeMap["ipv4_netmask"] = nil
		}

		if bridge.IPv6 != nil {
			bridgeMap["ipv6_address"] = bridge.IPv6.Address
			bridgeMap["ipv6_gateway"] = bridge.IPv6.Gateway
			bridgeMap["ipv6_netmask"] = bridge.IPv6.Netmask
		}

		result = append(result, bridgeMap)
	}
	return result
}
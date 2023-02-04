package bonds

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBond() *schema.Resource {
	return &schema.Resource{
		Schema: bondDataSource,
	}
}

var filter = filters.FilterConfig{"node"}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkBondRead,
		Schema: map[string]*schema.Schema{
			"filter": filter.Schema(),
			"network_bonds": {
				Type:        schema.TypeList,
				Description: "The returned list of network bonds.",
				Computed:    true,
				Elem:        dataSourceBond(),
			},
		},
	}
}

func dataSourceNetworkBondRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*service.Proxmox)
	filterId, err := filters.MakeListId(d)
	if err != nil {
		return diag.Errorf("failed to generate filter id: %s", err)
	}
	nodes := filters.DetermineNodes(client, d)

	bonds := []service.NetworkBond{}
	for _, node := range nodes {
		b, err := client.DescribeNetworkBonds(ctx, node)
		if err != nil {
			return diag.Errorf("failed to list bonds: %s", err)
		}
		bonds = append(bonds, b...)
	}

	d.SetId(filterId)
	d.Set("network_bonds", flattenBonds(bonds))

	return diags
}

func flattenBonds(bonds []service.NetworkBond) []map[string]interface{} {
	var result []map[string]interface{}
	for _, bridge := range bonds {
		bridgeMap := map[string]interface{}{
			"id":   fmt.Sprintf("%s/%s", bridge.Node, bridge.Name),
			"node": bridge.Node,
			"name": bridge.Name,
			"active": bridge.Active,
			"autostart": bridge.Autostart,
			"hash_policy": bridge.HashPolicy,
			"mode": bridge.Mode,
			"mii_mon": bridge.MiiMon,
			"interfaces": bridge.Interfaces,
		}
		result = append(result, bridgeMap)
	}
	return result
}
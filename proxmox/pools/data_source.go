package pools

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePools() *schema.Resource {
	return &schema.Resource{
		Schema: poolDataSource,
	}
}

var filter = filters.FilterConfig{}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePoolsRead,
		Schema: map[string]*schema.Schema{
			"filter": filter.Schema(),
			"pools": {
				Type:        schema.TypeList,
				Description: "The returned list of pools.",
				Computed:    true,
				Elem:        dataSourcePools(),
			},
		},
	}
}

func dataSourcePoolsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*service.Proxmox)
	filterId, err := filters.MakeListId(d)
	if err != nil {
		return diag.Errorf("failed to generate filter id: %s", err)
	}

	pools, err := client.DescribePools(ctx)
	if err != nil {
		return diag.Errorf("failed to describe pools: %s", err)
	}

	d.SetId(filterId)
	d.Set("pools", flattenPools(pools))

	return diags
}

func flattenPools(pools []service.Pool) []map[string]interface{} {
	var result []map[string]interface{}
	for _, pool := range pools {
		result = append(result, map[string]interface{}{
			"id":   pool.Id,
			"comment": pool.Comment,
			"members": flattenMembers(pool.Members),
		})
	}
	return result
}

func flattenMembers(members []service.PoolMember) []map[string]interface{} {
	var result []map[string]interface{}
	for _, member := range members {
		result = append(result, map[string]interface{}{
			"id":   member.Id,
			"type": member.Type,
		})
	}
	return result
}
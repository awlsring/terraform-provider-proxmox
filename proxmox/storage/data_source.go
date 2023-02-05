package storage

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStorage() *schema.Resource {
	return &schema.Resource{
		Schema: storageDataSource,
	}
}

var filter = filters.FilterConfig{}
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceStorageRead,
		Schema: map[string]*schema.Schema{
			"filter": filter.Schema(),
			"storage": {
				Type:        schema.TypeList,
				Description: "The returned list of storage.",
				Computed:    true,
				Elem:        dataSourceStorage(),
			},
		},
	}
}

func dataSourceStorageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*service.Proxmox)
	filterId, err := filters.MakeListId(d)
	if err != nil {
		return diag.Errorf("failed to generate filter id: %s", err)
	}
	
	storage := []service.Storage{}
	s, err := client.DescribeStorage(context.Background())
	if err != nil {
		return diag.Errorf("failed to describe storage: %s", err)
	}
	storage = append(storage, s...)

	ls, err := client.DescribeLocalStorage(context.Background())
	if err != nil {
		return diag.Errorf("failed to describe local storage: %s", err)
	}
	storage = append(storage, ls...)


	d.SetId(filterId)
	d.Set("storage", flattenStorage(storage))
	
	return diags
}

func flattenStorage(storage []service.Storage) []map[string]interface{} {
	var result []map[string]interface{}
	for _, s := range storage {
		result = append(result, map[string]interface{}{
			"id": s.Id,
			"shared_nodes": s.SharedNodes,
			"shared": s.Shared,
			"local": s.Local,
			"source": s.Source,
			"content": s.Content,
			"size": s.Size,
			"type": s.Type,
		})
	}
	return result
}
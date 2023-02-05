package templates

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/vms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTemplate() *schema.Resource {
	return &schema.Resource{
		Schema: templateDataSource,
	}
}

var filter = filters.FilterConfig{"node"}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTemplateRead,
		Schema: map[string]*schema.Schema{
			"filter": filter.Schema(),
			"templates": {
				Type:        schema.TypeList,
				Description: "The returned list of templates.",
				Computed:    true,
				Elem:        dataSourceTemplate(),
			},
		},
	}
}

func dataSourceTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*service.Proxmox)
	filterId, err := filters.MakeListId(d)
	if err != nil {
		return diag.Errorf("failed to generate filter id: %s", err)
	}
	nodes := filters.DetermineNodes(client, d)

	templates := []vm.VirtualMachine{}
	for _, node := range nodes {
		t, err := client.DescribeTemplates(ctx, node)
		if err != nil {
			return diag.Errorf("failed to list templates: %s", err)
		}
		templates = append(templates, t...)
	}

	d.SetId(filterId)
	d.Set("templates", vms.FlattenVirtualMachines(templates))
	
	return diags
}
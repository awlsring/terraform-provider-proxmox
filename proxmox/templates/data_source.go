package templates

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
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

	templates := []service.VirtualMachineTemplate{}
	for _, node := range nodes {
		t, err := client.DescribeTemplates(ctx, node)
		if err != nil {
			return diag.Errorf("failed to list templates: %s", err)
		}
		templates = append(templates, t...)
	}

	d.SetId(filterId)
	d.Set("templates", flattenTemplates(templates))
	
	return diags
}

func flattenTemplates(templates []service.VirtualMachineTemplate) []map[string]interface{} {
	var result []map[string]interface{}
	for _, template := range templates {
		result = append(result, map[string]interface{}{
			"id": template.Id,
			"node": template.Node,
			"name": template.Name,
			"agent": template.Agent,
			"cores": template.Cores,
			"memory": template.Memory,
			"disks": flattenVirtualDisks(template.VirtualDisks),
			"network_interfaces": flattenNetworkInterfaces(template.VirtualNetworkDevices),
		})
	}
	return result
}

func flattenVirtualDisks(disks []service.VirtualDisk) []interface{} {
	var result []interface{}
	for _, disk := range disks {
		result = append(result, map[string]interface{}{
			"storage": disk.Storage,
			"type":  disk.Type,
			"position":  disk.Position,
			"size":    disk.Size,
			"discard":  disk.Discard,
		})
	}
	return result
}

func flattenNetworkInterfaces(networkInterfaces []service.VirtualNetworkDevice) []interface{} {
	var result []interface{}
	for _, networkInterface := range networkInterfaces {
		result = append(result, map[string]interface{}{
			"bridge": networkInterface.Bridge,
			"vlan":  networkInterface.Vlan,
			"model":  networkInterface.Model,
			"mac":  networkInterface.Mac,
			"position":  networkInterface.Position,
			"firewall":  networkInterface.FirewallEnabled,
		})
	}
	return result
}
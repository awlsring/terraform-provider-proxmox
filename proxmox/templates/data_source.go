package templates

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTemplate() *schema.Resource {
	return &schema.Resource{
		Schema: templateDataSource,
	}
}

var filters = filter.FilterConfig{"node"}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTemplateRead,
		Schema: map[string]*schema.Schema{
			"filter": filters.Schema(),
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
	filterId, err := filter.MakeListId(d)
	if err != nil {
		return diag.Errorf("failed to generate filter id: %s", err)
	}

	// there is probably a prettier way to do this
	nodes := []string{}
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
		switch name {
		case "node":
			if f["value"] == nil {
				continue
			}
			nodes = append(nodes, f["value"].(string))
		}
	}

	// If no nodes are specified, list all nodes and find templates on all
	if len(nodes) == 0 {
		nodeSummaries, err := client.ListNodes(context.Background())
		if err != nil {
			return diag.Errorf("failed to list nodes: %s", err)
		}
		for _, nodeSummary := range nodeSummaries {
			nodes = append(nodes, nodeSummary.Node)
		}
	}

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
package vms

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Schema: virtualMachineDataSource,
	}
}

var filter = filters.FilterConfig{"node"}

func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVirtualMachineRead,
		Schema: map[string]*schema.Schema{
			"filter": filter.Schema(),
			"virtual_machines": {
				Type:        schema.TypeList,
				Description: "The returned list of virtual machines.",
				Computed:    true,
				Elem:        dataSourceVirtualMachine(),
			},
		},
	}
}

func dataSourceVirtualMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*service.Proxmox)
	filterId, err := filters.MakeListId(d)
	if err != nil {
		return diag.Errorf("failed to generate filter id: %s", err)
	}
	nodes := filters.DetermineNodes(client, d)

	templates := []vm.VirtualMachine{}
	for _, node := range nodes {
		t, err := client.DescribeVirtualMachines(ctx, node)
		if err != nil {
			return diag.Errorf("failed to list virtual machines: %s", err)
		}
		templates = append(templates, t...)
	}

	d.SetId(filterId)
	d.Set("virtual_machines", FlattenVirtualMachines(templates))
	
	return diags
}

func FlattenVirtualMachines(templates []vm.VirtualMachine) []map[string]interface{} {
	var result []map[string]interface{}
	for _, template := range templates {
		result = append(result, map[string]interface{}{
			"id": template.Id,
			"node": template.Node,
			"name": template.Name,
			"agent": template.Agent,
			"cores": template.Cores,
			"memory": template.Memory,
			"tags": template.Tags,
			"disks": flattenVirtualDisks(template.VirtualDisks),
			"network_interfaces": flattenNetworkInterfaces(template.VirtualNetworkDevices),
		})
	}
	return result
}

func flattenVirtualDisks(disks []vm.VirtualDisk) []interface{} {
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

func flattenNetworkInterfaces(networkInterfaces []vm.VirtualNetworkDevice) []interface{} {
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
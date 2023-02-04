package filters

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DetermineNodes(client *service.Proxmox, d *schema.ResourceData) []string {
	nodes := []string{}
	filters := d.Get("filter")

	j, _ := json.Marshal(filters)
	ioutil.WriteFile("filter.json", j, 0644)

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
			if f["values"] == nil {
				continue
			}
			l := f["values"].([]interface{})
			for _, v := range l {
				nodes = append(nodes, v.(string))
			}
		}
	}

	if len(nodes) == 0 {
		nodeSummaries, err := client.ListNodes(context.Background())
		if err != nil {
			return nil
		}
		for _, nodeSummary := range nodeSummaries {
			nodes = append(nodes, nodeSummary.Node)
		}
	}

	return nodes
}
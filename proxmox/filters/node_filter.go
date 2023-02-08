package filters

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
)

func DetermineNode(client *service.Proxmox, filters []FilterModel) []string {
	nodes := []string{}

	for _, filter := range filters {
		if filter.Name.String() == "node" {
			for _, v := range filter.Values {
				nodes = append(nodes, v.String())
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

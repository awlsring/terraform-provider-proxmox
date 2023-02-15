package storage

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func DetermineContentTypes(ctx context.Context, c types.List, defaults []string) ([]string, error) {
	if c.IsNull() || c.IsUnknown() {
		return defaults, nil
	}
	tflog.Debug(ctx, "content types is not null or unknown")
	contentTypes := []string{}
	for _, contentType := range c.Elements() {
		v, err := contentType.ToTerraformValue(ctx)
		if err != nil {
			return nil, err
		}
		var c string
		v.As(&c)
		contentTypes = append(contentTypes, c)
	}
	return contentTypes, nil
}

func DetermineNodes(ctx context.Context, client *service.Proxmox, n types.List) ([]string, error) {
	if n.IsNull() || n.IsUnknown() {
		tflog.Debug(ctx, "nodes is null or unknown")
		nodes, err := client.ListNodesNames(ctx)
		if err != nil {
			return nil, err
		}
		return nodes, nil
	}
	tflog.Debug(ctx, "nodes is not null or unknown")
	nodes := []string{}
	for _, node := range n.Elements() {
		v, err := node.ToTerraformValue(ctx)
		if err != nil {
			return nil, err
		}
		var n string
		v.As(&n)
		nodes = append(nodes, n)
	}
	return nodes, nil
}

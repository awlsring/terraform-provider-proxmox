package service

import (
	"context"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
)

func (c *Proxmox) ListTemplates(ctx context.Context, node string) ([]proxmox.VirtualMachineSummary, error) {
	request := c.client.ListVirtualMachines(ctx, node)
	resp, _, err := c.client.ListVirtualMachinesExecute(request)
	if err != nil {
		return nil, err
	}

	templateSummaries := []proxmox.VirtualMachineSummary{}
	for _, vmSummary := range resp.Data {
		if vmSummary.HasTemplate() {
			if *vmSummary.Template == 1 {
				templateSummaries = append(templateSummaries, vmSummary)
			}
		}
	}

	return templateSummaries, nil
}

func (c *Proxmox) DescribeTemplate(ctx context.Context, node string, vmId int) (*VirtualMachine, error) {
	templates, err := c.ListTemplates(ctx, node)
	if err != nil {
		return nil, err
	}

	for _, vm := range templates {
		id := int(vm.Vmid)

		if id == vmId {
			cf, err := c.DescribeVirtualMachine(ctx, node, id)
			if err != nil {
				return nil, err
			}
			return cf, nil
		}
	}

	return nil, fmt.Errorf("template not found: %v", vmId)
}

func (c *Proxmox) DescribeTemplateFromName(ctx context.Context, node string, name string) (*VirtualMachine, error) {
	templates, err := c.ListTemplates(ctx, node)
	if err != nil {
		return nil, err
	}

	for _, vm := range templates {
		id := int(vm.Vmid)

		if vm.HasName() {
			if *vm.Name == name {
				cf, err := c.DescribeVirtualMachine(ctx, node, id)
				if err != nil {
					return nil, err
				}
				return cf, nil
			}
		}
	}

	return nil, fmt.Errorf("template not found: %s", name)
}

func (c *Proxmox) DescribeTemplates(ctx context.Context, node string) ([]*VirtualMachine, error) {
	templates, err := c.ListTemplates(ctx, node)
	if err != nil {
		return nil, err
	}

	var templateList []*VirtualMachine
	for _, vm := range templates {
		id := int(vm.Vmid)
		cf, err := c.DescribeVirtualMachine(ctx, node, id)
		if err != nil {
			return nil, err
		}
		templateList = append(templateList, cf)
	}

	return templateList, nil
}

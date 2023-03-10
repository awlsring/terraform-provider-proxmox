package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
)

func (c *Proxmox) GetVirtualMachineConfiguration(ctx context.Context, node string, vmId int) (*proxmox.VirtualMachineConfigurationSummary, error) {
	vmIdStr := strconv.Itoa(vmId)
	request := c.client.GetVirtualMachineConfiguration(ctx, node, vmIdStr)
	resp, _, err := c.client.GetVirtualMachineConfigurationExecute(request)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Proxmox) DeleteVirtualMachine(ctx context.Context, node string, vmid int) error {
	vmId := strconv.Itoa(vmid)
	request := c.client.DeleteVirtualMachine(ctx, node, vmId)
	_, h, err := c.client.DeleteVirtualMachineExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}
	return nil
}

func (c *Proxmox) DescribeVirtualMachinesOld(ctx context.Context, node string) ([]vm.VirtualMachine, error) {
	vms, err := c.ListVirtualMachines(ctx, node)
	if err != nil {
		panic(err)
	}

	virtualMachines := []vm.VirtualMachine{}
	for _, vmSummary := range vms {
		virtualMachine, err := c.vmFromSummary(ctx, node, vmSummary)
		if err != nil {
			return nil, err
		}

		virtualMachines = append(virtualMachines, *virtualMachine)
	}

	return virtualMachines, nil
}

func (c *Proxmox) vmFromSummary(ctx context.Context, node string, summary proxmox.VirtualMachineSummary) (*vm.VirtualMachine, error) {
	vmId := int(summary.Vmid)
	vmConfig, err := c.GetVirtualMachineConfiguration(ctx, node, vmId)
	if err != nil {
		return nil, err
	}

	virtualDisks, err := vm.ExtractDisksFromConfig(vmConfig)
	if err != nil {
		return nil, err
	}

	virtualNics, err := vm.ExtractNicsFromConfig(vmConfig)
	if err != nil {
		return nil, err
	}

	virtualMachine := vm.VirtualMachine{
		Id:                    vmId,
		Node:                  node,
		VirtualDisks:          virtualDisks,
		VirtualNetworkDevices: virtualNics,
	}

	if summary.HasTags() {
		virtualMachine.Tags = StringSemiColonPtrListToSlice(summary.Tags)
	}

	if summary.HasName() {
		virtualMachine.Name = *summary.Name
	}

	if vmConfig.HasMemory() {
		virtualMachine.Memory = int64(*vmConfig.Memory)
	}

	if vmConfig.HasCores() {
		virtualMachine.Cores = int(*vmConfig.Cores)
	}

	if vmConfig.HasAgent() {
		agentStr := *vmConfig.Agent
		if strings.Contains(agentStr, "1") {
			virtualMachine.Agent = true
		}
	}

	return &virtualMachine, nil
}

func (c *Proxmox) ListVirtualMachines(ctx context.Context, node string) ([]proxmox.VirtualMachineSummary, error) {
	request := c.client.ListVirtualMachines(ctx, node)
	resp, _, err := c.client.ListVirtualMachinesExecute(request)
	if err != nil {
		return nil, err
	}

	vmSummaries := []proxmox.VirtualMachineSummary{}
	for _, vmSummary := range resp.Data {
		if vmSummary.HasTemplate() {
			if *vmSummary.Template == 1 {
				continue
			}
		}
		vmSummaries = append(vmSummaries, vmSummary)
	}

	return vmSummaries, nil
}

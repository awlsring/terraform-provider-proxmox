package service

import (
	"context"
	"strconv"

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

func (c *Proxmox) GetVirtualMachineConfiguration(ctx context.Context, node string, vmId int) (*proxmox.VirtualMachineConfigurationSummary, error) {
	vmIdStr := strconv.Itoa(vmId)
	request := c.client.GetVirtualMachineConfiguration(ctx, node, vmIdStr)
	resp, _, err := c.client.GetVirtualMachineConfigurationExecute(request)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

type VirtualMachineTemplate struct {
	Id int
	Node string
	Name string
	Memory int64
	Bios VirtualBios
	Cores int
	VirtualDisks []VirtualDisk
	VirtualNetworkDevices []VirtualNetworkDevice
}

type VirtualBios string
const (
	VIRTUAL_BIOS_SEABIOS VirtualBios = "seabios"
	VIRTUAL_BIOS_OVMF VirtualBios = "ovmf"
)
func (e *VirtualBios) IsValid() (bool) {
	switch *e {
	case VIRTUAL_BIOS_SEABIOS, VIRTUAL_BIOS_OVMF:
		return true
	}
	return false
}

type VirtualNetworkDevice struct {
	Bridge string
	Vlan int
	Model VirtualNetworkDeviceModel
	Mac string
	FirewallEnabled bool
}

type VirtualNetworkDeviceModel string
const (
	VIRTUAL_NIC_INTEL_E1000 VirtualNetworkDeviceModel = "e1000"
	VIRTUAL_NIC_VIRTIO VirtualNetworkDeviceModel = "virtio"
	VIRTUAL_NIC_REALTEK_RTL8139 VirtualNetworkDeviceModel = "rtl8139"
	VIRTUAL_NIC_VMWARE_VMXNET3 VirtualNetworkDeviceModel = "vmxnet3"
)
func (e *VirtualNetworkDeviceModel) IsValid() (bool) {
	switch *e {
	case VIRTUAL_NIC_INTEL_E1000, VIRTUAL_NIC_VIRTIO, VIRTUAL_NIC_REALTEK_RTL8139, VIRTUAL_NIC_VMWARE_VMXNET3:
		return true
	}
	return false
}

type VirtualDisk struct {
	Storage string
	Type VirtualDiskType
	Position string //virtio0, virtio1, scsi0, scsi1, ide0, ide1, sata0, sata1, etc
	Size int64
	Discard bool
}

type VirtualDiskType string
const (
	VIRTUAL_DISK_SCSI VirtualDiskType = "scsi"
	VIRTUAL_DISK_VIRTIO_D VirtualDiskType = "virtio"
	VIRTUAL_DISK_SATA VirtualDiskType = "sata"
	VIRTUAL_DISK_IDE  VirtualDiskType= "ide"
)
func (e *VirtualDiskType) IsValid() (bool) {
	switch *e {
	case VIRTUAL_DISK_SCSI, VIRTUAL_DISK_VIRTIO_D, VIRTUAL_DISK_SATA, VIRTUAL_DISK_IDE:
		return true
	}
	return false
}

func (c *Proxmox) DescribeTemplates(ctx context.Context, node string) ([]VirtualMachineTemplate, error) {
	templates, err := c.ListTemplates(ctx, node)
	if err != nil {
		return nil, err
	}

	virtualMachineTemplates := []VirtualMachineTemplate{}
	for _, templateSummary := range templates {
		vmId:= int(templateSummary.Vmid)
		if err != nil {
			return nil, err
		}

		vmConfig, err := c.GetVirtualMachineConfiguration(ctx, node, vmId)
		if err != nil {
			return nil, err
		}

		virtualDisks, err := extractDisksFromConfig(vmConfig)
		if err != nil {
			return nil, err
		}

		virtualNics, err := extractNicsFromConfig(vmConfig)
		if err != nil {
			return nil, err
		}

		virtualMachineTemplate := VirtualMachineTemplate{
			Id: vmId,
			Node: node,
			Name: *templateSummary.Name,
			Memory: int64(*vmConfig.Memory),
			Bios: VirtualBios(*vmConfig.Bios),
			Cores: int(*vmConfig.Cores),
			VirtualDisks: virtualDisks,
			VirtualNetworkDevices: virtualNics,
		}

		virtualMachineTemplates = append(virtualMachineTemplates, virtualMachineTemplate)
	}

	return virtualMachineTemplates, nil
}
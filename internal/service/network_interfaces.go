package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type IP struct {
	Address string
	Netmask string
}

func (c *Proxmox) createNetworkInterface(ctx context.Context, node string, content proxmox.CreateNetworkInterfaceRequestContent) error {
	request := c.client.CreateNetworkInterface(ctx, node)
	request = request.CreateNetworkInterfaceRequestContent(content)

	h, err := c.client.CreateNetworkInterfaceExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	err = c.ApplyChanges(ctx, node)
	if err != nil {
		return err
	}

	return nil
}

func (c *Proxmox) modifyNetworkInterface(ctx context.Context, node string, iface string, content proxmox.UpdateNetworkInterfaceRequestContent) error {
	request := c.client.UpdateNetworkInterface(ctx, node, iface)
	request = request.UpdateNetworkInterfaceRequestContent(content)

	h, err := c.client.UpdateNetworkInterfaceExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	err = c.ApplyChanges(ctx, node)
	if err != nil {
		return err
	}

	return nil
}

func (c *Proxmox) ApplyChanges(ctx context.Context, node string) error {
	request := c.client.ApplyNetworkInterfaceConfiguration(ctx, node)
	_, h, err := c.client.ApplyNetworkInterfaceConfigurationExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	return nil
}

func (c *Proxmox) DeleteNetworkInterface(ctx context.Context, node string, name string) error {
	request := c.client.DeleteNetworkInterface(ctx, node, name)
	h, err := c.client.DeleteNetworkInterfaceExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	err = c.ApplyChanges(ctx, node)
	if err != nil {
		return err
	}

	return nil
}

func Ipv4FromInterface(iface proxmox.NetworkInterfaceSummary) *IP {
	nic := IP{}

	if iface.HasAddress() {
		nic.Address = *iface.Address
	}

	if iface.HasNetmask() {
		nic.Netmask = *iface.Netmask
	}

	if nic.Address == "" && nic.Netmask == "" {
		return nil
	}
	return &nic
}

func Ipv6FromInterface(iface proxmox.NetworkInterfaceSummary) *IP {
	nic := IP{}

	if iface.HasAddress6() {
		nic.Address = *iface.Address6
	}

	if iface.HasNetmask6() {
		nic.Netmask = *iface.Netmask6
	}

	if nic.Address == "" && nic.Netmask == "" {
		return nil
	}
	return &nic
}

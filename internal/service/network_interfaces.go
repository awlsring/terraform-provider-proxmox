package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/errors"
)

type NetworkInterfaceIpv4 struct {
	Address string
	Netmask string
	Gateway string
}

type NetworkInterfaceIpv6 struct {
	Address string
	Netmask string
	Gateway string
}

type NetworkBridge struct {
	Active     bool
	Autostart  bool
	VLANAware  bool
	Interfaces []string
	Name       string
	Node       string
	IPv4       *NetworkInterfaceIpv4
	IPv6       *NetworkInterfaceIpv6
}

type NetworkBond struct {
	Active      bool
	Autostart   bool
	HashPolicy  *proxmox.NetworkInterfaceBondHashPolicy
	BondPrimary *string
	Mode        proxmox.NetworkInterfaceBondMode
	MiiMon      *string
	Interfaces  []string
	Name        string
	Node        string
	Comments    *string
}

func (c *Proxmox) DescribeNetworkBridges(ctx context.Context, node string) ([]NetworkBridge, error) {
	bridges, err := c.ListNetworkBridges(ctx, node)
	if err != nil {
		return nil, err
	}

	networkBridges := []NetworkBridge{}
	for _, bridge := range bridges {
		iface := NetworkBridge{
			Active:     BooleanIntegerConversion(bridge.Active),
			Autostart:  BooleanIntegerConversion(bridge.Autostart),
			VLANAware:  BooleanIntegerConversion(bridge.BridgeVlanAware),
			Interfaces: StringSpacePtrListToSlice(bridge.BridgePorts),
			Name:       bridge.Iface,
			Node:       node,
			IPv4:       Ipv4FromInterface(bridge),
			IPv6:       Ipv6FromInterface(bridge),
		}
		networkBridges = append(networkBridges, iface)
	}
	return networkBridges, nil
}

type CreateNetworkBondInput struct {
	Interfaces []string
	Name       string
	Node       string
	Mode       proxmox.NetworkInterfaceBondMode
	// optional
	HashPolicy  *proxmox.NetworkInterfaceBondHashPolicy
	BondPrimary *string
	AutoStart   *bool
	Comments    *string
}

func (c *Proxmox) CreateNetworkBond(ctx context.Context, input *CreateNetworkBondInput) error {
	request := c.client.CreateNetworkInterface(ctx, input.Node)
	request = request.CreateNetworkInterfaceRequestContent(
		proxmox.CreateNetworkInterfaceRequestContent{
			Type:               "bond",
			Iface:              input.Name,
			BondMode:           &input.Mode,
			BondXmitHashPolicy: input.HashPolicy,
			BondPrimary:        input.BondPrimary,
			Autostart:          input.AutoStart,
			Comments:           input.Comments,
			Slaves:             StringSliceToStringSpacePtr(input.Interfaces),
		},
	)

	h, err := c.client.CreateNetworkInterfaceExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	err = c.ApplyChanges(ctx, input.Node)
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

func (c *Proxmox) DescribeNetworkBonds(ctx context.Context, node string) ([]NetworkBond, error) {
	bonds, err := c.ListNetworkBonds(ctx, node)
	if err != nil {
		return nil, err
	}

	networkBonds := []NetworkBond{}
	for _, bond := range bonds {
		iface := NetworkBond{
			Active:     BooleanIntegerConversion(bond.Active),
			Autostart:  BooleanIntegerConversion(bond.Autostart),
			Interfaces: StringSpacePtrListToSlice(bond.Slaves),
			Name:       bond.Iface,
			Node:       node,
		}

		if bond.HasBondMiimon() {
			iface.MiiMon = bond.BondMiimon
		}

		if bond.HasBondXmitHashPolicy() {
			iface.HashPolicy = bond.BondXmitHashPolicy
		}

		// if bond.HasBondPrimary() {
		// 	iface.BondPrimary = *bond.BondPrimary
		// }

		if bond.HasBondMode() {
			iface.Mode = *bond.BondMode
		}

		networkBonds = append(networkBonds, iface)
	}
	return networkBonds, nil
}

func (c *Proxmox) ListNetworkBridges(ctx context.Context, node string) ([]proxmox.NetworkInterfaceSummary, error) {
	request := c.client.ListNetworkInterfaces(ctx, node)
	request = request.Type_("bridge")
	resp, _, err := c.client.ListNetworkInterfacesExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Proxmox) ListNetworkBonds(ctx context.Context, node string) ([]proxmox.NetworkInterfaceSummary, error) {
	request := c.client.ListNetworkInterfaces(ctx, node)
	request = request.Type_("bond")
	resp, _, err := c.client.ListNetworkInterfacesExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Proxmox) GetNetworkBond(ctx context.Context, node string, name string) (*NetworkBond, error) {
	request := c.client.GetNetworkInterface(ctx, node, name)
	resp, _, err := c.client.GetNetworkInterfaceExecute(request)
	if err != nil {
		return nil, err
	}

	iface := NetworkBond{
		Active:     BooleanIntegerConversion(resp.Data.Active),
		Autostart:  BooleanIntegerConversion(resp.Data.Autostart),
		Interfaces: StringSpacePtrListToSlice(resp.Data.Slaves),
		Name:       name,
		Node:       node,
	}

	if resp.Data.HasBondMiimon() {
		iface.MiiMon = resp.Data.BondMiimon
	}

	if resp.Data.HasBondXmitHashPolicy() {
		iface.HashPolicy = resp.Data.BondXmitHashPolicy
	}

	if resp.Data.HasBondPrimary() {
		iface.BondPrimary = resp.Data.BondPrimary
	}

	if resp.Data.HasBondMode() {
		iface.Mode = *resp.Data.BondMode
	}

	if resp.Data.HasComments() {
		iface.Comments = resp.Data.Comments
	}

	return &iface, nil
}

type UpdateNetworkBondInput struct {
	Interfaces []string
	Name       string
	Node       string
	Mode       proxmox.NetworkInterfaceBondMode
	// optional
	HashPolicy  *proxmox.NetworkInterfaceBondHashPolicy
	BondPrimary *string
	AutoStart   *bool
	Comments    *string
}

func (c *Proxmox) UpdateNetworkBond(ctx context.Context, input *UpdateNetworkBondInput) error {
	request := c.client.UpdateNetworkInterface(ctx, input.Node, input.Name)
	request = request.UpdateNetworkInterfaceRequestContent(
		proxmox.UpdateNetworkInterfaceRequestContent{
			Type:               "bond",
			BondMode:           &input.Mode,
			BondXmitHashPolicy: input.HashPolicy,
			BondPrimary:        input.BondPrimary,
			Autostart:          input.AutoStart,
			Comments:           input.Comments,
			Slaves:             StringSliceToStringSpacePtr(input.Interfaces),
		},
	)

	h, err := c.client.UpdateNetworkInterfaceExecute(request)
	if err != nil {
		return errors.ApiError(h, err)
	}

	err = c.ApplyChanges(ctx, input.Node)
	if err != nil {
		return err
	}

	return nil
}

func (c *Proxmox) DeleteNetworkBond(ctx context.Context, node string, name string) error {
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

func Ipv4FromInterface(iface proxmox.NetworkInterfaceSummary) *NetworkInterfaceIpv4 {
	nic := NetworkInterfaceIpv4{}

	if iface.HasAddress() {
		nic.Address = *iface.Address
	}

	if iface.HasNetmask() {
		nic.Netmask = *iface.Netmask
	}

	if iface.HasGateway() {
		nic.Gateway = *iface.Gateway
	}

	if nic.Address == "" && nic.Netmask == "" && nic.Gateway == "" {
		return nil
	}
	return &nic
}

func Ipv6FromInterface(iface proxmox.NetworkInterfaceSummary) *NetworkInterfaceIpv6 {
	nic := NetworkInterfaceIpv6{}

	if iface.HasAddress6() {
		nic.Address = *iface.Address6
	}

	if iface.HasNetmask6() {
		nic.Netmask = *iface.Netmask6
	}

	if iface.HasGateway6() {
		nic.Gateway = *iface.Gateway6
	}

	if nic.Address == "" && nic.Netmask == "" && nic.Gateway == "" {
		return nil
	}
	return &nic
}

package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
)

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
	IPv4        *IP
	IPv6        *IP
	IPv4Gateway *string
	IPv6Gateway *string
	Comments    *string
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
	IPv4        *IP
	IPv6        *IP
	IPv4Gateway *string
	IPv6Gateway *string
	Comments    *string
}

func (c *Proxmox) CreateNetworkBond(ctx context.Context, input *CreateNetworkBondInput) error {
	content := proxmox.CreateNetworkInterfaceRequestContent{
		Type:               "bond",
		Iface:              input.Name,
		BondMode:           &input.Mode,
		BondXmitHashPolicy: input.HashPolicy,
		BondPrimary:        input.BondPrimary,
		Autostart:          input.AutoStart,
		Comments:           input.Comments,
		Slaves:             StringSliceToStringSpacePtr(input.Interfaces),
		Gateway:            input.IPv4Gateway,
		Gateway6:           input.IPv6Gateway,
	}
	if input.IPv4 != nil {
		cidr := fmt.Sprintf("%s/%s", input.IPv4.Address, input.IPv4.Netmask)
		content.Cidr = &cidr
	}

	if input.IPv6 != nil {
		cidr := fmt.Sprintf("%s/%s", input.IPv6.Address, input.IPv6.Netmask)
		content.Cidr6 = &cidr
	}

	return c.createNetworkInterface(ctx, input.Node, content)
}

func summaryToBond(node string, name string, summary proxmox.NetworkInterfaceSummary) NetworkBond {
	bond := NetworkBond{
		Active:      BooleanIntegerConversion(summary.Active),
		Autostart:   BooleanIntegerConversion(summary.Autostart),
		Interfaces:  StringSpacePtrListToSlice(summary.Slaves),
		Name:        name,
		Node:        node,
		Mode:        *summary.BondMode,
		IPv4:        Ipv4FromInterface(summary),
		IPv6:        Ipv6FromInterface(summary),
		IPv4Gateway: summary.Gateway,
		IPv6Gateway: summary.Gateway6,
		MiiMon:      summary.BondMiimon,
		HashPolicy:  summary.BondXmitHashPolicy,
		BondPrimary: summary.BondPrimary,
	}

	if summary.HasComments() {
		trim := strings.TrimSpace(*summary.Comments)
		bond.Comments = &trim
	}

	return bond
}

func (c *Proxmox) DescribeNetworkBonds(ctx context.Context, node string) ([]NetworkBond, error) {
	bonds, err := c.ListNetworkBonds(ctx, node)
	if err != nil {
		return nil, err
	}

	networkBonds := []NetworkBond{}
	for _, bond := range bonds {
		iface := summaryToBond(node, bond.Iface, bond)
		networkBonds = append(networkBonds, iface)
	}
	return networkBonds, nil
}

func (c *Proxmox) GetNetworkBond(ctx context.Context, node string, name string) (*NetworkBond, error) {
	request := c.client.GetNetworkInterface(ctx, node, name)
	resp, _, err := c.client.GetNetworkInterfaceExecute(request)
	if err != nil {
		return nil, err
	}

	iface := summaryToBond(node, name, resp.Data)

	return &iface, nil
}

type UpdateNetworkBondInput struct {
	Interfaces []string
	Name       string
	Node       string
	Mode       proxmox.NetworkInterfaceBondMode
	// optional
	IPv4        *IP
	IPv6        *IP
	IPv4Gateway *string
	IPv6Gateway *string
	HashPolicy  *proxmox.NetworkInterfaceBondHashPolicy
	BondPrimary *string
	AutoStart   *bool
	Comments    *string
}

func (c *Proxmox) UpdateNetworkBond(ctx context.Context, input *UpdateNetworkBondInput) error {
	content := proxmox.UpdateNetworkInterfaceRequestContent{
		Type:               "bond",
		BondMode:           &input.Mode,
		BondXmitHashPolicy: input.HashPolicy,
		BondPrimary:        input.BondPrimary,
		Autostart:          input.AutoStart,
		Comments:           input.Comments,
		Slaves:             StringSliceToStringSpacePtr(input.Interfaces),
	}
	delete := []string{}

	if input.IPv4 != nil {
		cidr := fmt.Sprintf("%s/%s", input.IPv4.Address, input.IPv4.Netmask)
		content.Cidr = &cidr
	} else {
		delete = append(delete, []string{"cidr", "address", "netmask"}...)
	}

	if input.IPv6 != nil {
		cidr := fmt.Sprintf("%s/%s", input.IPv6.Address, input.IPv6.Netmask)
		content.Cidr6 = &cidr
	} else {
		delete = append(delete, []string{"cidr6", "address6", "netmask6"}...)
	}

	if input.IPv4Gateway != nil {
		content.Gateway = input.IPv4Gateway
	} else {
		delete = append(delete, "gateway")
	}

	if input.IPv6Gateway != nil {
		content.Gateway6 = input.IPv6Gateway
	} else {
		delete = append(delete, "gateway6")
	}

	if len(delete) > 0 {
		content.Delete = StringSliceToStringSpacePtr(delete)
	}

	return c.modifyNetworkInterface(ctx, input.Node, input.Name, content)
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

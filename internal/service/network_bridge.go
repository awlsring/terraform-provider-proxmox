package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
)

type NetworkBridge struct {
	Active      bool
	Autostart   bool
	VLANAware   bool
	Interfaces  []string
	Name        string
	Node        string
	IPv4        *IP
	IPv6        *IP
	IPv4Gateway *string
	IPv6Gateway *string
	Comments    *string
}

func (c *Proxmox) DescribeNetworkBridges(ctx context.Context, node string) ([]NetworkBridge, error) {
	bridges, err := c.ListNetworkBridges(ctx, node)
	if err != nil {
		return nil, err
	}

	networkBridges := []NetworkBridge{}
	for _, bridge := range bridges {
		iface := summaryToBridge(node, bridge.Iface, bridge)
		networkBridges = append(networkBridges, iface)
	}
	return networkBridges, nil
}

type CreateNetworkBridgeInput struct {
	Interfaces []string
	Name       string
	Node       string
	// optional
	IPv4        *IP
	IPv6        *IP
	IPv4Gateway *string
	IPv6Gateway *string
	VLANAware   *bool
	AutoStart   *bool
	Comments    *string
}

func (c *Proxmox) CreateNetworkBridge(ctx context.Context, input *CreateNetworkBridgeInput) error {
	content := proxmox.CreateNetworkInterfaceRequestContent{
		Type:            "bridge",
		Iface:           input.Name,
		Autostart:       input.AutoStart,
		Comments:        input.Comments,
		BridgePorts:     StringSliceToStringSpacePtr(input.Interfaces),
		BridgeVlanAware: input.VLANAware,
		Gateway:         input.IPv4Gateway,
		Gateway6:        input.IPv6Gateway,
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

func summaryToBridge(node string, name string, summary proxmox.NetworkInterfaceSummary) NetworkBridge {
	bridge := NetworkBridge{
		Active:      BooleanIntegerConversion(summary.Active),
		Autostart:   BooleanIntegerConversion(summary.Autostart),
		VLANAware:   BooleanIntegerConversion(summary.BridgeVlanAware),
		Interfaces:  StringSpacePtrListToSlice(summary.BridgePorts),
		Name:        name,
		Node:        node,
		IPv4:        Ipv4FromInterface(summary),
		IPv6:        Ipv6FromInterface(summary),
		IPv4Gateway: summary.Gateway,
		IPv6Gateway: summary.Gateway6,
	}

	if summary.HasComments() {
		trim := strings.TrimSpace(*summary.Comments)
		bridge.Comments = &trim
	}

	return bridge
}

func (c *Proxmox) GetNetworkBridge(ctx context.Context, node string, name string) (*NetworkBridge, error) {
	request := c.client.GetNetworkInterface(ctx, node, name)
	resp, _, err := c.client.GetNetworkInterfaceExecute(request)
	if err != nil {
		return nil, err
	}

	bridge := resp.Data

	iface := summaryToBridge(node, name, bridge)

	return &iface, nil
}

type UpdateNetworkBridgeInput struct {
	Interfaces []string
	Name       string
	Node       string
	// optional
	IPv4        *IP
	IPv6        *IP
	IPv4Gateway *string
	IPv6Gateway *string
	VLANAware   *bool
	AutoStart   *bool
	Comments    *string
}

func (c *Proxmox) UpdateNetworkBridge(ctx context.Context, input *UpdateNetworkBridgeInput) error {
	content := proxmox.UpdateNetworkInterfaceRequestContent{
		Type:            "bridge",
		Autostart:       input.AutoStart,
		Comments:        input.Comments,
		BridgePorts:     StringSliceToStringSpacePtr(input.Interfaces),
		BridgeVlanAware: input.VLANAware,
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

func (c *Proxmox) ListNetworkBridges(ctx context.Context, node string) ([]proxmox.NetworkInterfaceSummary, error) {
	request := c.client.ListNetworkInterfaces(ctx, node)
	request = request.Type_("bridge")
	resp, _, err := c.client.ListNetworkInterfacesExecute(request)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

package service

import (
	"context"

	"github.com/awlsring/proxmox-go/proxmox"
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
	Active bool
	Autostart bool
	VLANAware bool
	Interfaces []string
	Name string
	Node string
	IPv4 *NetworkInterfaceIpv4
	IPv6 *NetworkInterfaceIpv6
}

func (c *Proxmox) DescribeNetworkBridges(ctx context.Context, node string) ([]NetworkBridge, error) {
	bridges, err := c.ListNetworkBridges(ctx, node)
	if err != nil {
		return nil, err
	}

	networkBridges := []NetworkBridge{}
	for _, bridge := range bridges {
		iface := NetworkBridge{
			Active: BooleanIntegerConversion(bridge.Active),
			Autostart: BooleanIntegerConversion(bridge.Autostart),
			VLANAware: BooleanIntegerConversion(bridge.BridgeVlanAware),
			Interfaces: StringSpacePtrListToSlice(bridge.BridgePorts),
			Name: bridge.Iface,
			Node: node,
			IPv4: Ipv4FromInterface(bridge),
			IPv6: Ipv6FromInterface(bridge),
		}
		networkBridges = append(networkBridges, iface)
	}
	return networkBridges, nil
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

func Ipv4FromInterface(iface proxmox.NetworkInterfaceSummary) (*NetworkInterfaceIpv4) {
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

func Ipv6FromInterface(iface proxmox.NetworkInterfaceSummary) (*NetworkInterfaceIpv6) {
	nic := NetworkInterfaceIpv6{}

	if iface.HasAddress6() {
		nic.Address = *iface.Address6
	}

	// if iface.HasNetmask6() {
	// 	nic.Netmask = *iface.Netmask6
	// }

	if iface.HasGateway6() {
		nic.Gateway = *iface.Gateway6
	}

	if nic.Address == "" && nic.Netmask == "" && nic.Gateway == "" {
		return nil
	}
	return &nic
}
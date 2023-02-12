package bridges

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type bridgesDataSourceModel struct {
	Bridges []bridgeModel         `tfsdk:"network_bridges"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type bridgeModel struct {
	ID          types.String            `tfsdk:"id"`
	Node        types.String            `tfsdk:"node"`
	Name        types.String            `tfsdk:"name"`
	Active      types.Bool              `tfsdk:"active"`
	Autostart   types.Bool              `tfsdk:"autostart"`
	VLANAware   types.Bool              `tfsdk:"vlan_aware"`
	Interfaces  []types.String          `tfsdk:"interfaces"`
	IPv4        *network.IpAddressModel `tfsdk:"ipv4"`
	IPv6        *network.IpAddressModel `tfsdk:"ipv6"`
	IPv4Gateway types.String            `tfsdk:"ipv4_gateway"`
	IPv6Gateway types.String            `tfsdk:"ipv6_gateway"`
	Comments    types.String            `tfsdk:"comments"`
}

func BridgeToModel(bridge *service.NetworkBridge) bridgeModel {
	b := bridgeModel{
		ID:        types.StringValue(network.FormId(bridge.Node, bridge.Name)),
		Node:      types.StringValue(bridge.Node),
		Name:      types.StringValue(bridge.Name),
		Active:    types.BoolValue(bridge.Active),
		Autostart: types.BoolValue(bridge.Autostart),
		VLANAware: types.BoolValue(bridge.VLANAware),
	}

	if bridge.Comments != nil {
		b.Comments = types.StringValue(*bridge.Comments)
	}

	if bridge.IPv4 != nil {
		b.IPv4 = &network.IpAddressModel{
			Address: types.StringValue(bridge.IPv4.Address),
			Netmask: types.StringValue(bridge.IPv4.Netmask),
		}
	}

	if bridge.IPv6 != nil {
		b.IPv6 = &network.IpAddressModel{
			Address: types.StringValue(bridge.IPv6.Address),
			Netmask: types.StringValue(bridge.IPv6.Netmask),
		}
	}

	if bridge.IPv4Gateway != nil {
		b.IPv4Gateway = types.StringValue(*bridge.IPv4Gateway)
	}

	if bridge.IPv6Gateway != nil {
		b.IPv6Gateway = types.StringValue(*bridge.IPv6Gateway)
	}

	for _, iface := range bridge.Interfaces {
		b.Interfaces = append(b.Interfaces, types.StringValue(iface))
	}

	return b
}

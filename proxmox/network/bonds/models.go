package bonds

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type bondsDataSourceModel struct {
	Bonds   []bondModel           `tfsdk:"network_bonds"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type bondModel struct {
	ID          types.String            `tfsdk:"id"`
	Node        types.String            `tfsdk:"node"`
	Name        types.String            `tfsdk:"name"`
	Active      types.Bool              `tfsdk:"active"`
	Autostart   types.Bool              `tfsdk:"autostart"`
	HashPolicy  types.String            `tfsdk:"hash_policy"`
	BondPrimary types.String            `tfsdk:"bond_primary"`
	Mode        types.String            `tfsdk:"mode"`
	MiiMon      types.String            `tfsdk:"mii_mon"`
	Comments    types.String            `tfsdk:"comments"`
	Interfaces  []types.String          `tfsdk:"interfaces"`
	IPv4        *network.IpAddressModel `tfsdk:"ipv4"`
	IPv6        *network.IpAddressModel `tfsdk:"ipv6"`
	IPv4Gateway types.String            `tfsdk:"ipv4_gateway"`
	IPv6Gateway types.String            `tfsdk:"ipv6_gateway"`
}

func BondToModel(bond *service.NetworkBond) bondModel {
	b := bondModel{
		ID:        types.StringValue(network.FormId(bond.Node, bond.Name)),
		Node:      types.StringValue(bond.Node),
		Name:      types.StringValue(bond.Name),
		Active:    types.BoolValue(bond.Active),
		Autostart: types.BoolValue(bond.Autostart),
		Mode:      types.StringValue(string(bond.Mode)),
	}

	if bond.BondPrimary != nil {
		b.BondPrimary = types.StringValue(*bond.BondPrimary)
	}

	if bond.Comments != nil {
		b.Comments = types.StringValue(*bond.Comments)
	}

	if bond.HashPolicy != nil {
		b.HashPolicy = types.StringValue(string(*bond.HashPolicy))
	}

	if bond.MiiMon != nil {
		b.MiiMon = types.StringValue(*bond.MiiMon)
	}

	if bond.IPv4 != nil {
		b.IPv4 = &network.IpAddressModel{
			Address: types.StringValue(bond.IPv4.Address),
			Netmask: types.StringValue(bond.IPv4.Netmask),
		}
	}

	if bond.IPv6 != nil {
		b.IPv6 = &network.IpAddressModel{
			Address: types.StringValue(bond.IPv6.Address),
			Netmask: types.StringValue(bond.IPv6.Netmask),
		}
	}

	if bond.IPv4Gateway != nil {
		b.IPv4Gateway = types.StringValue(*bond.IPv4Gateway)
	}

	if bond.IPv6Gateway != nil {
		b.IPv6Gateway = types.StringValue(*bond.IPv6Gateway)
	}

	for _, iface := range bond.Interfaces {
		b.Interfaces = append(b.Interfaces, types.StringValue(iface))
	}

	return b
}

package bonds

import (
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type bondsDataSourceModel struct {
	Bonds   []bondModel           `tfsdk:"network_bonds"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type bondModel struct {
	ID          types.String   `tfsdk:"id"`
	Node        types.String   `tfsdk:"node"`
	Name        types.String   `tfsdk:"name"`
	Active      types.Bool     `tfsdk:"active"`
	Autostart   types.Bool     `tfsdk:"autostart"`
	HashPolicy  types.String   `tfsdk:"hash_policy"`
	BondPrimary types.String   `tfsdk:"bond_primary"`
	Mode        types.String   `tfsdk:"mode"`
	MiiMon      types.String   `tfsdk:"mii_mon"`
	Comments    types.String   `tfsdk:"comments"`
	Interfaces  []types.String `tfsdk:"interfaces"`
}

func BondToModel(bond *service.NetworkBond) bondModel {
	b := bondModel{
		ID:        types.StringValue(fmt.Sprintf("%s/%s", bond.Node, bond.Name)),
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

	for _, iface := range bond.Interfaces {
		b.Interfaces = append(b.Interfaces, types.StringValue(iface))
	}

	return b
}

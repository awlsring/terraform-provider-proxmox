package network

import "github.com/hashicorp/terraform-plugin-framework/types"

type IpAddressModel struct {
	Address types.String `tfsdk:"address"`
	Netmask types.String `tfsdk:"netmask"`
}

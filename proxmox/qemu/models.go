package qemu

import "github.com/hashicorp/terraform-plugin-framework/types"

type VirtualMachineModel struct {
	ID         types.Number            `tfsdk:"id"`
	Node       types.String            `tfsdk:"node"`
	Name       types.String            `tfsdk:"name"`
	Cores      types.Number            `tfsdk:"cores"`
	Memory     types.Int64             `tfsdk:"memory"`
	Agent      types.Bool              `tfsdk:"agent"`
	Tags       []types.String          `tfsdk:"tags"`
	Disks      []VirtualDiskModel      `tfsdk:"disks"`
	Interfaces []VirtualInterfaceModel `tfsdk:"network_interfaces"`
}

type VirtualDiskModel struct {
	Storage  types.String `tfsdk:"storage"`
	Size     types.Int64  `tfsdk:"size"`
	Type     types.String `tfsdk:"type"`
	Position types.String `tfsdk:"position"`
	Discard  types.Bool   `tfsdk:"discard"`
}

type VirtualInterfaceModel struct {
	Bridge     types.String `tfsdk:"bridge"`
	Vlan       types.Number `tfsdk:"vlan"`
	Model      types.String `tfsdk:"model"`
	MacAddress types.String `tfsdk:"mac_address"`
	Position   types.String `tfsdk:"position"`
	Firewall   types.Bool   `tfsdk:"firewall"`
}

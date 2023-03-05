package types

import "github.com/hashicorp/terraform-plugin-framework/types"

type VirtualMachineAgentModel struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	UseFSTrim types.Bool   `tfsdk:"use_fstrim"`
	Type      types.String `tfsdk:"type"`
}

type VirtualMachineCpuModel struct {
	Architecture types.String `tfsdk:"architecture"`
	Cores        types.Int64  `tfsdk:"cores"`
	Sockets      types.Int64  `tfsdk:"sockets"`
	EmulatedType types.String `tfsdk:"emulated_type"`
	CPUUnits     types.Int64  `tfsdk:"cpu_units"`
}

type VirtualMachineMemoryModel struct {
	Dedicated types.Int64 `tfsdk:"dedicated"`
	Floating  types.Int64 `tfsdk:"floating"`
	Shared    types.Int64 `tfsdk:"shared"`
}

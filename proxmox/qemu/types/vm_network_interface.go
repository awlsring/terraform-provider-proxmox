package types

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var VirtualMachineNetworkInterface = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"bridge":       types.StringType,
		"enabled":      types.BoolType,
		"use_firewall": types.BoolType,
		"mac_address":  types.StringType,
		"model":        types.StringType,
		"rate_limit":   types.Int64Type,
		"position":     types.Int64Type,
		"vlan":         types.Int64Type,
		"mtu":          types.Int64Type,
	},
}

func NewVirtualMachineNetworkInterfaceListType() VirtualMachineNetworkInterfaceListType {
	return VirtualMachineNetworkInterfaceListType{
		types.ListType{
			ElemType: VirtualMachineNetworkInterface,
		},
	}
}

type VirtualMachineNetworkInterfaceListType struct {
	types.ListType
}

func (c VirtualMachineNetworkInterfaceListType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := c.ListType.ValueFromTerraform(ctx, in)

	list := val.(types.List)

	nics := []VirtualMachineNetworkInterfaceModel{}
	for _, nic := range list.Elements() {
		var v VirtualMachineNetworkInterfaceModel
		t, err := nic.ToTerraformValue(ctx)
		if err != nil {
			return nil, fmt.Errorf("error converting disk to terraform value: %w", err)
		}
		t.As(&v)
		nics = append(nics, v)
	}

	return VirtualMachineNetworkInterfaceListValue{
		val.(types.List),
		nics,
	}, err
}

type VirtualMachineNetworkInterfaceListValue struct {
	types.List
	Nics []VirtualMachineNetworkInterfaceModel
}

func VirtualMachineNetworkInterfaceListValueFrom(ctx context.Context, nics []VirtualMachineNetworkInterfaceModel) VirtualMachineNetworkInterfaceListValue {
	l, diags := types.ListValueFrom(ctx, VirtualMachineDisk, nics)
	if diags.HasError() {
		tflog.Debug(ctx, fmt.Sprintf("diags: %v", diags))
	}

	return VirtualMachineNetworkInterfaceListValue{
		l,
		nics,
	}
}

type VirtualMachineNetworkInterfaceModel struct {
	Position    types.Int64  `tfsdk:"position"`
	Bridge      types.String `tfsdk:"bridge"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	UseFirewall types.Bool   `tfsdk:"use_firewall"`
	MacAddress  types.String `tfsdk:"mac_address"`
	Model       types.String `tfsdk:"model"`
	Vlan        types.Int64  `tfsdk:"vlan"`
	RateLimit   types.Int64  `tfsdk:"rate_limit"`
	MTU         types.Int64  `tfsdk:"mtu"`
}

func VirtualMachineNetworkInterfaceToListValue(ctx context.Context, nics []vm.VirtualMachineNetworkInterface) VirtualMachineNetworkInterfaceListValue {
	models := []VirtualMachineNetworkInterfaceModel{}
	for _, nic := range nics {
		m := VirtualMachineNetworkInterfaceModel{
			Position:    types.Int64Value(int64(nic.Position)),
			Bridge:      types.StringValue(nic.Bridge),
			Enabled:     types.BoolValue(nic.Enabled),
			UseFirewall: types.BoolValue(nic.Firewall),
			MacAddress:  types.StringValue(nic.MAC),
			Model:       types.StringValue(string(nic.Model)),
		}
		if nic.RateLimit != nil {
			m.RateLimit = types.Int64Value(int64(*nic.RateLimit))
		}
		if nic.MTU != nil {
			m.MTU = types.Int64Value(int64(*nic.MTU))
		}
		if nic.VLAN != nil {
			m.Vlan = types.Int64Value(int64(*nic.VLAN))
		}
		models = append(models, m)
	}
	return VirtualMachineNetworkInterfaceListValueFrom(ctx, models)
}

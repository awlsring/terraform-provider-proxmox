package types

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

func NewVirtualMachineNetworkInterfaceSetType() VirtualMachineNetworkInterfaceSetType {
	return VirtualMachineNetworkInterfaceSetType{
		types.SetType{
			ElemType: VirtualMachineNetworkInterface,
		},
	}
}

type VirtualMachineNetworkInterfaceSetType struct {
	types.SetType
}

func (c VirtualMachineNetworkInterfaceSetType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := c.SetType.ValueFromTerraform(ctx, in)

	set := val.(types.Set)

	nics := []VirtualMachineNetworkInterfaceModel{}
	for _, nic := range set.Elements() {
		var v VirtualMachineNetworkInterfaceModel
		t := nic.(types.Object)
		if err != nil {
			return nil, fmt.Errorf("error converting disk to terraform value: %w", err)
		}
		t.As(ctx, &v, basetypes.ObjectAsOptions{})
		nics = append(nics, v)
	}

	return VirtualMachineNetworkInterfaceSetValue{
		val.(types.Set),
		nics,
	}, err
}

func (d VirtualMachineNetworkInterfaceSetType) Equal(o attr.Type) bool {
	if d.ElemType == nil {
		return false
	}

	other, ok := o.(VirtualMachineNetworkInterfaceSetType)
	if !ok {
		other, ok := o.(types.SetType)
		if !ok {
			return false
		}
		return d.ElemType.Equal(other.ElemType)
	}
	return d.ElemType.Equal(other.ElemType)
}

type VirtualMachineNetworkInterfaceSetValue struct {
	types.Set
	Nics []VirtualMachineNetworkInterfaceModel
}

func VirtualMachineNetworkInterfaceSetValueFrom(ctx context.Context, nics []VirtualMachineNetworkInterfaceModel) VirtualMachineNetworkInterfaceSetValue {
	l, diags := types.SetValueFrom(ctx, VirtualMachineNetworkInterface, nics)
	if diags.HasError() {
		tflog.Debug(ctx, fmt.Sprintf("diags: %v", diags))
	}

	if len(nics) == 0 {
		l = types.SetNull(VirtualMachineNetworkInterface)
	}

	return VirtualMachineNetworkInterfaceSetValue{
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

func VirtualMachineNetworkInterfaceToSetValue(ctx context.Context, nics []vm.VirtualMachineNetworkInterface) VirtualMachineNetworkInterfaceSetValue {
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
	return VirtualMachineNetworkInterfaceSetValueFrom(ctx, models)
}

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var VirtualMachineDataSourceType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":          types.Int64Type,
		"node":        types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"tags": types.SetType{
			ElemType: types.StringType,
		},
		"agent": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled":    types.BoolType,
				"use_fstrim": types.BoolType,
				"type":       types.StringType,
			},
		},
		"bios": types.StringType,
		"cpu": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"architecture":  types.StringType,
				"cores":         types.Int64Type,
				"sockets":       types.Int64Type,
				"emulated_type": types.StringType,
				"cpu_units":     types.Int64Type,
			},
		},
		"disks":              NewVirtualMachineDiskSetType(),
		"network_interfaces": NewVirtualMachineNetworkInterfaceSetType(),
		"pci_devices":        NewVirtualMachinePCIDeviceSetType(),
		"memory": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"dedicated": types.Int64Type,
				"floating":  types.Int64Type,
				"shared":    types.Int64Type,
			},
		},
		"cloud_init": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"user": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"password": types.StringType,
						"public_keys": types.SetType{
							ElemType: types.StringType,
						},
						"name": types.StringType,
					},
				},
				"ip": NewCloudInitIpSetType(),
				"dns": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nameserver": types.StringType,
						"domain":     types.StringType,
					},
				},
			},
		},
		"machine_type":       types.StringType,
		"kvm_arguments":      types.StringType,
		"keyboard_layout":    types.StringType,
		"type":               types.StringType,
		"resource_pool":      types.StringType,
		"start_on_node_boot": types.BoolType,
	},
}

func NewVirtualMachineDataSourceType() VirtualMachineDataSourceSetType {
	return VirtualMachineDataSourceSetType{
		types.SetType{
			ElemType: VirtualMachineDataSourceType,
		},
	}
}

type VirtualMachineDataSourceSetType struct {
	types.SetType
}

func (c VirtualMachineDataSourceSetType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := c.SetType.ValueFromTerraform(ctx, in)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("error converting disk to terraform value: %v", err))
		return nil, err
	}

	set := val.(types.Set)

	tflog.Debug(ctx, "creating vm set")
	vms := []VirtualMachineDataSourceModel{}
	for _, vm := range set.Elements() {
		var v VirtualMachineDataSourceModel
		t := vm.(types.Object)
		if err != nil {
			return nil, fmt.Errorf("error converting vm to terraform value: %w", err)
		}
		t.As(ctx, &v, basetypes.ObjectAsOptions{})
		vms = append(vms, v)
	}
	tflog.Debug(ctx, fmt.Sprintf("vms: %v", vms))

	return VirtualMachineDataSourceSetValue{
		val.(types.Set),
		vms,
	}, err
}

type VirtualMachineDataSourceSetValue struct {
	types.Set
	Vms []VirtualMachineDataSourceModel
}

func VirtualMachineDataSourceSetValueFrom(ctx context.Context, vms []VirtualMachineDataSourceModel) (VirtualMachineDataSourceSetValue, error) {
	l, diags := types.SetValueFrom(ctx, VirtualMachineDataSourceType, vms)
	if diags.HasError() {
		return VirtualMachineDataSourceSetValue{}, fmt.Errorf("Error converting Models to VirtualMachineDataSourceSetValue: %v", diags)
	}

	if len(vms) == 0 {
		tflog.Debug(ctx, "no vms found, returning null set")
		l = types.SetNull(VirtualMachineDataSourceType)
	}

	return VirtualMachineDataSourceSetValue{
		l,
		vms,
	}, nil
}

func (ci VirtualMachineDataSourceSetType) Equal(o attr.Type) bool {
	if ci.ElemType == nil {
		return false
	}

	other, ok := o.(VirtualMachineDataSourceSetType)
	if !ok {
		other, ok := o.(types.SetType)
		if !ok {
			return false
		}
		return ci.ElemType.Equal(other.ElemType)
	}
	return ci.ElemType.Equal(other.ElemType)
}

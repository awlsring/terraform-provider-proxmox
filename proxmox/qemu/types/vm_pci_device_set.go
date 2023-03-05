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

var VirtualMachinePCIDevice = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":        types.StringType,
		"id":          types.StringType,
		"pcie":        types.BoolType,
		"mdev":        types.StringType,
		"rombar":      types.BoolType,
		"rom_file":    types.StringType,
		"primary_gpu": types.BoolType,
	},
}

func NewVirtualMachinePCIDeviceSetType() VirtualMachinePCIDeviceSetType {
	return VirtualMachinePCIDeviceSetType{
		types.SetType{
			ElemType: VirtualMachinePCIDevice,
		},
	}
}

type VirtualMachinePCIDeviceSetType struct {
	types.SetType
}

func (c VirtualMachinePCIDeviceSetType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := c.SetType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	set := val.(types.Set)

	disks := []VirtualMachinePCIDeviceModel{}
	for _, disk := range set.Elements() {
		var v VirtualMachinePCIDeviceModel
		t := disk.(types.Object)
		if err != nil {
			return nil, fmt.Errorf("error converting disk to terraform value: %w", err)
		}
		t.As(ctx, &v, basetypes.ObjectAsOptions{})
		disks = append(disks, v)
	}

	return VirtualMachinePCIDeviceSetValue{
		val.(types.Set),
		disks,
	}, err
}

func (d VirtualMachinePCIDeviceSetType) Equal(o attr.Type) bool {
	if d.ElemType == nil {
		return false
	}

	other, ok := o.(VirtualMachinePCIDeviceSetType)
	if !ok {
		other, ok := o.(types.SetType)
		if !ok {
			return false
		}
		return d.ElemType.Equal(other.ElemType)
	}
	return d.ElemType.Equal(other.ElemType)
}

type VirtualMachinePCIDeviceSetValue struct {
	types.Set
	PCIDevices []VirtualMachinePCIDeviceModel
}

func VirtualMachinePCIDeviceSetValueFrom(ctx context.Context, dev []VirtualMachinePCIDeviceModel) VirtualMachinePCIDeviceSetValue {
	l, diags := types.SetValueFrom(ctx, VirtualMachinePCIDevice, dev)
	if diags.HasError() {
		tflog.Debug(ctx, fmt.Sprintf("diags: %v", diags))
	}

	if len(dev) == 0 {
		l = types.SetNull(VirtualMachinePCIDevice)
	}

	return VirtualMachinePCIDeviceSetValue{
		l,
		dev,
	}
}

type VirtualMachinePCIDeviceModel struct {
	Name       types.String `tfsdk:"name"`
	ID         types.String `tfsdk:"id"`
	PCIE       types.Bool   `tfsdk:"pcie"`
	MDEV       types.String `tfsdk:"mdev"`
	ROMBAR     types.Bool   `tfsdk:"rombar"`
	ROMFile    types.String `tfsdk:"rom_file"`
	PrimaryGPU types.Bool   `tfsdk:"primary_gpu"`
}

func VirtualMachinePCIDeviceToSetValue(ctx context.Context, devices []vm.VirtualMachinePCIDevice) VirtualMachinePCIDeviceSetValue {
	models := []VirtualMachinePCIDeviceModel{}
	for _, dev := range devices {
		m := VirtualMachinePCIDeviceModel{
			Name:       types.StringValue(dev.Name),
			ID:         types.StringValue(dev.ID),
			PCIE:       types.BoolValue(dev.PCIE),
			ROMBAR:     types.BoolValue(dev.ROMBAR),
			PrimaryGPU: types.BoolValue(dev.PrimaryGPU),
		}
		if dev.MDEV != nil {
			m.MDEV = types.StringValue(*dev.MDEV)
		}

		if dev.ROMFile != nil {
			m.ROMFile = types.StringValue(*dev.ROMFile)
		}
		models = append(models, m)
	}
	return VirtualMachinePCIDeviceSetValueFrom(ctx, models)
}

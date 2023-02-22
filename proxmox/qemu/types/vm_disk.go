package types

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var VirtualMachineDisk = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"storage":      types.StringType,
		"file_format":  types.StringType,
		"size":         types.Int64Type,
		"use_iothread": types.BoolType,
		"speed_limits": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"read":            types.Int64Type,
				"write":           types.Int64Type,
				"read_burstable":  types.Int64Type,
				"write_burstable": types.Int64Type,
			},
		},
		"interface_type": types.StringType,
		"position":       types.Int64Type,
		"ssd_emulation":  types.BoolType,
		"discard":        types.BoolType,
	},
}

func NewVirtualMachineDiskSetType() VirtualMachineDiskSetType {
	return VirtualMachineDiskSetType{
		types.SetType{
			ElemType: VirtualMachineDisk,
		},
	}
}

type VirtualMachineDiskSetType struct {
	types.SetType
}

func (c VirtualMachineDiskSetType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := c.SetType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	set := val.(types.Set)

	disks := []VirtualMachineDiskModel{}
	for _, disk := range set.Elements() {
		var v VirtualMachineDiskModel
		t := disk.(types.Object)
		if err != nil {
			return nil, fmt.Errorf("error converting disk to terraform value: %w", err)
		}
		t.As(ctx, &v, basetypes.ObjectAsOptions{})
		disks = append(disks, v)
	}

	return VirtualMachineDiskSetValue{
		val.(types.Set),
		disks,
	}, err
}

type VirtualMachineDiskSetValue struct {
	types.Set
	Disks []VirtualMachineDiskModel
}

func VirtualMachineDiskSetValueFrom(ctx context.Context, disks []VirtualMachineDiskModel) VirtualMachineDiskSetValue {
	l, diags := types.SetValueFrom(ctx, VirtualMachineDisk, disks)
	if diags.HasError() {
		tflog.Debug(ctx, fmt.Sprintf("diags: %v", diags))
	}

	if len(disks) == 0 {
		l = types.SetNull(VirtualMachineDisk)
	}

	return VirtualMachineDiskSetValue{
		l,
		disks,
	}
}

type VirtualMachineDiskModel struct {
	Storage       types.String                        `tfsdk:"storage"`
	FileFormat    types.String                        `tfsdk:"file_format"`
	Size          types.Int64                         `tfsdk:"size"`
	UseIOThread   types.Bool                          `tfsdk:"use_iothread"`
	SpeedLimits   *VirtualMachineDiskSpeedLimitsModel `tfsdk:"speed_limits"`
	InterfaceType types.String                        `tfsdk:"interface_type"`
	SSDEmulation  types.Bool                          `tfsdk:"ssd_emulation"`
	Position      types.Int64                         `tfsdk:"position"`
	Discard       types.Bool                          `tfsdk:"discard"`
}

type VirtualMachineDiskSpeedLimitsModel struct {
	Read           types.Int64 `tfsdk:"read"`
	ReadBurstable  types.Int64 `tfsdk:"read_burstable"`
	Write          types.Int64 `tfsdk:"write"`
	WriteBurstable types.Int64 `tfsdk:"write_burstable"`
}

func VirtualMachineDiskToSetValue(ctx context.Context, disks []vm.VirtualMachineDisk) VirtualMachineDiskSetValue {
	models := []VirtualMachineDiskModel{}
	for _, disk := range disks {
		size := utils.BytesToGb(disk.Size)
		m := VirtualMachineDiskModel{
			Storage:       types.StringValue(disk.Storage),
			Size:          types.Int64Value(size),
			UseIOThread:   types.BoolValue(disk.UseIOThreads),
			InterfaceType: types.StringValue(string(disk.InterfaceType)),
			SSDEmulation:  types.BoolValue(disk.SSDEmulation),
			Position:      types.Int64Value(int64(disk.Position)),
			Discard:       types.BoolValue(disk.Discard),
		}
		if disk.FileFormat != nil {
			m.FileFormat = types.StringValue(string(*disk.FileFormat))
		}
		if disk.SpeedLimits != nil {
			if disk.SpeedLimits.Read != nil {
				m.SpeedLimits.Read = types.Int64Value(int64(*disk.SpeedLimits.Read))
			}
			if disk.SpeedLimits.ReadBurstable != nil {
				m.SpeedLimits.ReadBurstable = types.Int64Value(int64(*disk.SpeedLimits.ReadBurstable))
			}
			if disk.SpeedLimits.Write != nil {
				m.SpeedLimits.Write = types.Int64Value(int64(*disk.SpeedLimits.Write))
			}
			if disk.SpeedLimits.WriteBurstable != nil {
				m.SpeedLimits.WriteBurstable = types.Int64Value(int64(*disk.SpeedLimits.WriteBurstable))
			}
		}
		models = append(models, m)
	}
	return VirtualMachineDiskSetValueFrom(ctx, models)
}

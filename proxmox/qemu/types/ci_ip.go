package types

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ attr.TypeWithElementType = CloudInitIpSetType{}
)

var CloudInitIpSchema = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"position": schema.Int64Attribute{
			Required:    true,
			Description: "The position of the network interface in the VM as an int. Used to determine the interface name (net0, net1, etc).",
		},
		"v4": schema.SingleNestedAttribute{
			Optional:   true,
			Attributes: CloudInitIpConfigSchema,
		},
		"v6": schema.SingleNestedAttribute{
			Optional:   true,
			Attributes: CloudInitIpConfigSchema,
		},
	},
}

var CloudInitIpType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"position": types.Int64Type,
		"v4": types.ObjectType{
			AttrTypes: CloudInitIpConfigTypes,
		},
		"v6": types.ObjectType{
			AttrTypes: CloudInitIpConfigTypes,
		},
	},
}

var CloudInitIpConfigTypes = map[string]attr.Type{
	"dhcp":    types.BoolType,
	"address": types.StringType,
	"netmask": types.StringType,
	"gateway": types.StringType,
}

var CloudInitIpConfigSchema = map[string]schema.Attribute{
	"dhcp": schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether to use DHCP to get the IP address.",
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"address": schema.StringAttribute{
		Optional:    true,
		Description: "The IP address to use for the machine.",
	},
	"netmask": schema.StringAttribute{
		Optional:    true,
		Description: "The IP address netmask to use for the machine.",
	},
	"gateway": schema.StringAttribute{
		Optional:    true,
		Description: "The gateway to use for the machine.",
	},
}

type VirtualMachineCloudInitModel struct {
	User *VirtualMachineCloudInitUserModel `tfsdk:"user"`
	IP   CloudInitIpSetValue               `tfsdk:"ip"`
	DNS  *VirtualMachineCloudInitDnsModel  `tfsdk:"dns"`
}

type VirtualMachineCloudInitUserModel struct {
	Name       types.String `tfsdk:"name"`
	Password   types.String `tfsdk:"password"`
	PublicKeys types.List   `tfsdk:"public_keys"`
}

type VirtualMachineCloudInitIpModel struct {
	Positition types.Int64                           `tfsdk:"position"`
	V4         *VirtualMachineCloudInitIpConfigModel `tfsdk:"v4"`
	V6         *VirtualMachineCloudInitIpConfigModel `tfsdk:"v6"`
}

type VirtualMachineCloudInitIpConfigModel struct {
	DHCP    types.Bool   `tfsdk:"dhcp"`
	Address types.String `tfsdk:"address"`
	Netmask types.String `tfsdk:"netmask"`
	Gateway types.String `tfsdk:"gateway"`
}

type VirtualMachineCloudInitDnsModel struct {
	Nameserver types.String `tfsdk:"nameserver"`
	Domain     types.String `tfsdk:"domain"`
}

var CloudInitIpConfig = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"dhcp":    types.BoolType,
		"address": types.StringType,
		"netmask": types.StringType,
		"gateway": types.StringType,
	},
}

var CloudInitIp = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"position": types.Int64Type,
		"v4":       CloudInitIpConfig,
		"v6":       CloudInitIpConfig,
	},
}

func NewCloudInitIpSetType() CloudInitIpSetType {
	return CloudInitIpSetType{
		types.SetType{
			ElemType: CloudInitIp,
		},
	}
}

type CloudInitIpSetType struct {
	types.SetType
}

func (ci CloudInitIpSetType) Equal(o attr.Type) bool {
	if ci.ElemType == nil {
		return false
	}
	// pass if is a base Set or a CloudInitIpSet
	other, ok := o.(CloudInitIpSetType)
	if !ok {
		other, ok := o.(types.SetType)
		if !ok {
			return false
		}
		return ci.ElemType.Equal(other.ElemType)
	}
	return ci.ElemType.Equal(other.ElemType)
}

func (st CloudInitIpSetType) String() string {
	return "types.CloudInitIpSetType[" + st.ElemType.String() + "]"
}

func (c CloudInitIpSetType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := c.SetType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	set := val.(types.Set)

	configs := []VirtualMachineCloudInitIpModel{}
	for _, ip := range set.Elements() {
		var v VirtualMachineCloudInitIpModel
		t := ip.(types.Object)
		if err != nil {
			return nil, fmt.Errorf("error converting config to terraform value: %w", err)
		}
		t.As(ctx, &v, basetypes.ObjectAsOptions{})
		configs = append(configs, v)
	}

	return CloudInitIpSetValue{
		val.(types.Set),
		configs,
	}, err
}

type CloudInitIpSetValue struct {
	types.Set
	Configs []VirtualMachineCloudInitIpModel
}

func CloudInitIpSetValueFrom(ctx context.Context, configs []VirtualMachineCloudInitIpModel) CloudInitIpSetValue {
	l, diags := types.SetValueFrom(ctx, CloudInitIpType, configs)
	if diags.HasError() {
		tflog.Debug(ctx, fmt.Sprintf("diags: %v", diags))
	}

	if len(configs) == 0 {
		l = types.SetNull(CloudInitIpType)
	}

	return CloudInitIpSetValue{
		l,
		configs,
	}
}

func CloudInitIpToSetValue(ctx context.Context, ip []vm.VirtualMachineCloudInitIp) CloudInitIpSetValue {
	models := []VirtualMachineCloudInitIpModel{}
	for _, i := range ip {
		m := VirtualMachineCloudInitIpModel{
			Positition: types.Int64Value(int64(i.Position)),
			V4:         translateIpConfig(i.V4),
			V6:         translateIpConfig(i.V6),
		}
		models = append(models, m)
	}
	return CloudInitIpSetValueFrom(ctx, models)
}

func CloudInitToModel(ctx context.Context, ci *vm.VirtualMachineCloudInit) *VirtualMachineCloudInitModel {
	tflog.Debug(ctx, fmt.Sprintf("Passed cloudinit: %v", ci))
	if ci == nil {
		return nil
	}

	m := VirtualMachineCloudInitModel{
		IP: CloudInitIpToSetValue(ctx, ci.Ip),
	}

	if ci.User != nil {
		user := VirtualMachineCloudInitUserModel{
			Name:     utils.StringToTfType(ci.User.Name),
			Password: utils.StringToTfType(ci.User.Password),
		}
		user.PublicKeys = utils.UnpackListType(ci.User.PublicKeys)
		tflog.Debug(ctx, fmt.Sprintf("Converted cloudinit user: %v", user))
		m.User = &user
	}

	if ci.Dns != nil {
		dns := VirtualMachineCloudInitDnsModel{
			Domain:     utils.StringToTfType(ci.Dns.Domain),
			Nameserver: utils.StringToTfType(ci.Dns.Nameserver),
		}
		tflog.Debug(ctx, fmt.Sprintf("Converted cloudinit dns: %v", dns))
		m.DNS = &dns
	}

	return &m
}

func translateIpConfig(c *vm.VirtualMachineCloudInitIpConfig) *VirtualMachineCloudInitIpConfigModel {
	if c == nil {
		return nil
	}
	m := &VirtualMachineCloudInitIpConfigModel{
		DHCP:    types.BoolValue(c.DHCP),
		Address: utils.StringToTfType(c.Address),
		Netmask: utils.StringToTfType(c.Netmask),
		Gateway: utils.StringToTfType(c.Gateway),
	}
	return m
}

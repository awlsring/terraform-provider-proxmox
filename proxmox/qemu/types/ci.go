package types

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var CloudInitDataSourceAttributes = map[string]dschema.Attribute{
	"user": dschema.SingleNestedAttribute{
		Computed:   true,
		Attributes: CloudInitUserDataSourceAttributes,
	},
	"ip": dschema.SetNestedAttribute{
		Computed:     true,
		CustomType:   NewCloudInitIpSetType(),
		NestedObject: CloudInitIpDataSourceSchema,
	},
	"dns": dschema.SingleNestedAttribute{
		Computed:   true,
		Attributes: CloudInitDnsDataSourceAttributes,
	},
}

var CloudInitUserDataSourceAttributes = map[string]dschema.Attribute{
	"name": dschema.StringAttribute{
		Computed:    true,
		Description: "The name of the user.",
	},
	"password": dschema.StringAttribute{
		Computed:    true,
		Description: "The password of the user.",
	},
	"public_keys": dschema.SetAttribute{
		Computed:    true,
		Description: "The public ssh keys of the user.",
		ElementType: types.StringType,
	},
}

var CloudInitIpDataSourceSchema = dschema.NestedAttributeObject{
	Attributes: map[string]dschema.Attribute{
		"position": dschema.Int64Attribute{
			Computed:    true,
			Description: "The position of the network interface in the VM as an int. Used to determine the interface name (net0, net1, etc).",
		},
		"v4": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: CloudInitIpConfigDataSourceSchema,
		},
		"v6": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: CloudInitIpConfigDataSourceSchema,
		},
	},
}

var CloudInitIpConfigDataSourceSchema = map[string]dschema.Attribute{
	"dhcp": dschema.BoolAttribute{
		Computed:    true,
		Description: "Whether to use DHCP to get the IP address.",
	},
	"address": dschema.StringAttribute{
		Computed:    true,
		Description: "The IP address to use for the machine.",
	},
	"netmask": dschema.StringAttribute{
		Computed:    true,
		Description: "The IP address netmask to use for the machine.",
	},
	"gateway": dschema.StringAttribute{
		Computed:    true,
		Description: "The gateway to use for the machine.",
	},
}

var CloudInitDnsDataSourceAttributes = map[string]dschema.Attribute{
	"nameserver": dschema.StringAttribute{
		Computed:    true,
		Description: "The nameserver to use for the machine.",
	},
	"domain": dschema.StringAttribute{
		Computed:    true,
		Description: "The domain to use for the machine.",
	},
}

var CloudInitAttributes = map[string]schema.Attribute{
	"user": schema.SingleNestedAttribute{
		Optional:   true,
		Attributes: CloudInitUserAttributes,
	},
	"ip": schema.SetNestedAttribute{
		Optional:     true,
		CustomType:   NewCloudInitIpSetType(),
		NestedObject: CloudInitIpSchema,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
	},
	"dns": schema.SingleNestedAttribute{
		Optional:   true,
		Attributes: CloudInitDnsAttributes,
	},
}

var CloudInitUserAttributes = map[string]schema.Attribute{
	"name": schema.StringAttribute{
		Required:    true,
		Description: "The name of the user.",
	},
	"password": schema.StringAttribute{
		Optional:    true,
		Description: "The password of the user.",
	},
	"public_keys": schema.SetAttribute{
		Optional:    true,
		Description: "The public ssh keys of the user.",
		ElementType: types.StringType,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
	},
}

var CloudInitDnsAttributes = map[string]schema.Attribute{
	"nameserver": schema.StringAttribute{
		Optional:    true,
		Description: "The nameserver to use for the machine.",
	},
	"domain": schema.StringAttribute{
		Optional:    true,
		Description: "The domain to use for the machine.",
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
	PublicKeys types.Set    `tfsdk:"public_keys"`
}

type VirtualMachineCloudInitDnsModel struct {
	Nameserver types.String `tfsdk:"nameserver"`
	Domain     types.String `tfsdk:"domain"`
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
		user.PublicKeys = utils.UnpackSetType(ci.User.PublicKeys)
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

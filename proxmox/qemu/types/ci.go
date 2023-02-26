package types

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var CloudInitAttributes = map[string]schema.Attribute{
	"user": schema.SingleNestedAttribute{
		Optional:   true,
		Attributes: CloudInitUserAttributes,
	},
	"ip": schema.SetNestedAttribute{
		Optional:     true,
		CustomType:   NewCloudInitIpSetType(),
		NestedObject: CloudInitIpSchema,
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
	"public_keys": schema.ListAttribute{
		Optional:    true,
		Description: "The public ssh keys of the user.",
		ElementType: types.StringType,
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

package network

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	ds "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rs "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IpAddressType string

const (
	IP_ADDRESS_TYPE_4 IpAddressType = "ipv4"
	IP_ADDRESS_TYPE_6 IpAddressType = "ipv6"
)

var ipAttributes = map[string]attr.Type{
	"address": types.StringType,
	"netmask": types.StringType,
}

func IpResourceSchema(t IpAddressType) rs.ObjectAttribute {
	return rs.ObjectAttribute{
		Optional:       true,
		Description:    fmt.Sprintf("Information of the %v address.", t),
		AttributeTypes: ipAttributes,
		Validators: []validator.Object{
			objectvalidator.All(),
		},
	}
}

func IpDataSourceSchema(t IpAddressType) ds.ObjectAttribute {
	return ds.ObjectAttribute{
		Computed:       true,
		Description:    fmt.Sprintf("Information of the %v address.", t),
		AttributeTypes: ipAttributes,
	}
}

func IpGatewayResourceSchema(t IpAddressType) rs.StringAttribute {
	return rs.StringAttribute{
		Optional:    true,
		Description: fmt.Sprintf("The %v gateway.", t),
	}
}

func IpGatewayDataSourceSchema(t IpAddressType) ds.StringAttribute {
	return ds.StringAttribute{
		Optional:    true,
		Description: fmt.Sprintf("The %v gateway.", t),
	}
}

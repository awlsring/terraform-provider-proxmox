package filters

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FilterModel struct {
	Name   types.String   `tfsdk:"name"`
	Values []types.String `tfsdk:"values"`
}

type FilterConfig []string

func (f *FilterConfig) Schema() *schema.ListNestedAttribute {
	return &schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The name of the attribute to filter on.",
					Validators: []validator.String{
						stringvalidator.OneOf(*f...),
					},
				},
				"values": schema.ListAttribute{
					Required:    true,
					Description: "The value(s) to be used in the filter.",
					ElementType: types.StringType,
				},
			},
		},
	}
}

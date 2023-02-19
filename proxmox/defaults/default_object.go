package defaults

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Object = objectDefaultModifier{}

func DefaultObject(d map[string]attr.Value) planmodifier.Object {
	return objectDefaultModifier{
		Elements: d,
	}
}

func (m objectDefaultModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %v", m.Elements)
}

func (m objectDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%v`", m.Elements)
}

func (m objectDefaultModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue, resp.Diagnostics = types.ObjectValue(req.PlanValue.AttributeTypes(ctx), m.Elements)
}

type objectDefaultModifier struct {
	Elements map[string]attr.Value
}

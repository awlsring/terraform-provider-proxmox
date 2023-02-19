package defaults

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Map = mapDefaultModifier{}

func DefaultMap(elements map[string]attr.Value) mapDefaultModifier {
	return mapDefaultModifier{Elements: elements}
}

type mapDefaultModifier struct {
	Elements map[string]attr.Value
}

func (m mapDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.Elements)
}

func (m mapDefaultModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%s`", m.Elements)
}

func (m mapDefaultModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue, resp.Diagnostics = types.MapValue(req.PlanValue.ElementType(ctx), m.Elements)
}

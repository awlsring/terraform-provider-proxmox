package defaults

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.List = listDefaultModifier{}

func DefaultList(elements []attr.Value) listDefaultModifier {
	return listDefaultModifier{Elements: elements}
}

type listDefaultModifier struct {
	Elements []attr.Value
}

func (m listDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %v", m.Elements)
}

func (m listDefaultModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%v`", m.Elements)
}

func (m listDefaultModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue, resp.Diagnostics = types.ListValue(req.PlanValue.ElementType(ctx), m.Elements)
}

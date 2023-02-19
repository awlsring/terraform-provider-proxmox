package defaults

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Bool = boolDefaultModifier{}

func DefaultBool(d bool) planmodifier.Bool {
	return boolDefaultModifier{
		Default: d,
	}
}

func (m boolDefaultModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %v", m.Default)
}

func (m boolDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%v`", m.Default)
}

func (m boolDefaultModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = types.BoolValue(true)
}

type boolDefaultModifier struct {
	Computed bool
	Default  bool
}

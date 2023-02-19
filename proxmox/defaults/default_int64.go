package defaults

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Int64 = int64DefaultModifier{}

func DefaultInt64(d int64) planmodifier.Int64 {
	return int64DefaultModifier{
		Default: d,
	}
}

func (m int64DefaultModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %v", m.Default)
}

func (m int64DefaultModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%v`", m.Default)
}

func (m int64DefaultModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.ConfigValue.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue = types.Int64Value(m.Default)
}

type int64DefaultModifier struct {
	Default int64
}

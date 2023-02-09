package utils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func OptionalToPointerString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func PtrStringToString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func LoadPlanAndState(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, plan any, state any) (any, any, error) {
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return "", "", fmt.Errorf("error getting plan")
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return "", "", fmt.Errorf("error getting plan")
	}
	return plan, state, nil
}

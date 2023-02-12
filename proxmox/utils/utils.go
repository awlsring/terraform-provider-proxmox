package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func UnpackId(id string) (string, string, error) {
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("invalid id %s", id)
	}
	return s[0], s[1], nil
}

func FormId(node string, name string) string {
	return fmt.Sprintf("%s/%s", node, name)
}

func OptionalToPointerString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func OptionalToPointerBool(b bool) *bool {
	if !b {
		return nil
	}

	return &b
}

func PtrStringToString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Float32ToInt64(f float32) int64 {
	return int64(f)
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

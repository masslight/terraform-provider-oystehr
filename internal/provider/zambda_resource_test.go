package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestZambdaFunctionFieldsChanged_CodeOnlyChange_ReturnsFalse(t *testing.T) {
	state := Zambda{
		Name:           types.StringValue("orders-worker"),
		Runtime:        RuntimeValue{StringValue: basetypes.NewStringValue("nodejs20.x")},
		MemorySize:     types.Int32Value(1024),
		Timeout:        types.Int32Value(30),
		TriggerMethod:  types.StringValue("http_auth"),
		Schedule:       types.ObjectNull(scheduleAttrTypes()),
		SourceChecksum: types.StringValue("checksum-old"),
	}
	plan := state
	plan.SourceChecksum = types.StringValue("checksum-new")

	if zambdaFunctionFieldsChanged(plan, state) {
		t.Fatalf("expected no function field changes for code-only update, but got true")
	}
}

func TestZambdaFunctionFieldsChanged_FunctionChange_ReturnsTrue(t *testing.T) {
	state := Zambda{
		Name:          types.StringValue("orders-worker"),
		Runtime:       RuntimeValue{StringValue: basetypes.NewStringValue("nodejs20.x")},
		MemorySize:    types.Int32Value(1024),
		Timeout:       types.Int32Value(30),
		TriggerMethod: types.StringValue("http_auth"),
		Schedule:      types.ObjectNull(scheduleAttrTypes()),
	}
	plan := state
	plan.Timeout = types.Int32Value(60)

	if !zambdaFunctionFieldsChanged(plan, state) {
		t.Fatalf("expected function field changes when timeout changed, but got false")
	}
}

func scheduleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"expression": types.StringType,
		"start":      types.StringType,
		"end":        types.StringType,
		"retry_policy": types.ObjectType{AttrTypes: map[string]attr.Type{
			"maximum_event_age": types.Int64Type,
			"maximum_retry":     types.Int64Type,
		}},
	}
}

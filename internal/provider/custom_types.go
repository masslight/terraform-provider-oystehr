package provider

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

var _ basetypes.StringTypable = RuntimeType{}

type RuntimeType struct {
	basetypes.StringType
}

func (t RuntimeType) String() string {
	return "Runtime"
}

func (t RuntimeType) ValueType(_ context.Context) attr.Value {
	return RuntimeValue{}
}

func (t RuntimeType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return RuntimeValue{in}, nil
}

func (t RuntimeType) ValueFromTerraform(ct context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ct, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value of type %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ct, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to convert value from string: %v", diags)
	}

	return stringValuable, nil
}

func (t RuntimeType) Equal(o attr.Type) bool {
	other, ok := o.(RuntimeType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t RuntimeType) GoType(_ context.Context) reflect.Type {
	return reflect.TypeOf("")
}

var _ basetypes.StringValuable = RuntimeValue{}
var _ xattr.ValidateableAttribute = RuntimeValue{}

type RuntimeValue struct {
	basetypes.StringValue
}

func (v RuntimeValue) Equal(o attr.Value) bool {
	otherVal, ok := o.(RuntimeValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(otherVal.StringValue)
}

func (v RuntimeValue) Type(ctx context.Context) attr.Type {
	return RuntimeType{}
}

func (v RuntimeValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	value := v.ValueString()

	// Iterate through client.ValidRuntimes to check for value
	for _, runtime := range client.ValidRuntimes {
		if runtime == value {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Runtime Value",
		fmt.Sprintf("Runtime must be one of: %v.", client.ValidRuntimes),
	)
}

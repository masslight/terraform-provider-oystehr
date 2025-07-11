package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

func convertListToStringSlice(list types.List) []string {
	elements := make([]string, len(list.Elements()))
	for i, elem := range list.Elements() {
		elements[i] = elem.(types.String).ValueString()
	}
	return elements
}

func convertStringSliceToList(ctx context.Context, slice []string) types.List {
	elements := make([]types.String, len(slice))
	for i, v := range slice {
		elements[i] = types.StringValue(v)
	}
	list, _ := types.ListValueFrom(ctx, types.StringType, elements)
	return list
}

func convertRoleStubSliceToList(ctx context.Context, slice []client.RoleStub) types.List {
	elements := make([]types.Object, len(slice))
	for i, v := range slice {
		elements[i] = convertRoleStubToObject(ctx, v)
	}
	list, _ := types.ListValueFrom(ctx, roleStubType, elements)
	return list
}

func convertRoleStubToObject(ctx context.Context, roleStub client.RoleStub) types.Object {
	if roleStub.ID == nil {
		return types.ObjectNull(roleStubAttributesType)
	}
	roleStubObj, _ := types.ObjectValueFrom(ctx, roleStubAttributesType, map[string]attr.Value{
		"id":   types.StringValue(*roleStub.ID),
		"name": types.StringValue(*roleStub.Name),
	})
	return roleStubObj
}

func convertListToRoleStubSlice(list types.List) []client.RoleStub {
	elements := make([]client.RoleStub, len(list.Elements()))
	for i, elem := range list.Elements() {
		var roleStub client.RoleStub
		if obj, ok := elem.(types.Object); ok {
			obj.As(context.Background(), &roleStub, basetypes.ObjectAsOptions{
				UnhandledNullAsEmpty:    true,
				UnhandledUnknownAsEmpty: true,
			})
		}
		elements[i] = roleStub
	}
	return elements
}

func tfStringToStringPointer(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	val := value.ValueString()
	return &val
}

func tfBoolToBoolPointer(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	val := value.ValueBool()
	return &val
}

func tfInt32ToInt32Pointer(value types.Int32) *int32 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	val := value.ValueInt32()
	return &val
}

func tfInt64ToInt64Pointer(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	val := value.ValueInt64()
	return &val
}

func stringPointerToTfString(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func boolPointerToTfBool(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}

func int32PointerToTfInt32(value *int32) types.Int32 {
	if value == nil {
		return types.Int32Null()
	}
	return types.Int32Value(*value)
}

func int64PointerToTfInt64(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}

func convertClientAccessPolicyToAccessPolicy(ctx context.Context, accessPolicy *client.AccessPolicy) types.Object {
	if accessPolicy == nil {
		return types.ObjectNull(map[string]attr.Type{})
	}
	accessPolicyObj, _ := types.ObjectValueFrom(ctx, accessPolicyAttributesType, accessPolicy)
	return accessPolicyObj
}

func convertAccessPolicyToClientAccessPolicy(ctx context.Context, accessPolicy types.Object) *client.AccessPolicy {
	// if accessPolicy.IsNull() || accessPolicy.IsUnknown() {
	// 	return nil
	// }
	var clientAccessPolicy client.AccessPolicy
	accessPolicy.As(ctx, &clientAccessPolicy, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	if clientAccessPolicy.Rule == nil {
		clientAccessPolicy.Rule = make([]client.Rule, 0)
	}
	return &clientAccessPolicy
}

func convertGoValueToTfValue(ctx context.Context, v any) (attr.Value, diag.Diagnostics) {
	switch vv := v.(type) {
	case string:
		return types.StringValue(vv), nil
	case bool:
		return types.BoolValue(vv), nil
	case int, int32, int64:
		return types.Int64Value(int64(vv.(int))), nil
	case float32, float64:
		return types.Float64Value(float64(vv.(float32))), nil
	case map[string]any:
		nestedValue, diags := convertGoMapToTfObject(ctx, vv)
		if diags.HasError() {
			return nil, diags
		}
		return nestedValue, nil
	case []any:
		elements := make([]attr.Value, len(vv))
		for i, elem := range vv {
			if elemMap, ok := elem.(map[string]any); ok {
				convertedElem, diags := convertGoMapToTfObject(ctx, elemMap)
				if diags.HasError() {
					return nil, diags
				}
				elements[i] = convertedElem
			} else {
				convertedElem, diags := convertGoValueToTfValue(ctx, elem)
				if diags.HasError() {
					return nil, diags
				}
				elements[i] = convertedElem
			}
		}
		return types.ListValueMust(types.StringType, elements), nil
	default:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			fmt.Sprintf("Unsupported type %T", v),
			fmt.Sprintf("The value of type %T is not supported for conversion to Terraform value.", v),
		)}
	}
}

func convertGoMapToTfObject(ctx context.Context, m map[string]any) (types.Object, diag.Diagnostics) {
	mapValues := make(map[string]attr.Value, len(m))
	mapTypes := make(map[string]attr.Type, len(m))
	for k, v := range m {
		value, diags := convertGoValueToTfValue(ctx, v)
		if diags.HasError() {
			return types.ObjectNull(map[string]attr.Type{}), diags
		}
		mapValues[k] = value
		mapTypes[k] = value.Type(ctx)
	}
	return types.ObjectValueFrom(ctx, mapTypes, mapValues)
}

func convertMapToDynamicValue(ctx context.Context, m map[string]any) (types.Dynamic, diag.Diagnostics) {
	if m == nil {
		return types.DynamicNull(), nil
	}
	if len(m) == 0 {
		return types.DynamicNull(), nil
	}
	objectValue, diags := convertGoMapToTfObject(ctx, m)
	if diags.HasError() {
		return types.DynamicNull(), diags
	}
	dynamicValue := types.DynamicValue(objectValue)
	return dynamicValue, nil
}

func convertTfValueToGoValue(ctx context.Context, value attr.Value) (any, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	switch val := value.(type) {
	case types.String:
		return val.ValueString(), nil
	case types.Bool:
		return val.ValueBool(), nil
	case types.Int64:
		return val.ValueInt64(), nil
	case types.Float64:
		return val.ValueFloat64(), nil
	case types.Object:
		return convertTfObjectToGoMap(ctx, val)
	case types.List:
		elements := make([]any, len(val.Elements()))
		for i, elem := range val.Elements() {
			if elem.IsNull() || elem.IsUnknown() {
				elements[i] = nil
				continue
			}
			elemValue, diags := convertTfValueToGoValue(ctx, elem)
			if diags.HasError() {
				return nil, diags
			}
			elements[i] = elemValue
		}
		return elements, nil
	default:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			fmt.Sprintf("Unsupported type %T", value),
			fmt.Sprintf("The value of type %T is not supported for conversion to Go map.", value),
		)}
	}
}

func convertTfObjectToGoMap(ctx context.Context, obj types.Object) (map[string]any, diag.Diagnostics) {
	attrs := obj.Attributes()
	m := make(map[string]any, len(attrs))
	for k, v := range attrs {
		val, diags := convertTfValueToGoValue(ctx, v)
		if diags.HasError() {
			return nil, diags
		}
		m[k] = val
	}
	return m, nil
}

func convertDynamicValueToMap(ctx context.Context, value types.Dynamic) (map[string]any, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() || value.IsUnderlyingValueNull() || value.IsUnderlyingValueUnknown() {
		return nil, nil
	}
	uv := value.UnderlyingValue()
	switch uv := uv.(type) {
	case basetypes.ObjectValue:
		convertedObj, diags := convertTfObjectToGoMap(ctx, uv)
		if diags.HasError() {
			return nil, diags
		}
		return convertedObj, nil
	default:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Dynamic Value Type",
			"The dynamic value must be an object type.",
		)}
	}
}

package provider

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

func getStringFromValue(m attr.Value) string {
	if m == nil {
		return ""
	}
	switch mv := m.(type) {
	case basetypes.StringValue:
		return mv.ValueString()
	case basetypes.BoolValue:
		return fmt.Sprintf("%t", mv.ValueBool())
	case basetypes.Int64Value:
		return fmt.Sprintf("%d", mv.ValueInt64())
	case basetypes.Float64Value:
		return fmt.Sprintf("%f", mv.ValueFloat64())
	case basetypes.Int32Value:
		return fmt.Sprintf("%d", mv.ValueInt32())
	case basetypes.Float32Value:
		return fmt.Sprintf("%f", mv.ValueFloat32())
	case basetypes.NumberValue:
		return mv.ValueBigFloat().String()
	default:
		return ""
	}
}

func valueToTerraformValue(ctx context.Context, v any) (attr.Type, attr.Value, diag.Diagnostics) {
	if v == nil {
		return nil, nil, nil
	}

	var value attr.Value
	var ty attr.Type
	switch val := v.(type) {
	case string:
		value = types.StringValue(val)
		ty = types.StringType
	case bool:
		value = types.BoolValue(val)
		ty = types.BoolType
	case int64:
		value = types.Int64Value(val)
		ty = types.Int64Type
	case float64:
		value = types.Float64Value(val)
		ty = types.Float64Type
	case map[string]any:
		tflog.Debug(ctx, "Converting nested map to Terraform object", map[string]any{
			"map": val,
		})
		nestedObject, diags := mapToTerraformObject(ctx, val)
		if diags.HasError() {
			return nil, nil, diags
		}
		value = nestedObject
		ty = types.ObjectType{
			AttrTypes: nestedObject.AttributeTypes(ctx),
		}
	case []any:
		tflog.Debug(ctx, "Converting slice to Terraform list", map[string]any{
			"slice": val,
		})
		elements := make([]attr.Value, len(val))
		var nestedType attr.Type
		var nestedTypes []attr.Type
		for i, elem := range val {
			nt, nv, diags := valueToTerraformValue(ctx, elem)
			if diags.HasError() {
				return nil, nil, diags
			}
			elements[i] = nv
			if nestedType == nt {
				// still the same type
				nestedType = nt
			} else {
				// now we need a tuple
				if nestedTypes == nil {
					nestedTypes = make([]attr.Type, len(val))
					for j := 0; j < i; j++ {
						nestedTypes[j] = nestedType
					}
				}
				nestedTypes[i] = nt
				nestedType = nil
			}
		}
		if nestedType == nil {
			// we have a tuple of types
			tv, diags := types.TupleValue(nestedTypes, elements)
			if diags.HasError() {
				return nil, nil, diags
			}
			value = tv
			ty = types.TupleType{
				ElemTypes: nestedTypes,
			}
		} else {
			// we have a single type
			lv, diags := types.ListValue(nestedType, elements)
			if diags.HasError() {
				return nil, nil, diags
			}
			value = lv
			ty = types.ListType{
				ElemType: nestedType,
			}
		}
	default:
		diags := diag.NewErrorDiagnostic("Unsupported type", fmt.Sprintf("unsupported type for attr %+v: %T", val, v))
		return nil, nil, diag.Diagnostics{diags}
	}
	return ty, value, nil
}

func mapToTerraformObject(ctx context.Context, m map[string]any) (types.Object, diag.Diagnostics) {
	if m == nil {
		return types.ObjectNull(map[string]attr.Type{}), nil
	}
	attrMap := make(map[string]attr.Value, len(m))
	typeMap := make(map[string]attr.Type, len(m))
	for k, v := range m {
		ty, value, diags := valueToTerraformValue(ctx, v)
		if diags.HasError() {
			return types.ObjectNull(map[string]attr.Type{}), diags
		}
		attrMap[k] = value
		typeMap[k] = ty
	}
	tflog.Debug(ctx, "Converting map to Terraform object", map[string]any{
		"attrMap": attrMap,
		"typeMap": typeMap,
	})
	return types.ObjectValue(typeMap, attrMap)
}

func terraformValueToValue(ctx context.Context, v attr.Value) (any, diag.Diagnostics) {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil, nil
	}

	switch val := v.(type) {
	case types.String:
		return val.ValueString(), nil
	case types.Bool:
		return val.ValueBool(), nil
	case types.Int64:
		return val.ValueInt64(), nil
	case types.Float64:
		return val.ValueFloat64(), nil
	case types.Number:
		bf := val.ValueBigFloat()
		if bf.IsInt() {
			bfi, acc := bf.Int64()
			if acc == big.Exact {
				return bfi, nil
			}
			//fallthrough
		}
		bff, _ := bf.Float64()
		return bff, nil
	case types.Object:
		m := make(map[string]any, len(val.Attributes()))
		for k, elem := range val.Attributes() {
			nv, diags := terraformValueToValue(ctx, elem)
			if diags.HasError() {
				return nil, diags
			}
			m[k] = nv
		}
		return m, nil
	case types.List:
		slice := make([]any, len(val.Elements()))
		for i, elem := range val.Elements() {
			nv, diags := terraformValueToValue(ctx, elem)
			if diags.HasError() {
				return nil, diags
			}
			slice[i] = nv
		}
		return slice, nil
	case types.Tuple:
		slice := make([]any, len(val.Elements()))
		for i, elem := range val.Elements() {
			nv, diags := terraformValueToValue(ctx, elem)
			if diags.HasError() {
				return nil, diags
			}
			slice[i] = nv
		}
		return slice, nil
	default:
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unsupported type", fmt.Sprintf("unsupported type for attr %+v: %T", val, v))}
	}
}

func terraformObjectToMap(ctx context.Context, obj types.Object) (map[string]any, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	m := make(map[string]any, len(obj.Attributes()))
	for k, elem := range obj.Attributes() {
		nv, diags := terraformValueToValue(ctx, elem)
		if diags.HasError() {
			return nil, diags
		}
		m[k] = nv
	}
	return m, nil
}

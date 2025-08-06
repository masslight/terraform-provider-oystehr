package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

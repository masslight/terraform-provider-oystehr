package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func convertListToStringSlice(list types.List) []string {
	allowedCallbackUrls := make([]string, len(list.Elements()))
	for i, elem := range list.Elements() {
		allowedCallbackUrls[i] = elem.(types.String).ValueString()
	}
	return allowedCallbackUrls
}

func convertStringSliceToList(ctx context.Context, slice []string) types.List {
	elements := make([]types.String, len(slice))
	for i, v := range slice {
		elements[i] = types.StringValue(v)
	}
	list, _ := types.ListValueFrom(ctx, types.StringType, elements)
	return list
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

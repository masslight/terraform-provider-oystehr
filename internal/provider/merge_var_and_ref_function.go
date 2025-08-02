package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	// regexpVar is a regular expression to match variables in the form of #{var/VAR_NAME}
	regexpVar = regexp.MustCompile(`\#\{var/([^\}]+)\}`)
	// regexpRef is a regular expression to match references in the form of #{ref/component/KEY/property.access.path}
	regexpRef = regexp.MustCompile(`\#\{ref/([^\}/]+)/([^\}/]+)/([^\}]+)\}`)
)

var _ function.Function = &MergeVarAndRefFunction{}

type MergeVarAndRefFunction struct{}

func NewMergeVarAndRefFunction() function.Function {
	return &MergeVarAndRefFunction{}
}

func (f *MergeVarAndRefFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "merge_var_and_ref"
}

func (f *MergeVarAndRefFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Merge variables and references into a template string",
		Description: "Merge values into a template string from variables and references",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:           "value",
				Description:    "The template string to merge values into. It can contain variables and references in the form of #{var/VAR_NAME} and #{ref/component/KEY/property.access.path}.",
				AllowNullValue: true,
			},
			function.MapParameter{
				Name:        "resource",
				Description: "A map of component names to their resource values. Used to resolve references in the template string.",
				ElementType: basetypes.ObjectType{
					AttrTypes: map[string]attr.Type{
						"resource": basetypes.StringType{},
						"instance": basetypes.StringType{},
					},
				},
			},
			function.DynamicParameter{
				Name:        "variables",
				Description: "A map of variable names to their values.",
			},
			function.DynamicParameter{
				Name:        "spec",
				Description: "A dynamic object parameter that contains components and their properties. ",
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (f *MergeVarAndRefFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var valueParam basetypes.DynamicValue
	var resourceMapParam basetypes.MapValue
	var variables basetypes.DynamicValue
	var spec basetypes.DynamicValue

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &valueParam, &resourceMapParam, &variables, &spec))
	if resp.Error != nil {
		return
	}

	if valueParam.IsNull() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, valueParam))
		return
	}

	var value string
	switch v := valueParam.UnderlyingValue().(type) {
	case basetypes.StringValue:
		value = v.ValueString()
	default:
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, valueParam))
		return
	}

	if resourceMapParam.IsNull() || variables.IsNull() || spec.IsNull() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("resourceMap, variables, and spec must not be null"))
		return
	}

	var vars map[string]attr.Value
	switch variablesUnderlying := variables.UnderlyingValue().(type) {
	case basetypes.MapValue:
		vars = variablesUnderlying.Elements()
	case basetypes.ObjectValue:
		vars = variablesUnderlying.Attributes()
	default:
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("variables must be a map, got %T", variablesUnderlying)))
		return
	}

	// Vars
	matches := regexpVar.FindAllStringSubmatch(value, -1)
	if len(matches) > 0 {
		tflog.Debug(ctx, "Found variable matches", map[string]interface{}{
			"value":   value,
			"matches": matches,
		})
	}
	for _, match := range matches {
		varName := match[1]
		vv, ok := vars[varName]
		if ok {
			value = strings.Replace(value, match[0], getStringFromValue(vv), 1)
		}
	}

	// Refs
	matches = regexpRef.FindAllStringSubmatch(value, -1)
	if len(matches) > 0 {
		tflog.Debug(ctx, "Found reference matches", map[string]interface{}{
			"value":   value,
			"matches": matches,
		})
		// matches=["[#{ref/fhirResources/LOCATION_TELEMED_OH/id} fhirResources LOCATION_TELEMED_OH id]"]
		// "{\"active\":true,\"actor\":[{\"reference\":\"#{ref/fhirResources/LOCATION_TELEMED_OH/id}\"}],..."
	}
	for _, match := range matches {
		component := match[1]
		key := match[2]
		property := match[3]
		rc, ok := getValueOrValueFromMap(ctx, spec.UnderlyingValue(), component)
		if ok {
			tflog.Debug(ctx, "Found component in spec", map[string]interface{}{
				"component": component,
				"value":     value,
			})
			_, ok := getValueOrValueFromMap(ctx, rc, key)
			if ok {
				tflog.Debug(ctx, "Found key in spec", map[string]interface{}{
					"component": component,
					"key":       key,
					"value":     value,
				})
				ref, err := getTerraformRef(ctx, resourceMapParam.Elements(), component, key, property)
				if err != nil {
					resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))
					return
				}
				value = strings.Replace(value, match[0], ref, 1)
			}
		}
	}

	tflog.Debug(ctx, "Final merged value", map[string]interface{}{
		"value": value,
	})
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, value))
}

func getValueOrValueFromMap(ctx context.Context, m attr.Value, key string) (attr.Value, bool) {
	if m == nil {
		tflog.Debug(ctx, "Map value is nil", map[string]interface{}{
			"key": key,
		})
		return nil, false
	}
	switch mv := m.(type) {
	case basetypes.MapValue:
		tflog.Debug(ctx, "Found map value", map[string]interface{}{
			"mv":  mv,
			"key": key,
		})
		v, ok := mv.Elements()[key]
		return v, ok
	case basetypes.ObjectValue:
		tflog.Debug(ctx, "Found object value", map[string]interface{}{
			"mv":  mv,
			"key": key,
		})
		v, ok := mv.Attributes()[key]
		return v, ok
	default:
		tflog.Debug(ctx, "Found scalar value", map[string]interface{}{
			"mv":   mv,
			"type": fmt.Sprintf("%T", mv),
			"key":  key,
		})
		return m, false
	}
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

func getTerraformRef(ctx context.Context, resourceMap map[string]attr.Value, component, key, property string) (string, error) {
	rcValue, ok := resourceMap[component]
	if !ok {
		return "", fmt.Errorf("component %s not found in resource map", component)
	}
	resource, _ := getValueOrValueFromMap(ctx, rcValue, "resource")
	instance, _ := getValueOrValueFromMap(ctx, rcValue, "instance")
	partReferences := []string{}
	for _, part := range strings.Split(property, ".") {
		partReferences = append(partReferences, fmt.Sprintf("[\"%s\"]", part))
	}
	return fmt.Sprintf("${%s.%s%s%s}", getStringFromValue(resource), getStringFromValue(instance), fmt.Sprintf("[\"%s\"]", key), strings.Join(partReferences, "")), nil
}

package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
			function.StringParameter{
				Name:        "value",
				Description: "The template string to merge values into. It can contain variables and references in the form of #{var/VAR_NAME} and #{ref/component/KEY/property.access.path}.",
			},
			function.StringParameter{
				Name:        "resource",
				Description: "The Oystehr Terraform provider resource. Used for constructing references.",
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
		Return: function.StringReturn{},
	}
}

func (f *MergeVarAndRefFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var value string
	var resource string
	var variables basetypes.DynamicValue
	var spec basetypes.DynamicValue

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &value, &resource, &variables, &spec))

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
	for _, match := range matches {
		varName := match[1]
		vv, ok := vars[varName]
		if ok {
			value = strings.Replace(value, match[0], getStringFromValue(vv), 1)
		}
	}

	// Refs
	matches = regexpRef.FindAllStringSubmatch(value, -1)
	for _, match := range matches {
		component := match[1]
		key := match[2]
		property := match[3]
		rc, ok := getValueOrValueFromMap(spec.UnderlyingValue(), component)
		if ok {
			_, ok := getValueOrValueFromMap(rc, key)
			if ok {
				value = strings.Replace(value, match[0], getTerraformRef(resource, component, key, property), 1)
			}
		}
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, value))
}

func getValueOrValueFromMap(m attr.Value, key string) (attr.Value, bool) {
	if m == nil {
		return nil, false
	}
	switch mv := m.(type) {
	case basetypes.MapValue:
		v, ok := mv.Elements()[key]
		return v, ok
	case basetypes.ObjectValue:
		v, ok := mv.Attributes()[key]
		return v, ok
	default:
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

func getTerraformRef(resource, component, key, property string) string {
	partReferences := []string{}
	for _, part := range strings.Split(property, ".") {
		partReferences = append(partReferences, fmt.Sprintf("[\"%s\"]", part))
	}
	return fmt.Sprintf("${%s.%s%s%s}", resource, component, fmt.Sprintf("[\"%s\"]", key), strings.Join(partReferences, ""))
}

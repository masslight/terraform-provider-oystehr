package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	accessPolicyAttributes = map[string]schema.Attribute{
		"rule": schema.ListNestedAttribute{
			Optional:    true,
			Description: "A list of rules in the access policy.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"resource": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The resources the rule applies to.",
					},
					"action": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The actions the rule allows or denies.",
					},
					"effect": schema.StringAttribute{
						Required:    true,
						Description: "The effect of the rule (Allow or Deny).",
					},
					"condition": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "Conditions for the rule.",
					},
				},
			},
		},
	}
	ruleAttributesType = map[string]attr.Type{
		"resource":  types.ListType{ElemType: types.StringType},
		"action":    types.ListType{ElemType: types.StringType},
		"effect":    types.StringType,
		"condition": types.MapType{ElemType: types.StringType},
	}
	ruleType = types.ObjectType{
		AttrTypes: ruleAttributesType,
	}
	accessPolicyAttributesType = map[string]attr.Type{
		"rule": types.ListType{ElemType: ruleType},
	}
	accessPolicyType = types.ObjectType{
		AttrTypes: accessPolicyAttributesType,
	}
	roleStubAttributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required:    true,
			Description: "The ID of the role.",
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name of the role.",
		},
	}
	roleStubAttributesType = map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	}
	roleStubType = types.ObjectType{
		AttrTypes: roleStubAttributesType,
	}
)

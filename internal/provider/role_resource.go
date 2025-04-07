package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type RoleResource struct {
	client *client.Client
}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

func (r *RoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_role"
}

func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the role.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the role.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the role.",
			},
			"access_policy": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The access policy associated with the role.",
				Attributes: map[string]schema.Attribute{
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
				},
			},
		},
	}
}

func (r *RoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			"Expected *client.Client but got a different type.",
		)
		return
	}

	r.client = client
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan client.Role

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdRole, err := r.client.Role.CreateRole(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Role", err.Error())
		return
	}

	resp.State.Set(ctx, createdRole)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state client.Role

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.Role.GetRole(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Role", err.Error())
		return
	}

	resp.State.Set(ctx, role)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan client.Role

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedRole, err := r.client.Role.UpdateRole(ctx, plan.ID, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Role", err.Error())
		return
	}

	resp.State.Set(ctx, updatedRole)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state client.Role

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Role.DeleteRole(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Role", err.Error())
		return
	}
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type Role struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	AccessPolicy types.Object `tfsdk:"access_policy"`
}

func convertRoleToClientRole(ctx context.Context, role Role) client.Role {
	return client.Role{
		ID:           tfStringToStringPointer(role.ID),
		Name:         tfStringToStringPointer(role.Name),
		Description:  tfStringToStringPointer(role.Description),
		AccessPolicy: convertAccessPolicyToClientAccessPolicy(ctx, role.AccessPolicy),
	}
}

func convertClientRoleToRole(ctx context.Context, clientRole *client.Role) Role {
	return Role{
		ID:           stringPointerToTfString(clientRole.ID),
		Name:         stringPointerToTfString(clientRole.Name),
		Description:  stringPointerToTfString(clientRole.Description),
		AccessPolicy: convertClientAccessPolicyToAccessPolicy(ctx, clientRole.AccessPolicy),
	}
}

var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithConfigure = &RoleResource{}

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
				Required:    true,
				Description: "The access policy associated with the role.",
				Attributes:  accessPolicyAttributes,
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
	var plan Role

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role := convertRoleToClientRole(ctx, plan)

	createdRole, err := r.client.Role.CreateRole(ctx, &role)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Role", err.Error())
		return
	}

	retRole := convertClientRoleToRole(ctx, createdRole)

	resp.State.Set(ctx, retRole)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Role

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.Role.GetRole(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Role", err.Error())
		return
	}

	retRole := convertClientRoleToRole(ctx, role)

	resp.State.Set(ctx, retRole)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Role
	var state Role

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role := convertRoleToClientRole(ctx, plan)

	updatedRole, err := r.client.Role.UpdateRole(ctx, state.ID.ValueString(), &role)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Role", err.Error())
		return
	}

	retRole := convertClientRoleToRole(ctx, updatedRole)

	resp.State.Set(ctx, retRole)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Role

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Role.DeleteRole(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Role", err.Error())
		return
	}
}

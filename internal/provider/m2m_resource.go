package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type M2M struct {
	ID           types.String `tfsdk:"id"`
	ClientID     types.String `tfsdk:"client_id"`
	Profile      types.String `tfsdk:"profile"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	AccessPolicy types.Object `tfsdk:"access_policy"`
	Roles        types.List   `tfsdk:"roles"`
	JwksURL      types.String `tfsdk:"jwks_url"`
}

func convertM2MToClientM2M(ctx context.Context, m2m M2M) client.M2M {
	return client.M2M{
		ID:           tfStringToStringPointer(m2m.ID),
		ClientID:     tfStringToStringPointer(m2m.ClientID),
		Profile:      tfStringToStringPointer(m2m.Profile),
		Name:         tfStringToStringPointer(m2m.Name),
		Description:  tfStringToStringPointer(m2m.Description),
		AccessPolicy: convertAccessPolicyToClientAccessPolicy(ctx, m2m.AccessPolicy),
		Roles:        convertListToRoleStubSlice(m2m.Roles),
		JwksURL:      tfStringToStringPointer(m2m.JwksURL),
	}
}

func convertClientM2MToM2M(ctx context.Context, clientM2M *client.M2M) M2M {
	return M2M{
		ID:           stringPointerToTfString(clientM2M.ID),
		ClientID:     stringPointerToTfString(clientM2M.ClientID),
		Profile:      stringPointerToTfString(clientM2M.Profile),
		Name:         stringPointerToTfString(clientM2M.Name),
		Description:  stringPointerToTfString(clientM2M.Description),
		AccessPolicy: convertClientAccessPolicyToAccessPolicy(ctx, clientM2M.AccessPolicy),
		Roles:        convertRoleStubSliceToList(ctx, clientM2M.Roles),
		JwksURL:      stringPointerToTfString(clientM2M.JwksURL),
	}
}

type M2MResource struct {
	client *client.Client
}

func NewM2MResource() resource.Resource {
	return &M2MResource{}
}

func (r *M2MResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_m2m"
}

func (r *M2MResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the M2M resource.",
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "The client ID of the M2M resource.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the M2M resource.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the M2M resource.",
			},
			"access_policy": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The access policy associated with the M2M resource.",
				Attributes:  accessPolicyAttributes,
			},
			"roles": schema.ListNestedAttribute{
				Description: "A list of roles associated with the M2M resource.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: roleStubAttributes,
				},
			},
			"profile": schema.StringAttribute{
				Computed:    true,
				Description: "The profile associated with the M2M resource.",
			},
			"jwks_url": schema.StringAttribute{
				Optional:    true,
				Description: "The JWKS URL for the M2M resource.",
			},
		},
	}
}

func (r *M2MResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *M2MResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan M2M

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	m2m := convertM2MToClientM2M(ctx, plan)

	createdM2M, err := r.client.M2M.CreateM2M(ctx, &m2m)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating M2M", err.Error())
		return
	}

	retM2M := convertClientM2MToM2M(ctx, createdM2M)

	diags = resp.State.Set(ctx, retM2M)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *M2MResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state M2M

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	m2m, err := r.client.M2M.GetM2M(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading M2M", err.Error())
		return
	}

	retM2M := convertClientM2MToM2M(ctx, m2m)

	resp.State.Set(ctx, retM2M)
}

func (r *M2MResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan M2M
	var state M2M

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	m2m := convertM2MToClientM2M(ctx, plan)

	updatedM2M, err := r.client.M2M.UpdateM2M(ctx, state.ID.ValueString(), &m2m)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating M2M", err.Error())
		return
	}

	retM2M := convertClientM2MToM2M(ctx, updatedM2M)

	resp.State.Set(ctx, retM2M)
}

func (r *M2MResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state M2M

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.M2M.DeleteM2M(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting M2M", err.Error())
		return
	}
}

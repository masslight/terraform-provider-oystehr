package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

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
			"profile": schema.StringAttribute{
				Required:    true,
				Description: "The profile associated with the M2M resource.",
			},
			"jwks_url": schema.StringAttribute{
				Optional:    true,
				Description: "The JWKS URL for the M2M resource.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the M2M resource.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the M2M resource.",
			},
			"roles": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of roles associated with the M2M resource.",
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
	var plan client.M2M

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdM2M, err := r.client.M2M.CreateM2M(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating M2M", err.Error())
		return
	}

	resp.State.Set(ctx, createdM2M)
}

func (r *M2MResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state client.M2M

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	m2m, err := r.client.M2M.GetM2M(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading M2M", err.Error())
		return
	}

	resp.State.Set(ctx, m2m)
}

func (r *M2MResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan client.M2M

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedM2M, err := r.client.M2M.UpdateM2M(ctx, plan.ID, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating M2M", err.Error())
		return
	}

	resp.State.Set(ctx, updatedM2M)
}

func (r *M2MResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state client.M2M

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.M2M.DeleteM2M(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting M2M", err.Error())
		return
	}
}

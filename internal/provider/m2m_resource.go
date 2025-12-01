package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type M2M struct {
	ID                  types.String `tfsdk:"id"`
	ClientID            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	ClientSecretVersion types.Int64  `tfsdk:"client_secret_version"`
	Profile             types.String `tfsdk:"profile"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	AccessPolicy        types.Object `tfsdk:"access_policy"`
	Roles               types.List   `tfsdk:"roles"`
	JwksURL             types.String `tfsdk:"jwks_url"`
}

func convertM2MToClientM2M(ctx context.Context, m2m M2M) client.M2M {
	return client.M2M{
		ID:           tfStringToStringPointer(m2m.ID),
		ClientID:     tfStringToStringPointer(m2m.ClientID),
		Profile:      tfStringToStringPointer(m2m.Profile),
		Name:         tfStringToStringPointer(m2m.Name),
		Description:  tfStringToStringPointer(m2m.Description),
		AccessPolicy: convertAccessPolicyToClientAccessPolicy(ctx, m2m.AccessPolicy),
		Roles:        convertListToStringSlice(m2m.Roles),
		JwksURL:      tfStringToStringPointer(m2m.JwksURL),
	}
}

func convertClientM2MToM2M(ctx context.Context, clientM2M *client.M2M, clientSecret *string, clientSecretVersion int64) M2M {
	return M2M{
		ID:                  stringPointerToTfString(clientM2M.ID),
		ClientID:            stringPointerToTfString(clientM2M.ClientID),
		ClientSecret:        types.StringPointerValue(clientSecret),
		ClientSecretVersion: types.Int64Value(clientSecretVersion),
		Profile:             stringPointerToTfString(clientM2M.Profile),
		Name:                stringPointerToTfString(clientM2M.Name),
		Description:         stringPointerToTfString(clientM2M.Description),
		AccessPolicy:        convertClientAccessPolicyToAccessPolicy(ctx, clientM2M.AccessPolicy),
		Roles:               convertStringSliceToList(ctx, clientM2M.Roles),
		JwksURL:             stringPointerToTfString(clientM2M.JwksURL),
	}
}

var _ resource.Resource = &M2MResource{}
var _ resource.ResourceWithConfigure = &M2MResource{}
var _ resource.ResourceWithModifyPlan = &M2MResource{}
var _ resource.ResourceWithIdentity = &M2MResource{}
var _ resource.ResourceWithImportState = &M2MResource{}

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "The client ID of the M2M resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the M2M resource.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the M2M resource.",
			},
			"client_secret": schema.StringAttribute{
				Computed:    true,
				Description: "The client secret of the M2M resource. This is only set on creation and when rotated through the API.",
				Sensitive:   true,
			},
			"client_secret_version": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Increment this value to trigger a rotation of the client secret.",
				Default:     int64default.StaticInt64(0),
			},
			"access_policy": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The access policy associated with the M2M resource.",
				Attributes:  accessPolicyAttributes,
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(accessPolicyAttributesType, map[string]attr.Value{
						"rule": types.ListValueMust(
							ruleType,
							[]attr.Value{},
						),
					}),
				),
			},
			"roles": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "A list of roles associated with the M2M resource.",
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"profile": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The profile associated with the M2M resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"jwks_url": schema.StringAttribute{
				Optional:    true,
				Description: "The JWKS URL for the M2M resource.",
			},
		},
	}
}

func (*M2MResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = idIdentitySchema
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

	clientSecret, err := r.client.M2M.RotateM2MSecret(ctx, *createdM2M.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Rotating M2M Secret", err.Error())
		return
	}

	retM2M := convertClientM2MToM2M(ctx, createdM2M, clientSecret, plan.ClientSecretVersion.ValueInt64())
	identity := IDIdentityModel{
		ID: retM2M.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retM2M)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *M2MResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state M2M
	var identity IDIdentityModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	var id string
	if !identity.ID.IsNull() {
		id = identity.ID.ValueString()
	} else {
		id = state.ID.ValueString()
	}

	m2m, err := r.client.M2M.GetM2M(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code: 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading M2M", err.Error())
		return
	}

	retM2M := convertClientM2MToM2M(ctx, m2m, state.ClientSecret.ValueStringPointer(), state.ClientSecretVersion.ValueInt64())
	retIdentity := IDIdentityModel{
		ID: retM2M.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retM2M)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
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

	clientSecret := state.ClientSecret.ValueStringPointer()
	if !state.ClientSecretVersion.IsNull() && state.ClientSecretVersion.ValueInt64() < plan.ClientSecretVersion.ValueInt64() {
		newSecret, err := r.client.M2M.RotateM2MSecret(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Rotating M2M Secret", err.Error())
			return
		}
		clientSecret = newSecret
	}

	retM2M := convertClientM2MToM2M(ctx, updatedM2M, clientSecret, plan.ClientSecretVersion.ValueInt64())

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

func (r *M2MResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *M2MResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state M2M
	var plan M2M

	if req.Plan.Raw.IsNull() {
		// If the plan is null, we cannot modify it, so we return early.
		resp.Plan = req.Plan
		return
	}

	if !req.State.Raw.IsNull() {
		diags := req.State.Get(ctx, &state)
		resp.Diagnostics.Append(diags...)
	}
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.State.Raw.IsNull() {
		// On create, always set client_secret to unknown
		plan.ClientSecret = types.StringUnknown()
	} else if !plan.ClientSecretVersion.IsNull() && state.ClientSecretVersion.ValueInt64() < plan.ClientSecretVersion.ValueInt64() {
		// If client_secret_version is set and has changed, rotate the client secret
		plan.ClientSecret = types.StringUnknown()
	} else {
		plan.ClientSecret = state.ClientSecret
	}

	resp.Plan.Set(ctx, &plan)
}

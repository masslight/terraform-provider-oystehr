package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type ApplicationResource struct {
	client *client.Client
}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

func (r *ApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_application"
}

func (r *ApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the application.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the application.",
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "The client ID of the application.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the application.",
			},
			"login_redirect_uri": schema.StringAttribute{
				Optional:    true,
				Description: "The login redirect URI for the application.",
			},
			"allowed_callback_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of allowed callback URLs.",
			},
			"allowed_logout_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of allowed logout URLs.",
			},
			"allowed_web_origins_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of allowed web origins URLs.",
			},
			"allowed_cors_origins_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of allowed CORS origins URLs.",
			},
			"login_with_email_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether login with email is enabled.",
			},
			"passwordless_sms": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether passwordless SMS is enabled.",
			},
			"mfa_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether MFA is enabled.",
			},
			"should_send_invite_email": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether an invite email should be sent.",
			},
			"logo_uri": schema.StringAttribute{
				Optional:    true,
				Description: "The logo URI for the application.",
			},
			"refresh_token_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether refresh tokens are enabled.",
			},
		},
	}
}

func (r *ApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan client.Application

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdApp, err := r.client.Application.CreateApplication(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application", err.Error())
		return
	}

	resp.State.Set(ctx, createdApp)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state client.Application

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.Application.GetApplication(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Application", err.Error())
		return
	}

	resp.State.Set(ctx, app)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan client.Application

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedApp, err := r.client.Application.UpdateApplication(ctx, plan.ID, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Application", err.Error())
		return
	}

	resp.State.Set(ctx, updatedApp)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state client.Application

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Application.DeleteApplication(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application", err.Error())
		return
	}
}

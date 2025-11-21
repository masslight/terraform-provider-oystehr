package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type Application struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	ClientID               types.String `tfsdk:"client_id"`
	ConnectionName         types.String `tfsdk:"connection_name"`
	Description            types.String `tfsdk:"description"`
	LoginRedirectURI       types.String `tfsdk:"login_redirect_uri"`
	LoginWithEmailEnabled  types.Bool   `tfsdk:"login_with_email_enabled"`
	AllowedCallbackUrls    types.List   `tfsdk:"allowed_callback_urls"`
	AllowedLogoutUrls      types.List   `tfsdk:"allowed_logout_urls"`
	AllowedWebOriginsUrls  types.List   `tfsdk:"allowed_web_origins_urls"`
	AllowedCORSOriginsUrls types.List   `tfsdk:"allowed_cors_origins_urls"`
	PasswordlessSMS        types.Bool   `tfsdk:"passwordless_sms"`
	MFAEnabled             types.Bool   `tfsdk:"mfa_enabled"`
	ShouldSendInviteEmail  types.Bool   `tfsdk:"should_send_invite_email"`
	LogoURI                types.String `tfsdk:"logo_uri"`
	RefreshTokenEnabled    types.Bool   `tfsdk:"refresh_token_enabled"`
}

func clientAppToApplication(ctx context.Context, app *client.Application) Application {
	return Application{
		ID:                     stringPointerToTfString(app.ID),
		Name:                   stringPointerToTfString(app.Name),
		ClientID:               stringPointerToTfString(app.ClientID),
		ConnectionName:         stringPointerToTfString(app.ConnectionName),
		Description:            stringPointerToTfString(app.Description),
		LoginRedirectURI:       stringPointerToTfString(app.LoginRedirectURI),
		LoginWithEmailEnabled:  boolPointerToTfBool(app.LoginWithEmailEnabled),
		AllowedCallbackUrls:    convertStringSliceToList(ctx, app.AllowedCallbackUrls),
		AllowedLogoutUrls:      convertStringSliceToList(ctx, app.AllowedLogoutUrls),
		AllowedWebOriginsUrls:  convertStringSliceToList(ctx, app.AllowedWebOriginsUrls),
		AllowedCORSOriginsUrls: convertStringSliceToList(ctx, app.AllowedCORSOriginsUrls),
		PasswordlessSMS:        boolPointerToTfBool(app.PasswordlessSMS),
		MFAEnabled:             boolPointerToTfBool(app.MFAEnabled),
		ShouldSendInviteEmail:  boolPointerToTfBool(app.ShouldSendInviteEmail),
		LogoURI:                stringPointerToTfString(app.LogoURI),
		RefreshTokenEnabled:    boolPointerToTfBool(app.RefreshTokenEnabled),
	}
}

func applicationToClientApp(plan Application) client.Application {
	allowedCallbackUrls := convertListToStringSlice(plan.AllowedCallbackUrls)
	allowedLogoutUrls := convertListToStringSlice(plan.AllowedLogoutUrls)
	allowedWebOriginsUrls := convertListToStringSlice(plan.AllowedWebOriginsUrls)
	allowedCORSOriginsUrls := convertListToStringSlice(plan.AllowedCORSOriginsUrls)
	app := client.Application{
		Name:                   tfStringToStringPointer(plan.Name),
		Description:            tfStringToStringPointer(plan.Description),
		LoginRedirectURI:       tfStringToStringPointer(plan.LoginRedirectURI),
		LoginWithEmailEnabled:  tfBoolToBoolPointer(plan.LoginWithEmailEnabled),
		AllowedCallbackUrls:    allowedCallbackUrls,
		AllowedLogoutUrls:      allowedLogoutUrls,
		AllowedWebOriginsUrls:  allowedWebOriginsUrls,
		AllowedCORSOriginsUrls: allowedCORSOriginsUrls,
		PasswordlessSMS:        tfBoolToBoolPointer(plan.PasswordlessSMS),
		MFAEnabled:             tfBoolToBoolPointer(plan.MFAEnabled),
		ShouldSendInviteEmail:  tfBoolToBoolPointer(plan.ShouldSendInviteEmail),
		LogoURI:                tfStringToStringPointer(plan.LogoURI),
		RefreshTokenEnabled:    tfBoolToBoolPointer(plan.RefreshTokenEnabled),
	}
	return app
}

var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithConfigure = &ApplicationResource{}
var _ resource.ResourceWithModifyPlan = &ApplicationResource{}
var _ resource.ResourceWithIdentity = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the application.",
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "The client ID of the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_name": schema.StringAttribute{
				Computed:    true,
				Description: "The connection name of the application, for use by frontend components.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Computed:    true,
				Description: "A list of allowed callback URLs.",
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"allowed_logout_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "A list of allowed logout URLs.",
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"allowed_web_origins_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "A list of allowed web origins URLs.",
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"allowed_cors_origins_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "A list of allowed CORS origins URLs.",
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"login_with_email_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether login with email is enabled.",
				Default:     booldefault.StaticBool(true),
			},
			"passwordless_sms": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether passwordless SMS is enabled.",
				Default:     booldefault.StaticBool(false),
			},
			"mfa_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether MFA is enabled.",
				Default:     booldefault.StaticBool(false),
			},
			"should_send_invite_email": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether an invite email should be sent.",
				Default:     booldefault.StaticBool(false),
			},
			"logo_uri": schema.StringAttribute{
				Optional:    true,
				Description: "The logo URI for the application.",
			},
			"refresh_token_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether refresh tokens are enabled.",
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *ApplicationResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = idIdentitySchema
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
	var plan Application

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := applicationToClientApp(plan)

	createdApp, err := r.client.Application.CreateApplication(ctx, &app)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application", err.Error())
		return
	}

	retApp := clientAppToApplication(ctx, createdApp)
	identity := IDIdentityModel{
		ID: retApp.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retApp)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Application
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

	app, err := r.client.Application.GetApplication(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Application", err.Error())
		return
	}

	retApp := clientAppToApplication(ctx, app)
	retIdentity := IDIdentityModel{
		ID: retApp.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retApp)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Application
	var state Application
	var identity IDIdentityModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	diags = req.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := applicationToClientApp(plan)

	updatedApp, err := r.client.Application.UpdateApplication(ctx, identity.ID.ValueString(), &app)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Application", err.Error())
		return
	}

	retApp := clientAppToApplication(ctx, updatedApp)

	resp.State.Set(ctx, retApp)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Application
	var identity IDIdentityModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	diags = req.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Application.DeleteApplication(ctx, identity.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application", err.Error())
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *ApplicationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state Application
	var plan Application

	if !req.State.Raw.IsNull() {
		diags := req.State.Get(ctx, &state)
		resp.Diagnostics.Append(diags...)
	}
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !req.State.Raw.IsNull() {
		if !state.LoginWithEmailEnabled.Equal(plan.LoginWithEmailEnabled) {
			plan.ConnectionName = types.StringUnknown()
		}
	}
	resp.Plan.Set(ctx, &plan)
}

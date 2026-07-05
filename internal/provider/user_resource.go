package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
	"github.com/masslight/terraform-provider-oystehr/internal/retry"
)

type User struct {
	ID              types.String `tfsdk:"id"`
	Email           types.String `tfsdk:"email"`
	Username        types.String `tfsdk:"username"`
	ApplicationID   types.String `tfsdk:"application_id"`
	ResourceType    types.String `tfsdk:"resource_type"`
	Profile         types.String `tfsdk:"profile"`
	Roles           types.List   `tfsdk:"roles"`
	Password        types.String `tfsdk:"password"`
	PasswordVersion types.Int64  `tfsdk:"password_version"`
}

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithConfigure = &UserResource{}
var _ resource.ResourceWithIdentity = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

type UserResource struct {
	client *client.Client
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_user"
}

func (*UserResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = idIdentitySchema
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The email address of the user. Changing this forces a new user to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "The username of the user. Defaults to the email address if not set.",
			},
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the application to invite the user to.",
			},
			"resource_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The FHIR resource type to create as the user's profile. Must be either 'Practitioner' or 'Patient'.",
				Default:     stringdefault.StaticString("Practitioner"),
			},
			"profile": schema.StringAttribute{
				Optional:    true,
				Description: "An existing FHIR profile reference to associate with the user. If set, no new profile resource is created.",
			},
			"roles": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of role IDs to assign to the user.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The password to set for the user. This value is write-only and is never read back from the API.",
			},
			"password_version": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Increment this value to trigger the password to be re-applied.",
				Default:     int64default.StaticInt64(0),
			},
		},
	}
}

func convertUserToClientInvite(plan User) client.UserInvite {
	invite := client.UserInvite{
		Email:         tfStringToStringPointer(plan.Email),
		Username:      tfStringToStringPointer(plan.Username),
		ApplicationID: tfStringToStringPointer(plan.ApplicationID),
		Profile:       tfStringToStringPointer(plan.Profile),
		Roles:         convertListToStringSlice(plan.Roles),
	}
	if !plan.ResourceType.IsNull() && !plan.ResourceType.IsUnknown() {
		invite.Resource = map[string]any{
			"resourceType": plan.ResourceType.ValueString(),
		}
	}
	return invite
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan User

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	invite := convertUserToClientInvite(plan)

	createdUser, err := r.client.User.InviteUser(ctx, &invite)
	if err != nil {
		resp.Diagnostics.AddError("Error Inviting User", err.Error())
		return
	}
	if createdUser == nil || createdUser.ID == nil {
		resp.Diagnostics.AddError("Error Inviting User", "invite response did not include a user ID")
		return
	}

	userID := *createdUser.ID

	// Setting the password may briefly fail while the newly invited user
	// propagates in Auth0, so retry on transient errors.
	_, err = retry.RetryWithBackoff(ctx, func() (struct{}, error) {
		return struct{}{}, r.client.User.ChangePassword(ctx, userID, plan.Password.ValueString())
	}, retry.RetryConfig{
		BaseBackoff:   retry.BaseBackoffDefault,
		MaxBackoff:    retry.MaxBackoffDefault,
		MaxDuration:   retry.MaxDurationDefault,
		MaxAttempts:   retry.MaxAttemptsDefault,
		DisableJitter: false,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Setting User Password", err.Error())
		return
	}

	retUser := plan
	retUser.ID = stringPointerToTfString(createdUser.ID)
	identity := IDIdentityModel{ID: retUser.ID}

	resp.Diagnostics.Append(resp.State.Set(ctx, retUser)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state User
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

	stateIdentity := IDIdentityModel{ID: state.ID}

	user, err := r.client.User.GetUser(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code: 404") {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.Append(resp.Identity.Set(ctx, stateIdentity)...)
			return
		}
		resp.Diagnostics.AddError("Error Reading User", err.Error())
		return
	}

	// Map metadata only; never read the password back from the API.
	retUser := state
	retUser.ID = stringPointerToTfString(user.ID)
	retUser.Email = stringPointerToTfString(user.Email)
	retIdentity := IDIdentityModel{ID: retUser.ID}

	resp.Diagnostics.Append(resp.State.Set(ctx, retUser)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
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

type UserResource struct {
	client *client.Client
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_user"
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

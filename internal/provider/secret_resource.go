package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type SecretIdentityModel struct {
	Name types.String `tfsdk:"name"`
}

type Secret struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func convertSecretToClientSecret(ctx context.Context, secret Secret) client.Secret {
	return client.Secret{
		Name:  tfStringToStringPointer(secret.Name),
		Value: tfStringToStringPointer(secret.Value),
	}
}

func convertClientSecretToSecret(ctx context.Context, clientSecret *client.Secret) Secret {
	return Secret{
		Name:  stringPointerToTfString(clientSecret.Name),
		Value: stringPointerToTfString(clientSecret.Value),
	}
}

var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithConfigure = &SecretResource{}
var _ resource.ResourceWithIdentity = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

type SecretResource struct {
	client *client.Client
}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

func (r *SecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_secret"
}

func (r *SecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the secret.",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "The value of the secret.",
				Sensitive:   true,
			},
		},
	}
}

func (*SecretResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"name": identityschema.StringAttribute{
				Description:       "The name of the secret.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *SecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Secret

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret := convertSecretToClientSecret(ctx, plan)

	createdSecret, err := r.client.Secret.SetSecret(ctx, &secret)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Secret", err.Error())
		return
	}

	retSecret := convertClientSecretToSecret(ctx, createdSecret)
	identity := SecretIdentityModel{
		Name: retSecret.Name,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retSecret)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Secret
	var identity SecretIdentityModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var name string
	if !identity.Name.IsNull() {
		name = identity.Name.ValueString()
	} else {
		name = state.Name.ValueString()
	}

	secret, err := r.client.Secret.GetSecret(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Secret", err.Error())
		return
	}

	retSecret := convertClientSecretToSecret(ctx, secret)
	retIdentity := SecretIdentityModel{
		Name: retSecret.Name,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retSecret)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Secret

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret := convertSecretToClientSecret(ctx, plan)

	updatedSecret, err := r.client.Secret.SetSecret(ctx, &secret)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Secret", err.Error())
		return
	}

	resp.State.Set(ctx, convertClientSecretToSecret(ctx, updatedSecret))
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Secret

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Secret.DeleteSecret(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Secret", err.Error())
		return
	}
}

func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("name"), path.Root("name"), req, resp)
}

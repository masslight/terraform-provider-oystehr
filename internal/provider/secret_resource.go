package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

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

	resp.State.Set(ctx, convertClientSecretToSecret(ctx, createdSecret))
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Secret

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.Secret.GetSecret(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Secret", err.Error())
		return
	}

	resp.State.Set(ctx, convertClientSecretToSecret(ctx, secret))
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

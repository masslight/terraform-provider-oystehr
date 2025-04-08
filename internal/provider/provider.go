package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

var _ provider.Provider = &OystehrProvider{}

type OystehrProviderModel struct {
	ProjectID    types.String `tfsdk:"project_id"`
	AccessToken  types.String `tfsdk:"access_token"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

type OystehrProvider struct {
	version string
}

func (o *OystehrProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oystehr"
	resp.Version = o.version
}

func (o *OystehrProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Oystehr project ID",
				Required:            true,
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Oystehr developer temporary access token",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Oystehr developer client ID",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Oystehr developer client secret",
				Optional:            true,
			},
		},
	}
}

func (o *OystehrProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OystehrProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	hasAccessToken := false
	hasClientCreds := false
	if !data.AccessToken.IsNull() {
		hasAccessToken = true
	}
	if !data.ClientID.IsNull() && !data.ClientSecret.IsNull() {
		hasClientCreds = true
	}
	// Either neither present or both present
	if hasAccessToken == hasClientCreds {
		resp.Diagnostics.AddAttributeError(path.Root("access_token"), "Misconfigured credentials", "Either access token or client ID and client secret must be known")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	client := client.New(client.ClientConfig{
		ProjectID:    data.ProjectID.ValueStringPointer(),
		AccessToken:  data.AccessToken.ValueStringPointer(),
		ClientID:     data.ClientID.ValueStringPointer(),
		ClientSecret: data.ClientSecret.ValueStringPointer(),
	})
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (o *OystehrProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (o *OystehrProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewApplicationResource,
		NewFhirResource,
		NewM2MResource,
		NewRoleResource,
		NewSecretResource,
		NewZambdaResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OystehrProvider{
			version: version,
		}
	}
}

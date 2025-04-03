package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &OystehrProvider{}

type OystehrProviderModel struct {
	Credentials types.String `tfsdk:"credentials"`
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
			"credentials": schema.StringAttribute{
				MarkdownDescription: "Oystehr credentials",
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
	if data.Credentials.IsNull() {
		// throw error
		return
	}

	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (o *OystehrProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (o *OystehrProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OystehrProvider{
			version: version,
		}
	}
}

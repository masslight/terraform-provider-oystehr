package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type ProjectDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	SignupEnabled        types.Bool   `tfsdk:"signup_enabled"`
	DefaultPatientRoleId types.String `tfsdk:"default_patient_role_id"`
	FhirVersion          types.String `tfsdk:"fhir_version"`
	Sandbox              types.Bool   `tfsdk:"sandbox"`
}

func clientProjectToProjectDataSource(ctx context.Context, project *client.Project) ProjectDataSourceModel {
	return ProjectDataSourceModel{
		ID:                   stringPointerToTfString(project.ID),
		Name:                 stringPointerToTfString(project.Name),
		Description:          stringPointerToTfString(project.Description),
		SignupEnabled:        boolPointerToTfBool(project.SignupEnabled),
		DefaultPatientRoleId: stringPointerToTfString(project.DefaultPatientRole.ID),
		FhirVersion:          stringPointerToTfString(project.FhirVersion),
		Sandbox:              boolPointerToTfBool(project.Sandbox),
	}
}

var _ datasource.DataSource = &ProjectDataSource{}
var _ datasource.DataSourceWithConfigure = &ProjectDataSource{}

type ProjectDataSource struct {
	client *client.Client
}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

func (d *ProjectDataSource) Metadata(ctx context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oystehr_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The project's unique identifier",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The project's name",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The project's description",
			},
			"signup_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether signup is enabled for the project",
			},
			"default_patient_role_id": schema.StringAttribute{
				Computed:    true,
				Description: "The default patient role ID for the project",
			},
			"fhir_version": schema.StringAttribute{
				Computed:    true,
				Description: "The FHIR version for the project.",
			},
			"sandbox": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the project is in sandbox mode.",
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	project, err := d.client.Project.GetProject(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Project",
			"Could not read project: "+err.Error(),
		)
		return
	}

	data = clientProjectToProjectDataSource(ctx, project)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

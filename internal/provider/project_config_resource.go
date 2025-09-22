package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type Project struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	SignupEnabled        types.Bool   `tfsdk:"signup_enabled"`
	DefaultPatientRoleId types.String `tfsdk:"default_patient_role_id"`
	FhirVersion          types.String `tfsdk:"fhir_version"`
	Sandbox              types.Bool   `tfsdk:"sandbox"`
}

func clientProjectToProject(ctx context.Context, project *client.Project) Project {
	return Project{
		ID:                   stringPointerToTfString(project.ID),
		Name:                 stringPointerToTfString(project.Name),
		Description:          stringPointerToTfString(project.Description),
		SignupEnabled:        boolPointerToTfBool(project.SignupEnabled),
		DefaultPatientRoleId: stringPointerToTfString(project.DefaultPatientRole.ID),
		FhirVersion:          stringPointerToTfString(project.FhirVersion),
		Sandbox:              boolPointerToTfBool(project.Sandbox),
	}
}

func projectToClientProjectUpdateParams(plan Project) client.ProjectUpdateParams {
	project := client.ProjectUpdateParams{
		Name:                 tfStringToStringPointer(plan.Name),
		Description:          tfStringToStringPointer(plan.Description),
		SignupEnabled:        tfBoolToBoolPointer(plan.SignupEnabled),
		DefaultPatientRoleId: tfStringToStringPointer(plan.DefaultPatientRoleId),
	}
	return project
}

var _ resource.Resource = &ProjectConfigResource{}
var _ resource.ResourceWithConfigure = &ProjectConfigResource{}
var _ resource.ResourceWithIdentity = &ProjectConfigResource{}

type ProjectConfigResource struct {
	client *client.Client
}

func NewProjectConfigResource() resource.Resource {
	return &ProjectConfigResource{}
}

func (r *ProjectConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_project_configuration"
}

func (r *ProjectConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the project.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the project.",
			},
			"signup_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether signup is enabled for the project.",
			},
			"default_patient_role_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The default patient role ID for the project.",
			},
			"fhir_version": schema.StringAttribute{
				Computed:    true,
				Description: "The FHIR version for the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sandbox": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the project is in sandbox mode.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectConfigResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = idIdentitySchema
}

func (r *ProjectConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Project

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateParams := projectToClientProjectUpdateParams(plan)

	updatedProject, err := r.client.Project.UpdateProject(ctx, &updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Project Configuration", err.Error())
		return
	}

	retProject := clientProjectToProject(ctx, updatedProject)

	resp.State.Set(ctx, retProject)
}

func (r *ProjectConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Project
	var identity IDIdentityModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.Project.GetProject(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Project Configuration", err.Error())
		return
	}

	retProject := clientProjectToProject(ctx, project)
	retIdentity := IDIdentityModel{
		ID: retProject.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retProject)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

func (r *ProjectConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Project

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateParams := projectToClientProjectUpdateParams(plan)

	updatedProject, err := r.client.Project.UpdateProject(ctx, &updateParams)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Project Configuration", err.Error())
		return
	}

	retProject := clientProjectToProject(ctx, updatedProject)

	resp.State.Set(ctx, retProject)
}

func (r *ProjectConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Deleting Project Configuration re-sets the project to its default state.
	signupEnabled := false
	_, err := r.client.Project.UpdateProject(ctx, &client.ProjectUpdateParams{
		SignupEnabled:        &signupEnabled,
		DefaultPatientRoleId: nil,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Project Configuration", err.Error())
		return
	}
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type FhirResource struct {
	client *client.Client
}

func NewFhirResource() resource.Resource {
	return &FhirResource{}
}

func (r *FhirResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_fhir_resource"
}

func (r *FhirResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The FHIR resource type (e.g., Patient, Observation).",
			},
			"data": schema.DynamicAttribute{
				Required:    true,
				Description: "The FHIR resource data.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the FHIR resource.",
			},
		},
	}
}

func (r *FhirResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			"Expected *sdk.Client but got a different type.",
		)
		return
	}

	r.client = client
}

func (r *FhirResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan struct {
		Type types.String `tfsdk:"type"`
		Data types.Map    `tfsdk:"data"`
	}

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceData := make(map[string]interface{})
	plan.Data.ElementsAs(ctx, &resourceData, false)

	resource, err := r.client.CreateFhirResource(ctx, plan.Type.ValueString(), resourceData)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating FHIR Resource", err.Error())
		return
	}

	resp.State.Set(ctx, map[string]interface{}{
		"type": plan.Type,
		"data": plan.Data,
		"id":   types.StringValue(resource["id"].(string)),
	})
}

func (r *FhirResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state struct {
		Type types.String `tfsdk:"type"`
		ID   types.String `tfsdk:"id"`
	}

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := r.client.GetFhirResource(ctx, state.Type.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading FHIR Resource", err.Error())
		return
	}

	mappedData, diag := types.MapValueFrom(ctx, types.MapType{}, resource)
	if diag.HasError() {
		return
	}

	resp.State.Set(ctx, map[string]interface{}{
		"type": state.Type,
		"data": mappedData,
		"id":   state.ID,
	})
}

func (r *FhirResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan struct {
		Type types.String `tfsdk:"type"`
		Data types.Map    `tfsdk:"data"`
		ID   types.String `tfsdk:"id"`
	}

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceData := make(map[string]interface{})
	plan.Data.ElementsAs(ctx, &resourceData, false)

	resource, err := r.client.UpdateFhirResource(ctx, plan.Type.ValueString(), plan.ID.ValueString(), resourceData)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating FHIR Resource", err.Error())
		return
	}

	mappedData, diag := types.MapValueFrom(ctx, types.MapType{}, resource)
	if diag.HasError() {
		return
	}

	resp.State.Set(ctx, map[string]interface{}{
		"type": plan.Type,
		"data": mappedData,
		"id":   plan.ID,
	})
}

func (r *FhirResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state struct {
		Type types.String `tfsdk:"type"`
		ID   types.String `tfsdk:"id"`
	}

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFhirResource(ctx, state.Type.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting FHIR Resource", err.Error())
		return
	}
}

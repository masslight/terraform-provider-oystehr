package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type FhirResourceData struct {
	ID   types.String  `tfsdk:"id"`
	Type types.String  `tfsdk:"type"`
	Data types.Dynamic `tfsdk:"data"`
}

func convertFhirResourceToRawResource(ctx context.Context, resourceData FhirResourceData) (map[string]any, diag.Diagnostics) {
	data, diags := convertDynamicValueToMap(ctx, resourceData.Data)
	if diags.HasError() {
		return nil, diags
	}
	return map[string]any{
		"id":           resourceData.ID.ValueString(),
		"resourceType": resourceData.Type.ValueString(),
		"data":         data,
	}, nil
}

func convertRawResourceToFhirResource(ctx context.Context, rawResource map[string]any) (FhirResourceData, diag.Diagnostics) {
	id := rawResource["id"].(string)
	resourceType, ok := rawResource["resourceType"].(string)
	if !ok {
		return FhirResourceData{}, diag.Diagnostics{diag.NewErrorDiagnostic(
			"resourceType field is missing or not a string",
			"Expected a string for the resourceType field.",
		)}
	}
	delete(rawResource, "id")
	delete(rawResource, "resourceType")
	data, diags := convertMapToDynamicValue(ctx, rawResource)
	if diags.HasError() {
		return FhirResourceData{}, diags
	}
	return FhirResourceData{
		ID:   types.StringValue(id),
		Type: types.StringValue(resourceType),
		Data: data,
	}, nil
}

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
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the FHIR resource.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The FHIR resource type (e.g., Patient, Observation).",
			},
			"data": schema.DynamicAttribute{
				Required:    true,
				Description: "The FHIR resource data.",
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
	var plan FhirResourceData

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceData, diags := convertFhirResourceToRawResource(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdResource, err := r.client.Fhir.CreateResource(ctx, plan.Type.ValueString(), resourceData)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating FHIR Resource", err.Error())
		return
	}

	resource, diags := convertRawResourceToFhirResource(ctx, createdResource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, resource)
}

func (r *FhirResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FhirResourceData

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	returnedResource, err := r.client.Fhir.GetResource(ctx, state.Type.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading FHIR Resource", err.Error())
		return
	}

	resource, diags := convertRawResourceToFhirResource(ctx, returnedResource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, resource)
}

func (r *FhirResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FhirResourceData
	var state FhirResourceData

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceData, diags := convertFhirResourceToRawResource(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedResource, err := r.client.Fhir.UpdateResource(ctx, state.Type.ValueString(), state.ID.ValueString(), resourceData)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating FHIR Resource", err.Error())
		return
	}

	resource, diags := convertRawResourceToFhirResource(ctx, updatedResource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, resource)
}

func (r *FhirResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FhirResourceData

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Fhir.DeleteResource(ctx, state.Type.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting FHIR Resource", err.Error())
		return
	}
}

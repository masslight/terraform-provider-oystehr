package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type FhirResourceIdentityModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

type FhirResourceData struct {
	ID            types.String `tfsdk:"id"`
	Type          types.String `tfsdk:"type"`
	Data          types.String `tfsdk:"data"`
	Meta          types.Object `tfsdk:"meta"`
	RemovalPolicy types.String `tfsdk:"removal_policy"`
	ManagedFields types.Set    `tfsdk:"managed_fields"`
}

func convertFhirResourceToRawResource(ctx context.Context, resourceData FhirResourceData) (map[string]any, diag.Diagnostics) {
	var data map[string]any
	err := json.Unmarshal([]byte(resourceData.Data.ValueString()), &data)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Failed to unmarshal FHIR resource data",
			fmt.Sprintf("Expected a valid JSON string for the FHIR resource data. Resource: %+v", resourceData),
		)}
	}
	if data == nil {
		data = make(map[string]any)
	}
	if resourceData.ID.IsNull() || resourceData.ID.IsUnknown() {
		delete(data, "id")
	} else {
		data["id"] = resourceData.ID.ValueString()
	}
	data["resourceType"] = resourceData.Type.ValueString()
	return data, nil
}

func convertRawResourceToFhirResource(ctx context.Context, rawResource map[string]any, templ FhirResourceData) (FhirResourceData, diag.Diagnostics) {
	// Extract ID and resourceType from rawResource
	id, ok := rawResource["id"].(string)
	if !ok {
		return FhirResourceData{}, diag.Diagnostics{diag.NewErrorDiagnostic(
			"ID field is missing or not a string",
			"Expected a string for the ID field.",
		)}
	}
	delete(rawResource, "id")
	resourceType, ok := rawResource["resourceType"].(string)
	if !ok {
		return FhirResourceData{}, diag.Diagnostics{diag.NewErrorDiagnostic(
			"resourceType field is missing or not a string",
			"Expected a string for the resourceType field.",
		)}
	}
	delete(rawResource, "resourceType")

	// Extract computed meta fields
	rawMeta, ok := rawResource["meta"].(map[string]any)
	if !ok {
		return FhirResourceData{}, diag.Diagnostics{diag.NewErrorDiagnostic(
			"meta field is missing or not a map",
			"Expected a map for the meta field.",
		)}
	}
	computedMeta, diags := types.ObjectValue(map[string]attr.Type{
		"last_updated": types.StringType,
		"version_id":   types.StringType,
	}, map[string]attr.Value{
		"last_updated": types.StringValue(rawMeta["lastUpdated"].(string)),
		"version_id":   types.StringValue(rawMeta["versionId"].(string)),
	})
	if diags.HasError() {
		return FhirResourceData{}, diags
	}
	// Remove computed fields from rawResource
	delete(rawMeta, "lastUpdated")
	delete(rawMeta, "versionId")
	if len(rawMeta) == 0 {
		delete(rawResource, "meta")
	} else {
		rawResource["meta"] = rawMeta
	}
	data, err := json.Marshal(rawResource)
	if err != nil {
		return FhirResourceData{}, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Failed to marshal FHIR resource data",
			"Expected a valid JSON object for the FHIR resource data.",
		)}
	}
	return FhirResourceData{
		ID:            types.StringValue(id),
		Type:          types.StringValue(resourceType),
		Data:          types.StringValue(string(data)),
		Meta:          computedMeta,
		RemovalPolicy: templ.RemovalPolicy,
		ManagedFields: templ.ManagedFields,
	}, nil
}

func mergeJSONStrings(base, update string, managedFields []string) (string, error) {
	var baseMap, updateMap map[string]any
	if err := json.Unmarshal([]byte(base), &baseMap); err != nil {
		return "", err
	}
	if err := json.Unmarshal([]byte(update), &updateMap); err != nil {
		return "", err
	}

	for k, v := range updateMap {
		// No managed fields specified, or the field is in the managed fields list
		if len(managedFields) == 0 || slices.Contains(managedFields, k) {
			baseMap[k] = v
		}
	}

	result, err := json.Marshal(baseMap)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

var _ resource.Resource = &FhirResource{}
var _ resource.ResourceWithConfigure = &FhirResource{}
var _ resource.ResourceWithIdentity = &FhirResource{}
var _ resource.ResourceWithModifyPlan = &FhirResource{}
var _ resource.ResourceWithImportState = &FhirResource{}

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				Required:    true,
				Description: "The FHIR resource data in JSON format.",
			},
			"meta": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Metadata about the FHIR resource.",
				Attributes: map[string]schema.Attribute{
					"last_updated": schema.StringAttribute{
						Computed:    true,
						Description: "The last updated timestamp of the FHIR resource.",
					},
					"version_id": schema.StringAttribute{
						Computed:    true,
						Description: "The version ID of the FHIR resource.",
					},
				},
			},
			"removal_policy": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The removal policy for the FHIR resource. Valid values are 'delete' and 'retain'. Defaults to 'delete'.",
				Default:     stringdefault.StaticString("delete"),
			},
			"managed_fields": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "A set of fields to be managed by the provider. This is used to restrict updates to only the specified fields. Defaults to an empty set, which means all fields are managed.",
				Default: setdefault.StaticValue(
					types.SetValueMust(types.StringType, []attr.Value{}),
				),
			},
		},
	}
}

func (*FhirResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				Description:       "The ID of the FHIR resource.",
				RequiredForImport: true,
			},
			"type": identityschema.StringAttribute{
				Description:       "The FHIR resource type (e.g., Patient, Observation).",
				RequiredForImport: true,
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

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	resource, diags := convertRawResourceToFhirResource(ctx, createdResource, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	identity := FhirResourceIdentityModel{
		ID:   resource.ID,
		Type: resource.Type,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, resource)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *FhirResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FhirResourceData
	var identity FhirResourceIdentityModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	returnedResource, err := r.client.Fhir.GetResource(ctx, state.Type.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading FHIR Resource", err.Error())
		return
	}

	resource, diags := convertRawResourceToFhirResource(ctx, returnedResource, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	retIdentity := FhirResourceIdentityModel{
		ID:   resource.ID,
		Type: resource.Type,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, resource)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

func (r *FhirResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FhirResourceData
	var state FhirResourceData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceData, diags := convertFhirResourceToRawResource(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var versionID string
	versionIDValue, ok := state.Meta.Attributes()["version_id"]
	if ok && !versionIDValue.IsNull() {
		versionID = getStringFromValue(versionIDValue)
	}

	var resource FhirResourceData
	if !plan.Data.Equal(state.Data) {
		updatedResource, err := r.client.Fhir.UpdateResource(ctx, state.Type.ValueString(), state.ID.ValueString(), versionID, resourceData)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating FHIR Resource", err.Error())
			return
		}

		convertedResource, diags := convertRawResourceToFhirResource(ctx, updatedResource, plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resource = convertedResource
	} else {
		resource = plan
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, resource)...)
}

func (r *FhirResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FhirResourceData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.RemovalPolicy.ValueString() == "delete" {
		err := r.client.Fhir.DeleteResource(ctx, state.Type.ValueString(), state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Deleting FHIR Resource", err.Error())
			return
		}
	}
}

func (r *FhirResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// If the plan is null, we cannot modify it, so we return early.
		resp.Plan = req.Plan
		return
	}

	var plan FhirResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove resourceType and ID
	var data map[string]any
	err := json.Unmarshal([]byte(plan.Data.ValueString()), &data)
	if err != nil {
		resp.Diagnostics.AddError("Error Unmarshalling FHIR Resource Data", err.Error())
		return
	}
	delete(data, "resourceType")
	delete(data, "id")
	dataBytes, err := json.Marshal(data)
	if err != nil {
		resp.Diagnostics.AddError("Error Marshalling FHIR Resource Data", err.Error())
		return
	}
	plan.Data = types.StringValue(string(dataBytes))

	if req.State.Raw.IsNull() {
		// If the state is null, there's nothing more to check against, so we return early.
		resp.Plan = req.Plan
		return
	}

	var state FhirResourceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ID never changes once set
	plan.ID = state.ID

	// Handle `managed_fields` and merging of `data`
	if !plan.Data.Equal(state.Data) {
		var managedFields []string
		resp.Diagnostics.Append(plan.ManagedFields.ElementsAs(ctx, &managedFields, true)...)
		if resp.Diagnostics.HasError() {
			return
		}
		mergedData, err := mergeJSONStrings(state.Data.ValueString(), plan.Data.ValueString(), managedFields)
		if err != nil {
			resp.Diagnostics.AddError("Error Merging FHIR Resource Data", err.Error())
			return
		}
		dataChanged := mergedData != state.Data.ValueString()
		plan.Data = types.StringValue(mergedData)
		if !dataChanged {
			plan.Meta = state.Meta
		}
	} else {
		plan.Meta = state.Meta
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *FhirResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		parts := strings.Split(req.ID, "/")
		if len(parts) == 2 {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(parts[1]))...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), types.StringValue(parts[0]))...)
			return
		} else {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				"Expected format is 'resourceType/id', but got: "+req.ID,
			)
			return
		}
	}

	var identity FhirResourceIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(identity.ID.ValueString()))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), types.StringValue(identity.Type.ValueString()))...)
}

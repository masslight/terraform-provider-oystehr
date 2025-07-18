package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
	"github.com/masslight/terraform-provider-oystehr/internal/fs"
)

type Z3Object struct {
	Bucket         types.String `tfsdk:"bucket"`
	Key            types.String `tfsdk:"key"`
	LastModified   types.String `tfsdk:"last_modified"`
	Source         types.String `tfsdk:"source"`
	SourceChecksum types.String `tfsdk:"source_checksum"`
}

type Z3ObjectIdentityModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

func convertClientObjectToZ3Object(clientObject *client.Object, bucket, source, sourceChecksum string) Z3Object {
	return Z3Object{
		Bucket:         types.StringValue(bucket),
		Key:            stringPointerToTfString(clientObject.Key),
		LastModified:   stringPointerToTfString(clientObject.LastModified),
		Source:         types.StringValue(source),
		SourceChecksum: types.StringValue(sourceChecksum),
	}
}

var _ resource.Resource = &Z3ObjectResource{}
var _ resource.ResourceWithConfigure = &Z3ObjectResource{}
var _ resource.ResourceWithModifyPlan = &Z3ObjectResource{}
var _ resource.ResourceWithIdentity = &Z3ObjectResource{}
var _ resource.ResourceWithImportState = &Z3ObjectResource{}

type Z3ObjectResource struct {
	client *client.Client
}

func NewZ3ObjectResource() resource.Resource {
	return &Z3ObjectResource{}
}

func (r *Z3ObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_z3_object"
}

func (r *Z3ObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Z3 bucket.",
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "The key of the Z3 object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "The source file path for the Z3 object.",
			},
			"source_checksum": schema.StringAttribute{
				Computed:    true,
				Description: "The checksum of the source file.",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "The last modified timestamp of the Z3 object.",
			},
		},
	}
}

func (*Z3ObjectResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"bucket": identityschema.StringAttribute{
				Description:       "The name of the Z3 bucket.",
				RequiredForImport: true,
			},
			"key": identityschema.StringAttribute{
				Description:       "The key of the Z3 object.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *Z3ObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *Z3ObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Z3Object
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Z3.UploadObject(ctx, plan.Bucket.ValueString(), plan.Key.ValueString(), plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Z3 Object",
			"Could not create Z3 object: "+err.Error(),
		)
		return
	}

	returnedObject, err := r.client.Z3.ListObject(ctx, plan.Bucket.ValueString(), plan.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Z3 Object",
			"Could not retrieve Z3 object: "+err.Error(),
		)
		return
	}

	result := convertClientObjectToZ3Object(returnedObject, plan.Bucket.ValueString(), plan.Source.ValueString(), plan.SourceChecksum.ValueString())
	identity := Z3ObjectIdentityModel{
		Bucket: types.StringValue(plan.Bucket.ValueString()),
		Key:    types.StringValue(plan.Key.ValueString()),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *Z3ObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Z3Object
	var identity Z3ObjectIdentityModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading Z3 Object", map[string]interface{}{
		"bucket": state.Bucket.ValueString(),
		"key":    state.Key.ValueString(),
	})
	object, err := r.client.Z3.ListObject(ctx, state.Bucket.ValueString(), state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Z3 Object",
			"Could not read Z3 object: "+err.Error(),
		)
		return
	}
	tflog.Info(ctx, "Retrieved Z3 Object", map[string]interface{}{
		"bucket": state.Bucket.ValueString(),
		"key":    state.Key.ValueString(),
		"object": object,
	})

	result := convertClientObjectToZ3Object(object, state.Bucket.ValueString(), state.Source.ValueString(), state.SourceChecksum.ValueString())
	retIdentity := Z3ObjectIdentityModel{
		Bucket: types.StringValue(state.Bucket.ValueString()),
		Key:    types.StringValue(state.Key.ValueString()),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

func (r *Z3ObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Z3Object
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Z3.UploadObject(ctx, plan.Bucket.ValueString(), plan.Key.ValueString(), plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Z3 Object",
			"Could not update Z3 object: "+err.Error(),
		)
		return
	}

	returnedObject, err := r.client.Z3.ListObject(ctx, plan.Bucket.ValueString(), plan.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Z3 Object",
			"Could not retrieve Z3 object: "+err.Error(),
		)
		return
	}

	result := convertClientObjectToZ3Object(returnedObject, plan.Bucket.ValueString(), plan.Source.ValueString(), plan.SourceChecksum.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *Z3ObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Z3Object
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Z3.DeleteObject(ctx, state.Bucket.ValueString(), state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Z3 Object",
			"Could not delete Z3 object: "+err.Error(),
		)
		return
	}
}

func (r *Z3ObjectResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan Z3Object
	var state Z3Object

	if req.Plan.Raw.IsNull() {
		// If the plan is null, we cannot modify it, so we return early.
		tflog.Info(ctx, "Z3 Object plan is null, skipping modification")
		resp.Plan = req.Plan
		return
	}

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if !req.State.Raw.IsNull() {
		diags = req.State.Get(ctx, &state)
	}
	resp.Diagnostics.Append(diags...)
	tflog.Info(ctx, "Modifying Z3 Object Plan", map[string]interface{}{
		"plan": map[string]interface{}{
			"bucket":          plan.Bucket.ValueString(),
			"key":             plan.Key.ValueString(),
			"source":          plan.Source.ValueString(),
			"source_checksum": plan.SourceChecksum.ValueString(),
			"last_modified":   plan.LastModified.ValueString(),
		},
		"plan is null": req.Plan.Raw.IsNull(),
		"state": map[string]interface{}{
			"bucket":          state.Bucket.ValueString(),
			"key":             state.Key.ValueString(),
			"source":          state.Source.ValueString(),
			"source_checksum": state.SourceChecksum.ValueString(),
			"last_modified":   state.LastModified.ValueString(),
		},
		"state is null": req.State.Raw.IsNull(),
	})
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Validating z3 source file path", map[string]interface{}{
		"source": plan.Source.ValueString(),
	})
	if plan.Source.ValueString() != "" {
		sourceChecksum, err := fs.Sha256HashFile(plan.Source.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Calculating Source Checksum", err.Error())
			return
		}
		tflog.Info(ctx, "Calculated z3 source checksum", map[string]interface{}{
			"source_checksum":       sourceChecksum,
			"state_source_checksum": state.SourceChecksum.ValueString(),
		})
		plan.SourceChecksum = types.StringValue(sourceChecksum)
		if sourceChecksum != state.SourceChecksum.ValueString() {
			plan.LastModified = types.StringUnknown()
		}
	}

	tflog.Info(ctx, "Final z3 source checksum", map[string]interface{}{
		"plan_source_checksum": plan.SourceChecksum.ValueString(),
	})
	resp.Plan.Set(ctx, &plan)
}

func (r *Z3ObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		bucket, key, found := strings.Cut(req.ID, "/")
		if found {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bucket"), types.StringValue(bucket))...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), types.StringValue(key))...)
			return
		} else {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				"Expected format is 'bucket/key' but got: "+req.ID,
			)
			return
		}
	}

	var identity Z3ObjectIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bucket"), types.StringValue(identity.Bucket.ValueString()))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), types.StringValue(identity.Key.ValueString()))...)
}

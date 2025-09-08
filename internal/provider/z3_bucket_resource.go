package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type Z3BucketIdentityModel struct {
	Name types.String `tfsdk:"name"`
}

type Z3Bucket struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	RemovalPolicy types.String `tfsdk:"removal_policy"`
}

func convertZ3BucketToClientBucket(bucket Z3Bucket) client.Bucket {
	return client.Bucket{
		ID:   tfStringToStringPointer(bucket.ID),
		Name: tfStringToStringPointer(bucket.Name),
	}
}

func convertClientBucketToZ3Bucket(clientBucket *client.Bucket, templ Z3Bucket) Z3Bucket {
	return Z3Bucket{
		ID:            stringPointerToTfString(clientBucket.ID),
		Name:          stringPointerToTfString(clientBucket.Name),
		RemovalPolicy: templ.RemovalPolicy,
	}
}

var _ resource.Resource = &Z3BucketResource{}
var _ resource.ResourceWithConfigure = &Z3BucketResource{}
var _ resource.ResourceWithIdentity = &Z3BucketResource{}
var _ resource.ResourceWithImportState = &Z3BucketResource{}

type Z3BucketResource struct {
	client *client.Client
}

func NewZ3BucketResource() resource.Resource {
	return &Z3BucketResource{}
}

func (r *Z3BucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_z3_bucket"
}

func (r *Z3BucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Z3 bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Z3 bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"removal_policy": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The removal policy for the Z3 bucket. Valid values are 'delete' and 'retain'. Defaults to 'delete'.",
				Default:     stringdefault.StaticString("delete"),
			},
		},
	}
}

func (R *Z3BucketResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"name": identityschema.StringAttribute{
				Description:       "The name of the Z3 bucket.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *Z3BucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *Z3BucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Z3Bucket
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientBucket := convertZ3BucketToClientBucket(plan)

	createdBucket, err := r.client.Z3.CreateBucket(ctx, &clientBucket)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Z3 Bucket", err.Error())
		return
	}

	retZ3Bucket := convertClientBucketToZ3Bucket(createdBucket, plan)
	identity := Z3BucketIdentityModel{
		Name: retZ3Bucket.Name,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retZ3Bucket)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *Z3BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Z3Bucket
	var identity Z3BucketIdentityModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	clientBucket, err := r.client.Z3.GetBucket(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Z3 Bucket", err.Error())
		return
	}

	retZ3Bucket := convertClientBucketToZ3Bucket(clientBucket, state)
	retIdentity := Z3BucketIdentityModel{
		Name: retZ3Bucket.Name,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retZ3Bucket)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

func (r *Z3BucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state Z3Bucket
	var plan Z3Bucket
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.Name.ValueString() != plan.Name.ValueString() {
		resp.Diagnostics.AddError(
			"Name Change Not Allowed",
			"The name of a Z3 bucket cannot be changed after creation. Please create a new bucket with the desired name.",
		)
		return
	}
	retZ3Bucket := Z3Bucket{
		ID:            state.ID,
		Name:          state.Name,
		RemovalPolicy: plan.RemovalPolicy,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, retZ3Bucket)...)
}

func (r *Z3BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Z3Bucket
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.RemovalPolicy.ValueString() == "delete" {
		err := r.client.Z3.DeleteBucket(ctx, state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Deleting Z3 Bucket", err.Error())
			return
		}
	}
}

func (r *Z3BucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("name"), path.Root("name"), req, resp)
}

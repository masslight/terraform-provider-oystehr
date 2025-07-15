package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type Z3Bucket struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func convertZ3BucketToClientBucket(bucket Z3Bucket) client.Bucket {
	return client.Bucket{
		ID:   tfStringToStringPointer(bucket.ID),
		Name: tfStringToStringPointer(bucket.Name),
	}
}

func convertClientBucketToZ3Bucket(clientBucket *client.Bucket) Z3Bucket {
	return Z3Bucket{
		ID:   stringPointerToTfString(clientBucket.ID),
		Name: stringPointerToTfString(clientBucket.Name),
	}
}

var _ resource.Resource = &Z3BucketResource{}
var _ resource.ResourceWithConfigure = &Z3BucketResource{}

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
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Z3 bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

	retZ3Bucket := convertClientBucketToZ3Bucket(createdBucket)
	diags = resp.State.Set(ctx, retZ3Bucket)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Z3BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Z3Bucket
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientBucket, err := r.client.Z3.GetBucket(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Z3 Bucket", err.Error())
		return
	}

	retZ3Bucket := convertClientBucketToZ3Bucket(clientBucket)
	diags = resp.State.Set(ctx, retZ3Bucket)
	resp.Diagnostics.Append(diags...)
}

func (r *Z3BucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Updating Z3 buckets is not supported. Please recreate the resource with the new configuration.",
	)
}

func (r *Z3BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Z3Bucket
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Z3.DeleteBucket(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Z3 Bucket", err.Error())
		return
	}
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type ZambdaResource struct {
	client *client.Client
}

func NewZambdaResource() resource.Resource {
	return &ZambdaResource{}
}

func (r *ZambdaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_zambda"
}

func (r *ZambdaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Zambda function.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Zambda function.",
			},
			"runtime": schema.StringAttribute{
				Optional:    true,
				Description: "The runtime of the Zambda function.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the Zambda function.",
			},
			"trigger_method": schema.StringAttribute{
				Required:    true,
				Description: "The trigger method for the Zambda function.",
			},
			"schedule": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The schedule for the Zambda function.",
				Attributes: map[string]schema.Attribute{
					"expression": schema.StringAttribute{
						Required:    true,
						Description: "The cron expression for the schedule.",
					},
					"start": schema.StringAttribute{
						Optional:    true,
						Description: "The start time for the schedule.",
					},
					"end": schema.StringAttribute{
						Optional:    true,
						Description: "The end time for the schedule.",
					},
					"retry_policy": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "The retry policy for the schedule.",
						Attributes: map[string]schema.Attribute{
							"maximum_event_age": schema.Int64Attribute{
								Optional:    true,
								Description: "The maximum event age in seconds.",
							},
							"maximum_retry": schema.Int64Attribute{
								Optional:    true,
								Description: "The maximum number of retries.",
							},
						},
					},
				},
			},
			"file_info": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Information about the uploaded file.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the uploaded file.",
					},
					"size": schema.Int64Attribute{
						Computed:    true,
						Description: "The size of the uploaded file.",
					},
					"last_modified": schema.StringAttribute{
						Computed:    true,
						Description: "The last modified time of the uploaded file.",
					},
				},
			},
		},
	}
}

func (r *ZambdaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ZambdaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan client.ZambdaFunction

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdZambda, err := r.client.Zambda.CreateZambda(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Zambda", err.Error())
		return
	}

	resp.State.Set(ctx, createdZambda)
}

func (r *ZambdaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state client.ZambdaFunction

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	zambda, err := r.client.Zambda.GetZambda(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Zambda", err.Error())
		return
	}

	resp.State.Set(ctx, zambda)
}

func (r *ZambdaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan client.ZambdaFunction

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedZambda, err := r.client.Zambda.UpdateZambda(ctx, plan.ID, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Zambda", err.Error())
		return
	}

	resp.State.Set(ctx, updatedZambda)
}

func (r *ZambdaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state client.ZambdaFunction

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Zambda.DeleteZambda(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Zambda", err.Error())
		return
	}
}

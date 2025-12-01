package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
	"github.com/masslight/terraform-provider-oystehr/internal/fs"
	"github.com/masslight/terraform-provider-oystehr/internal/retry"
)

type RetryPolicy struct {
	MaximumEventAge types.Int64 `tfsdk:"maximum_event_age"` // Maximum event age in seconds
	MaximumRetry    types.Int64 `tfsdk:"maximum_retry"`     // Maximum number of retries
}

type Schedule struct {
	Expression  types.String `tfsdk:"expression"` // Cron expression
	Start       types.String `tfsdk:"start"`      // Optional start time
	End         types.String `tfsdk:"end"`        // Optional end time
	RetryPolicy *RetryPolicy `tfsdk:"retry_policy"`
}

type FileInfo struct {
	Name         types.String `tfsdk:"name"`          // Name of the uploaded file
	Size         types.Int64  `tfsdk:"size"`          // Size of the uploaded file
	LastModified types.String `tfsdk:"last_modified"` // Last modified time of the uploaded file
}

type Zambda struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Runtime RuntimeValue `tfsdk:"runtime"`
	// Runtime       types.String `tfsdk:"runtime"`
	MemorySize    types.Int32  `tfsdk:"memory_size"`
	Timeout       types.Int32  `tfsdk:"timeout"`
	Status        types.String `tfsdk:"status"`
	TriggerMethod types.String `tfsdk:"trigger_method"`
	// Schedule      Schedule     `tfsdk:"schedule"`
	Schedule types.Object `tfsdk:"schedule"`
	// FileInfo      FileInfo     `tfsdk:"file_info"`
	FileInfo       types.Object `tfsdk:"file_info"`
	Source         types.String `tfsdk:"source"`          // Pre-bundled source code of the Zambda function
	SourceChecksum types.String `tfsdk:"source_checksum"` // Checksum of the Zambda source code
}

type ZambdaV0 = Zambda

func convertZambdaToClientZambda(ctx context.Context, zambda Zambda) client.ZambdaFunction {
	var schedule *client.ZambdaSchedule
	if !zambda.Schedule.IsNull() && !zambda.Schedule.IsUnknown() {
		var tfSchedule Schedule
		zambda.Schedule.As(ctx, &tfSchedule, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		var retryPolicy *client.RetryPolicy
		if tfSchedule.RetryPolicy != nil {
			retryPolicy = &client.RetryPolicy{
				MaximumEventAge: tfInt64ToInt64Pointer(tfSchedule.RetryPolicy.MaximumEventAge),
				MaximumRetry:    tfInt64ToInt64Pointer(tfSchedule.RetryPolicy.MaximumRetry),
			}
		}
		schedule = &client.ZambdaSchedule{
			Expression:  tfStringToStringPointer(tfSchedule.Expression),
			Start:       tfStringToStringPointer(tfSchedule.Start),
			End:         tfStringToStringPointer(tfSchedule.End),
			RetryPolicy: retryPolicy,
		}
	}
	return client.ZambdaFunction{
		ID:               tfStringToStringPointer(zambda.ID),
		Name:             tfStringToStringPointer(zambda.Name),
		Runtime:          (*client.Runtime)(zambda.Runtime.ValueStringPointer()),
		MemorySize:       tfInt32ToInt32Pointer(zambda.MemorySize),
		TimeoutInSeconds: tfInt32ToInt32Pointer(zambda.Timeout),
		Status:           tfStringToStringPointer(zambda.Status),
		TriggerMethod:    (*client.TriggerMethod)(zambda.TriggerMethod.ValueStringPointer()),
		Schedule:         schedule,
	}
}

func convertClientZambdaToZambda(ctx context.Context, clientZambda *client.ZambdaFunction, sourceChecksum string) Zambda {
	var fi types.Object
	if clientZambda.FileInfo == nil {
		fi = types.ObjectNull(
			map[string]attr.Type{
				"name":          types.StringType,
				"size":          types.Int64Type,
				"last_modified": types.StringType,
			},
		)
	} else {
		tfFileInfo := FileInfo{
			Name:         stringPointerToTfString(clientZambda.FileInfo.Name),
			Size:         int64PointerToTfInt64(clientZambda.FileInfo.Size),
			LastModified: stringPointerToTfString(clientZambda.FileInfo.LastModified),
		}
		fi, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"name":          types.StringType,
			"size":          types.Int64Type,
			"last_modified": types.StringType,
		}, tfFileInfo)
	}
	var schedule types.Object
	if clientZambda.Schedule == nil {
		schedule = types.ObjectNull(
			map[string]attr.Type{
				"expression": types.StringType,
				"start":      types.StringType,
				"end":        types.StringType,
				"retry_policy": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"maximum_event_age": types.Int64Type,
						"maximum_retry":     types.Int64Type,
					},
				},
			},
		)
	} else {
		var tfRetryPolicy *RetryPolicy
		if clientZambda.Schedule.RetryPolicy != nil {
			tfRetryPolicy = &RetryPolicy{
				MaximumEventAge: int64PointerToTfInt64(clientZambda.Schedule.RetryPolicy.MaximumEventAge),
				MaximumRetry:    int64PointerToTfInt64(clientZambda.Schedule.RetryPolicy.MaximumRetry),
			}
		}
		tfSchedule := Schedule{
			Expression:  stringPointerToTfString(clientZambda.Schedule.Expression),
			Start:       stringPointerToTfString(clientZambda.Schedule.Start),
			End:         stringPointerToTfString(clientZambda.Schedule.End),
			RetryPolicy: tfRetryPolicy,
		}
		schedule, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"expression": types.StringType,
			"start":      types.StringType,
			"end":        types.StringType,
			"retry_policy": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"maximum_event_age": types.Int64Type,
					"maximum_retry":     types.Int64Type,
				},
			},
		}, tfSchedule)
	}
	zambda := Zambda{
		ID:             stringPointerToTfString(clientZambda.ID),
		Name:           stringPointerToTfString(clientZambda.Name),
		Runtime:        RuntimeValue{basetypes.NewStringValue(string(*clientZambda.Runtime))},
		MemorySize:     int32PointerToTfInt32(clientZambda.MemorySize),
		Timeout:        int32PointerToTfInt32(clientZambda.TimeoutInSeconds),
		Status:         stringPointerToTfString(clientZambda.Status),
		TriggerMethod:  types.StringValue(string(*clientZambda.TriggerMethod)),
		Schedule:       schedule,
		FileInfo:       fi,
		SourceChecksum: types.StringValue(sourceChecksum),
	}
	return zambda
}

var _ resource.Resource = (*ZambdaResource)(nil)
var _ resource.ResourceWithConfigure = &ZambdaResource{}
var _ resource.ResourceWithModifyPlan = (*ZambdaResource)(nil)
var _ resource.ResourceWithIdentity = &ZambdaResource{}
var _ resource.ResourceWithImportState = &ZambdaResource{}
var _ resource.ResourceWithUpgradeState = &ZambdaResource{}

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
	resp.Schema = zambdaSchemaV1
}

func (*ZambdaResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = idIdentitySchema
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
	var plan Zambda
	var config Zambda

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zambda := convertZambdaToClientZambda(ctx, plan)

	createdZambda, err := r.client.Zambda.CreateZambda(ctx, &zambda)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Zambda", err.Error())
		return
	}

	if config.Source.ValueString() != "" {
		err = r.client.Zambda.UploadZambdaSource(ctx, *createdZambda.ID, config.Source.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Uploading Zambda Source", err.Error())
			// Roll back create
			err = r.client.Zambda.DeleteZambda(ctx, *createdZambda.ID)
			if err != nil {
				resp.Diagnostics.AddError("Error Rolling Back Zambda Creation", err.Error())
			}
			return
		}
	}

	retrievedZambda, err := r.getZambdaAfterMutation(ctx, &resp.Diagnostics, *createdZambda.ID, plan.SourceChecksum.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Retrieving Created Zambda", err.Error())
		return
	}
	if retrievedZambda == nil {
		// Error already added to diagnostics in getZambdaAfterMutation
		return
	}

	retZambda := convertClientZambdaToZambda(ctx, retrievedZambda, plan.SourceChecksum.ValueString())
	retIdentity := IDIdentityModel{
		ID: retZambda.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retZambda)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

func (r *ZambdaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Zambda
	var identity IDIdentityModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var id string
	if !identity.ID.IsNull() {
		id = identity.ID.ValueString()
	} else {
		id = state.ID.ValueString()
	}

	zambda, err := r.client.Zambda.GetZambda(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code: 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Zambda", err.Error())
		return
	}

	retZambda := convertClientZambdaToZambda(ctx, zambda, state.SourceChecksum.ValueString())
	retIdentity := IDIdentityModel{
		ID: retZambda.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, retZambda)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, retIdentity)...)
}

func (r *ZambdaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Zambda
	var config Zambda
	var state Zambda

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zambda := convertZambdaToClientZambda(ctx, plan)

	updatedZambda, err := r.client.Zambda.UpdateZambda(ctx, state.ID.ValueString(), &zambda)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Zambda", err.Error())
		return
	}

	if config.Source.ValueString() != "" {
		// Different checksum, upload new source and use calculated checksum
		if plan.SourceChecksum.ValueString() != state.SourceChecksum.ValueString() {
			err = r.client.Zambda.UploadZambdaSource(ctx, *updatedZambda.ID, config.Source.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error Uploading Zambda Source", err.Error())
				// Roll back update
				previousStateZambda := convertZambdaToClientZambda(ctx, state)
				_, err = r.client.Zambda.UpdateZambda(ctx, state.ID.ValueString(), &previousStateZambda)
				if err != nil {
					resp.Diagnostics.AddError("Error Rolling Back Zambda Update", err.Error())
				}
				return
			}
		}
	}

	retrievedZambda, err := r.getZambdaAfterMutation(ctx, &resp.Diagnostics, *updatedZambda.ID, plan.SourceChecksum.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Retrieving Created Zambda", err.Error())
		return
	}
	if retrievedZambda == nil {
		// Error already added to diagnostics in getZambdaAfterMutation
		return
	}

	retZambda := convertClientZambdaToZambda(ctx, retrievedZambda, plan.SourceChecksum.ValueString())

	resp.State.Set(ctx, retZambda)
}

func (r *ZambdaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Zambda

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Zambda.DeleteZambda(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Zambda", err.Error())
		return
	}
}

func (r *ZambdaResource) getZambdaAfterMutation(ctx context.Context, diags *diag.Diagnostics, id string, expectedChecksum string) (*client.ZambdaFunction, error) {
	return retry.RetryWithBackoff(ctx, func() (*client.ZambdaFunction, error) {
		retrievedZambda, err := r.client.Zambda.GetZambda(ctx, id)
		if err != nil {
			diags.AddError("Error Retrieving Created Zambda", err.Error())
			// Bail out of retries
			return nil, nil
		}

		// Validation error checks to trigger retry
		if retrievedZambda.Status == nil {
			tflog.Debug(ctx, "Zambda status is nil, retrying")
			return nil, fmt.Errorf("Zambda status is nil")
		}
		if *retrievedZambda.Status == "Failed" {
			diags.AddError("Error Creating Zambda", "Zambda creation failed with status 'Failed'")
			// Bail out of retries
			return nil, nil
		}
		if *retrievedZambda.Status != "Active" {
			tflog.Debug(ctx, fmt.Sprintf("Zambda status is %s, retrying", *retrievedZambda.Status))
			return nil, fmt.Errorf("Zambda status is %s, expecting Active", *retrievedZambda.Status)
		}

		if retrievedZambda.FileInfo == nil {
			tflog.Debug(ctx, "Zambda FileInfo is nil, retrying")
			return nil, fmt.Errorf("Zambda FileInfo is nil")
		}
		if retrievedZambda.FileInfo.Checksum != nil && *retrievedZambda.FileInfo.Checksum != expectedChecksum {
			tflog.Debug(ctx, fmt.Sprintf("Zambda checksum is %s, expecting %s, retrying", *retrievedZambda.FileInfo.Checksum, expectedChecksum))
			return nil, fmt.Errorf("Zambda checksum is %s, expecting %s", *retrievedZambda.FileInfo.Checksum, expectedChecksum)
		}

		return retrievedZambda, nil
	}, retry.RetryConfig{
		BaseBackoff: retry.BaseBackoffDefault,
		MaxBackoff:  16000,
		MaxAttempts: 10,
	})
}

func (r *ZambdaResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan Zambda
	var config Zambda
	var state Zambda

	if req.Plan.Raw.IsNull() {
		// If the plan is null, we cannot modify it, so we return early.
		resp.Plan = req.Plan
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Source.ValueString() != "" {
		sourceChecksum, err := fs.Sha256HashFile(config.Source.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Calculating Source Checksum", err.Error())
			return
		}
		if sourceChecksum != state.SourceChecksum.ValueString() {
			plan.SourceChecksum = types.StringValue(sourceChecksum)
			plan.FileInfo = types.ObjectUnknown(map[string]attr.Type{
				"name":          types.StringType,
				"size":          types.Int64Type,
				"last_modified": types.StringType,
			})
		} else {
			plan.SourceChecksum = state.SourceChecksum
			plan.FileInfo = state.FileInfo
		}
	} else {
		plan.SourceChecksum = state.SourceChecksum
		plan.FileInfo = state.FileInfo
	}

	resp.Plan.Set(ctx, &plan)
}

func (r *ZambdaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *ZambdaResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// From version 0 to version 1
		0: resource.StateUpgrader{
			PriorSchema: &zambdaSchemaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState ZambdaV0
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}
				oldState.Source = types.StringNull()
				resp.Diagnostics.Append(resp.State.Set(ctx, &oldState)...)
			},
		},
	}
}

var (
	zambdaSchemaV0 = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Zambda function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Zambda function.",
			},
			"runtime": schema.StringAttribute{
				Required:    true,
				Description: "The runtime of the Zambda function.",
				CustomType:  RuntimeType{},
			},
			"memory_size": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The memory size allocated for the Zambda function in MB.",
				Default:     int32default.StaticInt32(1024),
			},
			"timeout": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The timeout for the Zambda function in seconds.",
				Default:     int32default.StaticInt32(27),
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the Zambda function.",
				// TODO? enum values
			},
			"trigger_method": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The trigger method for the Zambda function.",
				// TODO enum values
				Default: stringdefault.StaticString("http_auth"),
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
						Computed:    true,
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
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"maximum_event_age": types.Int64Type,
									"maximum_retry":     types.Int64Type,
								},
								map[string]attr.Value{
									"maximum_event_age": types.Int64Value(90),
									"maximum_retry":     types.Int64Value(0),
								}),
						),
					},
				},
			},
			"source": schema.StringAttribute{
				Optional:    true,
				Description: "The pre-bundled source code of the Zambda function.",
			},
			"source_checksum": schema.StringAttribute{
				Computed:    true,
				Description: "The checksum of the Zambda source code.",
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
		Version: 0,
	}
	zambdaSchemaV1 = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Zambda function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Zambda function.",
			},
			"runtime": schema.StringAttribute{
				Required:    true,
				Description: "The runtime of the Zambda function.",
				CustomType:  RuntimeType{},
			},
			"memory_size": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The memory size allocated for the Zambda function in MB.",
				Default:     int32default.StaticInt32(1024),
			},
			"timeout": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The timeout for the Zambda function in seconds.",
				Default:     int32default.StaticInt32(27),
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the Zambda function.",
				// TODO? enum values
			},
			"trigger_method": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The trigger method for the Zambda function.",
				// TODO enum values
				Default: stringdefault.StaticString("http_auth"),
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
						Computed:    true,
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
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"maximum_event_age": types.Int64Type,
									"maximum_retry":     types.Int64Type,
								},
								map[string]attr.Value{
									"maximum_event_age": types.Int64Value(90),
									"maximum_retry":     types.Int64Value(0),
								}),
						),
					},
				},
			},
			"source": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "The pre-bundled source code of the Zambda function.",
			},
			"source_checksum": schema.StringAttribute{
				Computed:    true,
				Description: "The checksum of the Zambda source code.",
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
		Version: 1,
	}
)

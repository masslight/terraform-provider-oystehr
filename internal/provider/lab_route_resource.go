package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type LabRoute struct {
	ID            types.String `tfsdk:"id"`
	LabID         types.String `tfsdk:"lab_id"`
	AccountNumber types.String `tfsdk:"account_number"`
}

type LabRouteIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func convertLabRouteToClientLabRoute(ctx context.Context, labRoute LabRoute) client.LabRoute {
	return client.LabRoute{
		RouteGUID:     tfStringToStringPointer(labRoute.ID),
		LabGUID:       tfStringToStringPointer(labRoute.LabID),
		AccountNumber: tfStringToStringPointer(labRoute.AccountNumber),
	}
}

func convertClientLabRouteToLabRoute(ctx context.Context, clientLabRoute *client.LabRoute) LabRoute {
	return LabRoute{
		ID:            stringPointerToTfString(clientLabRoute.RouteGUID),
		LabID:         stringPointerToTfString(clientLabRoute.LabGUID),
		AccountNumber: stringPointerToTfString(clientLabRoute.AccountNumber),
	}
}

var _ resource.Resource = &LabRouteResource{}
var _ resource.ResourceWithConfigure = &LabRouteResource{}
var _ resource.ResourceWithIdentity = &LabRouteResource{}
var _ resource.ResourceWithImportState = &LabRouteResource{}

type LabRouteResource struct {
	client *client.Client
}

func NewLabRouteResource() resource.Resource {
	return &LabRouteResource{}
}

func (r *LabRouteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_lab_route"
}

func (r *LabRouteResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lab_id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the lab.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_number": schema.StringAttribute{
				Required:    true,
				Description: "The account number associated with the lab route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *LabRouteResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				Description:       "The unique identifier for the route.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *LabRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LabRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LabRoute
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientLabRoute := convertLabRouteToClientLabRoute(ctx, plan)

	createdRoute, err := r.client.Lab.CreateLabRoute(ctx, &clientLabRoute)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Lab Route", err.Error())
		return
	}

	ret := convertClientLabRouteToLabRoute(ctx, createdRoute)
	identity := LabRouteIdentityModel{
		ID: ret.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &ret)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *LabRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LabRoute
	var identity LabRouteIdentityModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var routeID string
	if !identity.ID.IsNull() {
		routeID = identity.ID.ValueString()
	} else {
		routeID = state.ID.ValueString()
	}

	route, err := r.client.Lab.GetLabRoute(ctx, routeID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Lab Route", err.Error())
		return
	}

	ret := convertClientLabRouteToLabRoute(ctx, route)
	retIdentity := LabRouteIdentityModel{
		ID: ret.ID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &ret)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &retIdentity)...)
}

func (r *LabRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Updating Lab Routes is not supported. Please delete and recreate the resource.",
	)
}

func (r *LabRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LabRoute
	var identity LabRouteIdentityModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var routeID string
	if !identity.ID.IsNull() {
		routeID = identity.ID.ValueString()
	} else {
		routeID = state.ID.ValueString()
	}

	labRoute := convertLabRouteToClientLabRoute(ctx, state)
	err := r.client.Lab.DeleteLabRoute(ctx, routeID, &labRoute)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Lab Route", err.Error())
		return
	}
}

func (r *LabRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

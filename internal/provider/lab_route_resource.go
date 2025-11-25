package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type LabRouteAddress struct {
	Address1          types.String `tfsdk:"address1"`
	Address2          types.String `tfsdk:"address2"`
	City              types.String `tfsdk:"city"`
	StateProvinceCode types.String `tfsdk:"state_province_code"`
	PostalCode        types.String `tfsdk:"postal_code"`
}

type LabRoute struct {
	ID                        types.String `tfsdk:"id"`
	AccountNumber             types.String `tfsdk:"account_number"`
	LabID                     types.String `tfsdk:"lab_id"`
	LabName                   types.String `tfsdk:"lab_name"`
	PrimaryID                 types.String `tfsdk:"primary_id"`
	PrimaryName               types.String `tfsdk:"primary_name"`
	PrimaryAddress            types.Object `tfsdk:"primary_address"`
	ClientSiteID              types.String `tfsdk:"client_site_id"`
	EULAVersion               types.String `tfsdk:"eula_version"`
	EULAAccepterFullName      types.String `tfsdk:"eula_accepter_full_name"`
	EULAAcceptanceDateTimeUTC types.String `tfsdk:"eula_acceptance_date_time_utc"`
}

type LabRouteIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

var labRouteAddressAttributeTypes map[string]attr.Type = map[string]attr.Type{
	"address1":            types.StringType,
	"address2":            types.StringType,
	"city":                types.StringType,
	"state_province_code": types.StringType,
	"postal_code":         types.StringType,
}

func convertLabRouteToClientLabRoute(ctx context.Context, labRoute LabRoute) client.LabRoute {
	var primaryAddress *client.LabRouteAddress
	if !labRoute.PrimaryAddress.IsNull() {
		labRoute.PrimaryAddress.As(ctx, &primaryAddress, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	}
	return client.LabRoute{
		RouteGUID:                 tfStringToStringPointer(labRoute.ID),
		AccountNumber:             tfStringToStringPointer(labRoute.AccountNumber),
		LabGUID:                   tfStringToStringPointer(labRoute.LabID),
		LabName:                   tfStringToStringPointer(labRoute.LabName),
		PrimaryID:                 tfStringToStringPointer(labRoute.PrimaryID),
		PrimaryName:               tfStringToStringPointer(labRoute.PrimaryName),
		PrimaryAddress:            primaryAddress,
		ClientSiteID:              tfStringToStringPointer(labRoute.ClientSiteID),
		EULAVersion:               tfStringToStringPointer(labRoute.EULAVersion),
		EULAAccepterFullName:      tfStringToStringPointer(labRoute.EULAAccepterFullName),
		EULAAcceptanceDateTimeUTC: tfStringToStringPointer(labRoute.EULAAcceptanceDateTimeUTC),
	}
}

func convertClientLabRouteToLabRoute(ctx context.Context, clientLabRoute *client.LabRoute) LabRoute {
	var primaryAddress types.Object
	if clientLabRoute.PrimaryAddress == nil {
		primaryAddress = types.ObjectNull(labRouteAddressAttributeTypes)
	} else {
		primaryAddress, _ = types.ObjectValueFrom(ctx, labRouteAddressAttributeTypes, clientLabRoute.PrimaryAddress)
	}
	return LabRoute{
		ID:                        stringPointerToTfString(clientLabRoute.RouteGUID),
		AccountNumber:             stringPointerToTfString(clientLabRoute.AccountNumber),
		LabID:                     stringPointerToTfString(clientLabRoute.LabGUID),
		LabName:                   stringPointerToTfString(clientLabRoute.LabName),
		PrimaryID:                 stringPointerToTfString(clientLabRoute.PrimaryID),
		PrimaryName:               stringPointerToTfString(clientLabRoute.PrimaryName),
		PrimaryAddress:            primaryAddress,
		ClientSiteID:              stringPointerToTfString(clientLabRoute.ClientSiteID),
		EULAVersion:               stringPointerToTfString(clientLabRoute.EULAVersion),
		EULAAccepterFullName:      stringPointerToTfString(clientLabRoute.EULAAccepterFullName),
		EULAAcceptanceDateTimeUTC: stringPointerToTfString(clientLabRoute.EULAAcceptanceDateTimeUTC),
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
			"account_number": schema.StringAttribute{
				Required:    true,
				Description: "The account number associated with the lab route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lab_id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the lab.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lab_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the lab.",
			},
			"primary_id": schema.StringAttribute{
				Computed:    true,
				Description: "The primary identifier for the lab route.",
			},
			"primary_name": schema.StringAttribute{
				Optional:    true,
				Description: "The primary name for the lab route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"primary_address": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The primary address for the lab route.",
				Attributes: map[string]schema.Attribute{
					"address1": schema.StringAttribute{
						Required:    true,
						Description: "The first line of the address.",
					},
					"address2": schema.StringAttribute{
						Optional:    true,
						Description: "The second line of the address.",
					},
					"city": schema.StringAttribute{
						Required:    true,
						Description: "The city of the address.",
					},
					"state_province_code": schema.StringAttribute{
						Required:    true,
						Description: "The state or province code of the address.",
					},
					"postal_code": schema.StringAttribute{
						Required:    true,
						Description: "The postal code of the address.",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"client_site_id": schema.StringAttribute{
				Optional:    true,
				Description: "The client site identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"eula_version": schema.StringAttribute{
				Optional:    true,
				Description: "The version of the End User License Agreement (EULA) accepted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"eula_accepter_full_name": schema.StringAttribute{
				Optional:    true,
				Description: "The full name of the person who accepted the EULA.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"eula_acceptance_date_time_utc": schema.StringAttribute{
				Optional:    true,
				Description: "The UTC date and time when the EULA was accepted.",
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

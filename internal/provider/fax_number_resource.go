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
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

type FaxNumber struct {
	Number types.String `tfsdk:"number"`
}

type FaxNumberIdentityModel struct {
	Number types.String `tfsdk:"number"`
}

func convertClientFaxNumberToFaxNumber(ctx context.Context, clientFaxNumber *client.FaxNumber) FaxNumber {
	return FaxNumber{
		Number: stringPointerToTfString(clientFaxNumber.Number),
	}
}

var _ resource.Resource = &FaxNumberResource{}
var _ resource.ResourceWithConfigure = &FaxNumberResource{}
var _ resource.ResourceWithIdentity = &FaxNumberResource{}
var _ resource.ResourceWithImportState = &FaxNumberResource{}

type FaxNumberResource struct {
	client *client.Client
}

func NewFaxNumberResource() resource.Resource {
	return &FaxNumberResource{}
}

func (r *FaxNumberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oystehr_fax_number"
}

func (r *FaxNumberResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"number": schema.StringAttribute{
				Computed:    true,
				Description: "The project's configured fax number",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *FaxNumberResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"number": identityschema.StringAttribute{
				Description:       "The project's configured fax number.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *FaxNumberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FaxNumberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FaxNumber
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdFaxNumber, err := r.client.Fax.Onboard(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Fax Number", err.Error())
		return
	}

	ret := convertClientFaxNumberToFaxNumber(ctx, createdFaxNumber)
	identity := FaxNumberIdentityModel{
		Number: ret.Number,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &ret)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identity)...)
}

func (r *FaxNumberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FaxNumber
	var identity FaxNumberIdentityModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var faxNumber string
	if !identity.Number.IsNull() {
		faxNumber = identity.Number.ValueString()
	} else {
		faxNumber = state.Number.ValueString()
	}

	route, err := r.client.Fax.GetFaxNumber(ctx, faxNumber)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected status code: 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Fax Number", err.Error())
		return
	}

	ret := convertClientFaxNumberToFaxNumber(ctx, route)
	retIdentity := FaxNumberIdentityModel{
		Number: ret.Number,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &ret)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, &retIdentity)...)
}

func (r *FaxNumberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Updating Fax Numbers is not supported. Please delete and recreate the resource.",
	)
}

func (r *FaxNumberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FaxNumber
	var identity FaxNumberIdentityModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Fax.Offboard(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Fax Number", err.Error())
		return
	}
}

func (r *FaxNumberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("number"), path.Root("number"), req, resp)
}

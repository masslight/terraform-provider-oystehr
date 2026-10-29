package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLabRouteValidation_PrimaryPhoneFailsRegexp_ReturnsError(t *testing.T) {
	plan := LabRoute{
		AccountNumber: types.StringValue("12345"),
		LabID:         types.StringValue("12345"),
		LabName:       types.StringValue("Test"),
		PrimaryName:   types.StringValue("Primary"),
		PrimaryPhone:  types.StringValue("1111111111"),
	}
	route := convertLabRouteToClientLabRoute(context.Background(), plan)
	resp := &resource.CreateResponse{}
	if !validateLabRoute(route, resp) {
		t.Fatalf("expected validation error but got none")
	}
}

func TestLabRouteValidation_PrimaryPhonePassesRegexp_NoError(t *testing.T) {
	plan := LabRoute{
		AccountNumber: types.StringValue("12345"),
		LabID:         types.StringValue("12345"),
		LabName:       types.StringValue("Test"),
		PrimaryName:   types.StringValue("Primary"),
		PrimaryPhone:  types.StringValue("(111) 111-1111"),
	}
	route := convertLabRouteToClientLabRoute(context.Background(), plan)
	resp := &resource.CreateResponse{}
	if validateLabRoute(route, resp) {
		t.Fatalf("expected no validation error but got one")
	}
}

func TestLabRouteValidation_PrimaryAddressLine2Empty_ReturnsError(t *testing.T) {
	address := LabRouteAddress{
		Address1:          types.StringValue("10 Main St"),
		Address2:          types.StringValue(""),
		City:              types.StringValue("Manhattan"),
		StateProvinceCode: types.StringValue("KS"),
		PostalCode:        types.StringValue("12345"),
	}
	tfAddress, _ := types.ObjectValueFrom(context.Background(), labRouteAddressAttributeTypes, address)
	plan := LabRoute{
		AccountNumber:  types.StringValue("12345"),
		LabID:          types.StringValue("12345"),
		LabName:        types.StringValue("Test"),
		PrimaryName:    types.StringValue("Primary"),
		PrimaryPhone:   types.StringValue("(111) 111-1111"),
		PrimaryAddress: tfAddress,
	}
	route := convertLabRouteToClientLabRoute(context.Background(), plan)
	resp := &resource.CreateResponse{}
	if !validateLabRoute(route, resp) {
		t.Fatalf("expected validation error but got none")
	}
}

func TestLabRouteValidation_PrimaryAddressLine2Null_NoError(t *testing.T) {
	address := LabRouteAddress{
		Address1:          types.StringValue("10 Main St"),
		Address2:          types.StringNull(),
		City:              types.StringValue("Manhattan"),
		StateProvinceCode: types.StringValue("KS"),
		PostalCode:        types.StringValue("12345"),
	}
	tfAddress, _ := types.ObjectValueFrom(context.Background(), labRouteAddressAttributeTypes, address)
	plan := LabRoute{
		AccountNumber:  types.StringValue("12345"),
		LabID:          types.StringValue("12345"),
		LabName:        types.StringValue("Test"),
		PrimaryName:    types.StringValue("Primary"),
		PrimaryPhone:   types.StringValue("(111) 111-1111"),
		PrimaryAddress: tfAddress,
	}
	route := convertLabRouteToClientLabRoute(context.Background(), plan)
	resp := &resource.CreateResponse{}
	if validateLabRoute(route, resp) {
		t.Fatalf("expected no validation error but got one")
	}
}

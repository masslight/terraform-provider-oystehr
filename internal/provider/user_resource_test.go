package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/masslight/terraform-provider-oystehr/internal/client"
)

func TestConvertUserToClientInvite_BuildsResourceAndFields(t *testing.T) {
	plan := User{
		Email:         types.StringValue("demo@example.com"),
		Username:      types.StringValue("demo"),
		ApplicationID: types.StringValue("app-1"),
		ResourceType:  types.StringValue("Practitioner"),
		Profile:       types.StringValue("Practitioner/1"),
		Roles:         convertStringSliceToList(context.Background(), []string{"role-1", "role-2"}),
		Password:      types.StringValue("s3cret!"),
	}

	invite := convertUserToClientInvite(plan)

	if invite.Email == nil || *invite.Email != "demo@example.com" {
		t.Fatalf("expected email demo@example.com, got %v", invite.Email)
	}
	if invite.ApplicationID == nil || *invite.ApplicationID != "app-1" {
		t.Fatalf("expected applicationId app-1, got %v", invite.ApplicationID)
	}
	if invite.Resource == nil || invite.Resource["resourceType"] != "Practitioner" {
		t.Fatalf("expected resource with resourceType Practitioner, got %+v", invite.Resource)
	}
	if len(invite.Roles) != 2 || invite.Roles[0] != "role-1" {
		t.Fatalf("expected roles [role-1 role-2], got %+v", invite.Roles)
	}
}

func TestConvertUserToClientInvite_OmitsResourceWhenTypeNull(t *testing.T) {
	plan := User{
		Email:         types.StringValue("demo@example.com"),
		ApplicationID: types.StringValue("app-1"),
		ResourceType:  types.StringNull(),
	}

	invite := convertUserToClientInvite(plan)

	if invite.Resource != nil {
		t.Fatalf("expected no resource when resource_type is null, got %+v", invite.Resource)
	}
}

func TestShouldRotatePassword(t *testing.T) {
	cases := []struct {
		name         string
		stateVersion types.Int64
		planVersion  types.Int64
		want         bool
	}{
		{"bumped", types.Int64Value(0), types.Int64Value(1), true},
		{"unchanged", types.Int64Value(1), types.Int64Value(1), false},
		{"decreased", types.Int64Value(2), types.Int64Value(1), false},
		{"null state", types.Int64Null(), types.Int64Value(1), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := User{PasswordVersion: tc.stateVersion}
			plan := User{PasswordVersion: tc.planVersion}
			if got := shouldRotatePassword(state, plan); got != tc.want {
				t.Fatalf("shouldRotatePassword() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsUserNotFoundError(t *testing.T) {
	if !isUserNotFoundError(errors.New("failed to get user: unexpected status code: 404, response body: not found")) {
		t.Fatalf("expected 404 error to be detected as not found")
	}
	if isUserNotFoundError(errors.New("unexpected status code: 500")) {
		t.Fatalf("expected 500 error not to be detected as not found")
	}
	if isUserNotFoundError(nil) {
		t.Fatalf("expected nil error not to be detected as not found")
	}
}

// TestMapUserToState_NeverSourcesPassword asserts that the password stored in
// state always comes from config and is never populated from the API response.
func TestMapUserToState_NeverSourcesPassword(t *testing.T) {
	state := User{
		ID:            types.StringValue("user-123"),
		Email:         types.StringValue("old@example.com"),
		ApplicationID: types.StringValue("app-1"),
		Password:      types.StringValue("config-password"),
	}
	apiUser := &client.User{
		ID:    strPtr("user-123"),
		Email: strPtr("new@example.com"),
	}

	got := mapUserToState(state, apiUser)

	if got.Password.ValueString() != "config-password" {
		t.Fatalf("expected password to be preserved from config, got %q", got.Password.ValueString())
	}
	if got.Email.ValueString() != "new@example.com" {
		t.Fatalf("expected email to be refreshed from API, got %q", got.Email.ValueString())
	}
	if got.ApplicationID.ValueString() != "app-1" {
		t.Fatalf("expected application_id to be preserved from state, got %q", got.ApplicationID.ValueString())
	}
}

func strPtr(s string) *string {
	return &s
}

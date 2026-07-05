package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type User struct {
	ID              types.String `tfsdk:"id"`
	Email           types.String `tfsdk:"email"`
	Username        types.String `tfsdk:"username"`
	ApplicationID   types.String `tfsdk:"application_id"`
	ResourceType    types.String `tfsdk:"resource_type"`
	Profile         types.String `tfsdk:"profile"`
	Roles           types.List   `tfsdk:"roles"`
	Password        types.String `tfsdk:"password"`
	PasswordVersion types.Int64  `tfsdk:"password_version"`
}

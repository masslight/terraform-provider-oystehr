variable "bootstrap_user_password" {
  type        = string
  description = "Password for the Terraform-managed bootstrap user. Provide via a tfvars file or environment variable; do not hardcode."
  sensitive   = true
}

# Bootstrap user for a lower environment (e.g. dev/staging).
# The provider invites the user and sets the password on create, and deletes
# the user on destroy. There is generally no good production use case for
# Terraform-managed users.
resource "oystehr_user" "bootstrap" {
  email          = "demo@ottehr.com"
  application_id = "00000000-0000-0000-0000-000000000000" # Replace with your application ID
  resource_type  = "Practitioner"
  roles          = [] # Optionally assign role IDs

  password = var.bootstrap_user_password

  # Increment to re-apply the password on a subsequent apply.
  password_version = 1
}

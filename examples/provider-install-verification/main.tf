terraform {
  required_providers {
    oystehr = {
      source = "registry.terraform.io/masslight/oystehr"
    }
  }
}

provider "oystehr" {
  project_id = "" # Replace with your Oystehr project ID

  # Replace with your Oystehr client ID and secret.
  # You can use either a project-level or developer-level M2M client ID and secret.
  client_id     = ""
  client_secret = ""
}

# Create an application
resource "oystehr_application" "example" {
  name = "example"
}

# Import application by ID
resource "oystehr_application" "example2" {
  name        = "example"
  description = "example"
  allowed_callback_urls = [
    "https://example.com/callback",
    "https://example.com/another-callback"
  ]
}

import {
  to = oystehr_application.example2
  id = "58d80b94-6781-4782-bce0-9740ed6c29b7"
}

# Create a role
resource "oystehr_role" "example" {
  name          = "example2"
  description   = "example"
  access_policy = { "rule" : [] }
}

# Import role by ID
resource "oystehr_role" "example2" {
  name          = "example"
  description   = "example"
  access_policy = { rule : [] }
}

import {
  to = oystehr_role.example2
  id = "349b651d-f490-4926-953b-1bbfd5afc2f3"
}

# Create an M2M client
resource "oystehr_m2m" "example" {
  name        = "example"
  description = "example"
}

# Import M2M client using identity (requires Terraform 1.12+)
resource "oystehr_m2m" "example2" {
  name        = "example2"
  description = "example"
}

import {
  to = oystehr_m2m.example2
  identity = {
    id = "13714262-d261-44e3-80e4-c97e6981fdf2"
  }
}

# Create a secret
resource "oystehr_secret" "super_secret" {
  name  = "super_secret"
  value = "updated!! super_secret_value"
}

# Import secret by ID
resource "oystehr_secret" "in_person_previsit_questionnaire" {
  name  = "IN_PERSON_PREVISIT_QUESTIONNAIRE"
  value = "https://ottehr.com/FHIR/Questionnaire/intake-paperwork-inperson|1.0.3"
}

import {
  to = oystehr_secret.in_person_previsit_questionnaire
  id = "IN_PERSON_PREVISIT_QUESTIONNAIRE"
}

# Create a Zambda (serverless function)
resource "oystehr_zambda" "example" {
  name           = "example"
  runtime        = "nodejs20.x"
  trigger_method = "http_auth"
  memory_size    = 128
  timeout        = 20
  source         = "lorem-ipsum.zip"
}

# Import Zambda by ID
resource "oystehr_zambda" "example2" {
  name           = "test"
  trigger_method = "http_auth"
  runtime        = "python3.13"
  memory_size    = 1024
  timeout        = 27
}

import {
  to = oystehr_zambda.example2
  id = "3d078f46-32c8-4df2-8ee8-6893a677798d"
}

# Create a Patient FHIR resource
resource "oystehr_fhir_resource" "example" {
  type = "Patient"
  data = jsonencode({
    active = true
    name = [
      {
        use    = "official"
        family = "Doe"
        given  = ["John"]
      }
    ]
  })
  managed_fields = ["maritalStatus", "link", "name"]
}

# Import Patient FHIR resource using identity (requires Terraform 1.12+)
resource "oystehr_fhir_resource" "example2" {
  type = "Patient"
  data = "{}"
}

import {
  to = oystehr_fhir_resource.example2
  identity = {
    id   = "20b4fcd0-05e7-4763-88a3-68a28db505fa"
    type = "Patient"
  }
}

# Import Patient FHIR resource by ID
resource "oystehr_fhir_resource" "example3" {
  type = "Patient"
  data = "{}"
}

import {
  to = oystehr_fhir_resource.example3
  # Use `ResourceType/ResourceId` format to import FHIR resources by ID
  id = "Patient/7f2c28f8-cfb8-4002-8544-e80dd58d8a61"
}

resource "oystehr_fhir_resource" "example4" {
  type = "Patient"
  data = jsonencode({
    active = true
    name = [
      {
        use    = "official"
        family = "Doe"
        given  = ["Janey"]
      }
    ]
    meta = {
      tag = [
        {
          system  = "http://example.com/fhir/tag"
          code    = "example-"
          display = "Example Tag"
        }
      ]
    }
  })
  managed_fields = ["name", "meta"]
}

# Create a Z3 bucket
resource "oystehr_z3_bucket" "example" {
  name = "example_bucket"
}

# Import Z3 bucket using identity (requires Terraform 1.12+)
resource "oystehr_z3_bucket" "example2" {
  name           = "f6f98331-4079-465d-84ec-ef0f3a839a77-school-work-note-templates"
  removal_policy = "retain"
}

import {
  to = oystehr_z3_bucket.example2
  identity = {
    name = "f6f98331-4079-465d-84ec-ef0f3a839a77-school-work-note-templates"
  }
}

# Import Z3 bucket by ID
resource "oystehr_z3_bucket" "example3" {
  name = "f6f98331-4079-465d-84ec-ef0f3a839a77-patient-photos"
}

import {
  to = oystehr_z3_bucket.example3
  id = "f6f98331-4079-465d-84ec-ef0f3a839a77-patient-photos"
}

# Create Z3 objects (files) in the bucket
resource "oystehr_z3_object" "example" {
  bucket = oystehr_z3_bucket.example.name
  key    = "some/path/to/example_object"
  source = "lorem-ipsum.zip"
}

# Import Z3 object using identity (requires Terraform 1.12+)
resource "oystehr_z3_object" "example2" {
  bucket = oystehr_z3_bucket.example.name
  key    = "lorem-ipsum.zip"
  source = "lorem-ipsum.zip"
}

import {
  to = oystehr_z3_object.example2
  identity = {
    bucket = "example_bucket"
    key    = "lorem-ipsum.zip"
  }
}

# Import Z3 object by ID
resource "oystehr_z3_object" "example3" {
  bucket = oystehr_z3_bucket.example.name
  key    = "test/lorem-ipsum.zip"
  source = "lorem-ipsum.zip"
}

import {
  to = oystehr_z3_object.example3
  # Use `BucketName/Key` format to import Z3 objects by ID
  id = "example_bucket/test/lorem-ipsum.zip"
}

resource "oystehr_lab_route" "example" {
  account_number = "284c0181-b84f-4637-8585-c3d04298afcc"
  lab_id         = "790b282d-77e9-4697-9f59-0cef8238033a"
}

resource "oystehr_lab_route" "example2" {
  account_number = "bc94b94c-3d2e-4a77-92a3-4570b834eae5"
  lab_id         = "790b282d-77e9-4697-9f59-0cef8238033a"
}

import {
  to = oystehr_lab_route.example2
  id = "6607e9c4-4258-4fcc-8454-76bcf29ff684"
}

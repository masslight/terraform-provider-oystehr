terraform {
  required_providers {
    oystehr = {
      source = "registry.terraform.io/masslight/oystehr"
    }
  }
}

provider "oystehr" {

}

resource "oystehr_fhir_resource" "example" {
  type = "Patient"
  data = {
    resourceType = "Patient"
    id           = "example"
    active       = true
    name = [
      {
        use    = "official"
        family = "Doe"
        given  = ["John"]
      }
    ]
  }
}

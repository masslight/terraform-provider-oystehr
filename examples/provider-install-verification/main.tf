terraform {
  required_providers {
    oystehr = {
      source = "registry.terraform.io/masslight/oystehr"
    }
  }
}

provider "oystehr" {

}

data "oystehr_ds" "example" {}

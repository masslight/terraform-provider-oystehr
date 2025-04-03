package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/masslight/terraform-provider-oystehr/internal/provider"
)

var (
	version string = "dev"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/masslight/oystehr",
		Debug:   false,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

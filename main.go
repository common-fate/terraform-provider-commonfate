package main

import (
	"context"
	"flag"
	"log"

	commonfate "github.com/common-fate/terraform-provider-commonfate/internal"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	err := providerserver.Serve(context.Background(), commonfate.New, providerserver.ServeOpts{
		Debug:   debug,
		Address: "registry.terraform.io/common-fate/commonfate",
	})
	if err != nil {
		log.Fatal(err)
	}
}

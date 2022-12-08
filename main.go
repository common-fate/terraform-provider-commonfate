package main

import (
	"context"
	"flag"
	"log"

	"github.com/common-fate/common-fate-terraform-proto/commonfate"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	err := providerserver.Serve(context.Background(), commonfate.New, providerserver.ServeOpts{
		Debug:   debug,
		Address: "registry.terraform.io/example/commonfate",
	})

	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Println("Running provider")
	}
}

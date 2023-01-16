package main

import (
	"context"
	"flag"

	commonfate "github.com/common-fate/common-fate-terraform-proto/internal"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	providerserver.Serve(context.Background(), commonfate.New, providerserver.ServeOpts{
		Debug:   debug,
		Address: "commonfate.com/commonfate/commonfate",
	})

}

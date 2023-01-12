package main

import (
	"context"
	"flag"

	commonfate "github.com/common-fate/common-fate-terraform-proto/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	// fmt.Println("entering main")

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	providerserver.Serve(context.Background(), commonfate.New, providerserver.ServeOpts{
		Debug:   debug,
		Address: "commonfate.com/commonfate/commonfate",
	})

	// if err != nil {
	// 	log.Fatal(err.Error())
	// } else {
	// 	log.Println("Running provider")
	// }
}

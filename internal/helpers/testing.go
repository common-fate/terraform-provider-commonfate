package helpers

// const (
// 	// providerConfig is a shared configuration to combine with the actual
// 	// test configuration so the HashiCups client is properly configured.
// 	// It is also possible to use the HASHICUPS_ environment variables instead,
// 	// such as updating the Makefile and running the testing through that tool.
// 	ProviderConfig = `

// 	provider "commonfate" {
// 		api_url = "http://localhost:8080"
// 		authz_url = "http://localhost:5050"
// 		oidc_client_id     = ""
// 		oidc_issuer        = "https://cognito-idp.ap-southeast-2.amazonaws.com/ap-southeast-2_RyXr9bHis"
// 	  }

// `
// )

// var (
// 	// testAccProtoV6ProviderFactories are used to instantiate a provider during
// 	// acceptance testing. The factory function will be invoked for every Terraform
// 	// CLI command executed to create a provider server to which the CLI can
// 	// reattach.
// 	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
// 		"commonfate": providerserver.NewProtocol6WithError(internal.New()),
// 	}
// )
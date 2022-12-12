package commonfate

import (
	"context"
	"os"

	"github.com/common-fate/common-fate/governance"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider              = &commonfateProvider{}
	_ provider.ProviderWithMetadata  = &commonfateProvider{}
	_ provider.ProviderWithGetSchema = &commonfateProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &commonfateProvider{}
}

// commonfateProvider is the provider implementation.
type commonfateProvider struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Version  types.String `tfsdk:"version"`
}

// commonfateProviderModel maps provider schema data to a Go type.
type commonfateProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Version  types.String `tfsdk:"version"`
}

// Metadata returns the provider type name.
func (p *commonfateProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "commonfate"
}

// GetSchema defines the provider-level schema for configuration data.
func (p *commonfateProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				Type:     types.StringType,
				Optional: true,
			},
			"username": {
				Type:     types.StringType,
				Optional: true,
			},
			"password": {
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
			"version": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

// Configure prepares a the Common Fate API for data sources and resources.
func (p *commonfateProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config commonfateProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Common Fate API Host",
			"The provider cannot create the Common Fate API client as there is an unknown configuration value for the Common Fate API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the COMMONFATE_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Common Fate API Username",
			"The provider cannot create the Common Fate API client as there is an unknown configuration value for the Common Fate API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the COMMONFATE_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Common Fate API Password",
			"The provider cannot create the Common Fate API client as there is an unknown configuration value for the Common Fate API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the COMMONFATE_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("COMMONFATE_HOST")
	username := os.Getenv("COMMONFATE_USERNAME")
	password := os.Getenv("COMMONFATE_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Common Fate API Host",
			"The provider cannot create the Common Fate API client as there is a missing or empty value for the Common Fate API host. "+
				"Set the host value in the configuration or use the COMMONFATE_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Common Fate API Username",
			"The provider cannot create the Common Fate API client as there is a missing or empty value for the Common Fate API username. "+
				"Set the username value in the configuration or use the COMMONFATE_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Common Fate API Password",
			"The provider cannot create the Common Fate API client as there is a missing or empty value for the Common Fate API password. "+
				"Set the password value in the configuration or use the COMMONFATE_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := governance.NewClientWithResponses("http://localhost:8889")
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client"),
			"error setting up client",
			"error",
		)
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured commonfate client", map[string]any{"success": true})

}

// func withUserAgent(version string) func(ctx context.Context, req *http.Request) error {
// 	return func(ctx context.Context, req *http.Request) error {
// 		req.Header.Set("User-Agent", "terraform-provider-commonfate/"+version)
// 		return nil
// 	}
// }

// DataSources defines the data sources implemented in the provider.
func (p *commonfateProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *commonfateProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccessRuleResource,
	}
}

// With the resource.Resource implementation
func NewAccessRuleResource() resource.Resource {
	return &AccessRuleResource{}
}

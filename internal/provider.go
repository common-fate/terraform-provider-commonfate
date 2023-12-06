package internal

import (
	"context"
	"strings"

	config_client "github.com/common-fate/sdk/config"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
// _ provider.Provider              = &commonfateProvider{}
// // _ provider.ProviderWithMetadata  = &commonfateProvider{}
// _ provider.ProviderWithMetaSchema = &commonfateProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &CommonFateProvider{}
}

// commonfateProvider is the provider implementation.
type CommonFateProvider struct {
}

// commonfateProviderModel maps provider schema data to a Go type.
type CommonFateProviderModel struct {
	DeploymentAPIURL types.String `tfsdk:"deployment_api_url"`
	IssuerURL        types.String `tfsdk:"issuer_url"`
	OIDCClientId     types.String `tfsdk:"oidc_client_id"`
	OIDCClientSecret types.String `tfsdk:"oidc_client_secret"`
	OIDCIssuer       types.String `tfsdk:"oidc_issuer"`
}

// Metadata returns the provider type name.
func (p *CommonFateProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "commonfate"
}

// GetSchema defines the provider-level schema for configuration data.
func (p *CommonFateProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "",
		Attributes: map[string]schema.Attribute{
			"deployment_api_url": schema.StringAttribute{
				Description: "The API url of your Common Fate deployment.",
				Required:    true,
			},
			"issuer_url": schema.StringAttribute{
				Description: "The API  of your app client.",
				Required:    true,
			},
			"oidc_client_id": schema.StringAttribute{
				Required: true,
			},
			"oidc_client_secret": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"oidc_issuer": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

// Configure prepares a the Common Fate API for data sources and resources.
func (p *CommonFateProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config CommonFateProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//using context.Background() here causes a cancelled context issue
	//see https://github.com/databricks/databricks-sdk-go/issues/671
	cfg, err := config_client.NewServerContext(context.Background(), config_client.Opts{
		APIURL:       config.DeploymentAPIURL.ValueString(),
		AccessURL:    config.IssuerURL.ValueString(),
		ClientID:     config.OIDCClientId.ValueString(),
		ClientSecret: config.OIDCClientSecret.ValueString(),
		// @TODO consider changing this to use a direct issuer env var
		OIDCIssuer: strings.TrimSuffix(config.OIDCIssuer.ValueString(), "/"),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to load config",
			err.Error(),
		)

		return
	}

	// // Make the Common Fate client available during DataSource and Resource
	// // type Configure methods.
	resp.DataSourceData = cfg
	resp.ResourceData = cfg

	tflog.Info(ctx, "Configured Common Fate client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *CommonFateProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPagerDutyScheduleDataSource,
	}
}

func (p *CommonFateProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccessPolicyResource,
		NewScheduleResource,
		NewAWSJITPolicyResource,
		NewGCPJITPolicyResource,
	}
}

// With the resource.Resource implementation
func NewAccessPolicyResource() resource.Resource {
	return &PolicyResource{}
}

func NewAWSJITPolicyResource() resource.Resource {
	return &AWSJITPolicyResource{}
}
func NewGCPJITPolicyResource() resource.Resource {
	return &GCPJITPolicyResource{}
}
func NewScheduleResource() resource.Resource {
	return &ScheduleResource{}
}

func NewPagerDutyScheduleDataSource() datasource.DataSource {
	return &PagerdutyScheduleDataSource{}
}

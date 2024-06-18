package internal

import (
	"context"
	"strings"

	config_client "github.com/common-fate/sdk/config"
	"github.com/common-fate/terraform-provider-commonfate/internal/access"
	"github.com/common-fate/terraform-provider-commonfate/internal/auth0"
	"github.com/common-fate/terraform-provider-commonfate/internal/aws"
	"github.com/common-fate/terraform-provider-commonfate/internal/datastax"
	"github.com/common-fate/terraform-provider-commonfate/internal/entra"
	"github.com/common-fate/terraform-provider-commonfate/internal/gcp"
	"github.com/common-fate/terraform-provider-commonfate/internal/generic"
	"github.com/common-fate/terraform-provider-commonfate/internal/logs"
	"github.com/common-fate/terraform-provider-commonfate/internal/okta"
	"github.com/common-fate/terraform-provider-commonfate/internal/opsgenie"
	"github.com/common-fate/terraform-provider-commonfate/internal/pagerduty"
	"github.com/common-fate/terraform-provider-commonfate/internal/slack"
	"github.com/common-fate/terraform-provider-commonfate/internal/webhook"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &CommonFateProvider{}
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
	APIURL           types.String `tfsdk:"api_url"`
	AuthzURL         types.String `tfsdk:"authz_url"`
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
			"api_url": schema.StringAttribute{
				Description: "The API url of your Common Fate deployment.",
				Optional:    true,
			},
			"authz_url": schema.StringAttribute{
				Description: "The base URL of the Common Fate authz service. If not provided, will default to the same URL as the api_url",
				Optional:    true,
			},
			"oidc_client_id": schema.StringAttribute{
				Optional: true,
			},
			"oidc_client_secret": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"oidc_issuer": schema.StringAttribute{
				Optional: true,
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
	cfg, err := config_client.New(context.Background(), config_client.Opts{
		APIURL:        config.APIURL.ValueString(),
		ClientID:      config.OIDCClientId.ValueString(),
		ClientSecret:  config.OIDCClientSecret.ValueString(),
		OIDCIssuer:    strings.TrimSuffix(config.OIDCIssuer.ValueString(), "/"),
		AuthzURL:      config.AuthzURL.ValueString(),
		ConfigSources: []string{"env"},
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

	tflog.Debug(ctx, "Configured Common Fate client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *CommonFateProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *CommonFateProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPolicySetResource,
		NewAccessWorkflowResource,
		NewSelectorResource,
		NewGCPProjectSelectorResource,
		NewGCPProjectAvailabilitiesResource,
		NewGCPFolderSelectorResource,
		NewGCPFolderAvailabilitiesResource,
		NewGCPOrganizationSelectorResource,
		NewGCPOrganizationAvailabilitiesResource,
		NewAvailabilitySpecResource,
		NewGCPIntegrationResource,
		NewGCPRoleGroupResource,
		NewGCPRoleGroupFolderAvailabilitiesResource,
		NewGCPRoleGroupProjectAvailabilitiesResource,
		NewSlackAlertResource,
		NewWebhookProvisionerResource,
		NewAWSIDCIntegrationResource,
		NewAWSRDSDatabaseResource,
		NewSlackIntegrationResource,
		NewPagerDutyIntegrationResource,
		NewAWSAccountSelectorResource,
		NewAWSIDCAccountAvailabilitiesResource,
		NewOpsGenieIntegrationResource,
		NewEntraIntegrationResource,
		NewEntraGroupSelectorResource,
		NewEntraGroupAvailabilitiesResource,
		NewAWSRDSSelectorResource,
		NewAWSRDSAvailabilitiesResource,
		NewAWSIDCGroupSelectorResource,
		NewAWSIDCGroupAvailabilitiesResource,
		NewOktaIntegrationResource,
		NewOktaGroupSelectorResource,
		NewOktaGroupAvailabilitiesResource,
		NewDataStaxIntegrationResource,
		NewDataStaxOrganizationAvailabilitiesResource,
		NewDataStaxOrganizationSelectorResource,
		NewGCPBigQueryTableAvailabilitiesResource,
		NewGCPBigQueryTableSelectorResource,
		NewGCPBigQueryDatasetAvailabilitiesResource,
		NewGCPBigQueryDatasetSelectorResource,
		NewWebhookIntegrationResource,
		NewAuth0IntegrationResource,
		NewAuth0OrganizationSelectorResource,
		NewAuth0OrganizationAvailabilitiesResource,
		logs.NewS3LogDestinationResource,
	}
}

// With the resource.Resource implementation
func NewPolicySetResource() resource.Resource {
	return &access.PolicySetResource{}
}

func NewSelectorResource() resource.Resource {
	return &generic.SelectorResource{}
}

func NewGCPProjectSelectorResource() resource.Resource {
	return &gcp.GCPProjectSelectorResource{}
}

func NewGCPProjectAvailabilitiesResource() resource.Resource {
	return &gcp.GCPProjectAvailabilitiesResource{}
}

func NewGCPFolderSelectorResource() resource.Resource {
	return &gcp.GCPFolderSelectorResource{}
}

func NewGCPFolderAvailabilitiesResource() resource.Resource {
	return &gcp.GCPFolderAvailabilitiesResource{}
}

func NewGCPRoleGroupResource() resource.Resource {
	return &gcp.GCPRoleGroupResource{}
}

func NewGCPRoleGroupFolderAvailabilitiesResource() resource.Resource {
	return &gcp.GCPRoleGroupFolderAvailabilitiesResource{}
}

func NewGCPRoleGroupProjectAvailabilitiesResource() resource.Resource {
	return &gcp.GCPRoleGroupProjectAvailabilitiesResource{}
}

func NewGCPOrganizationSelectorResource() resource.Resource {
	return &gcp.GCPOrganizationSelectorResource{}
}

func NewGCPOrganizationAvailabilitiesResource() resource.Resource {
	return &gcp.GCPOrganizationAvailabilitiesResource{}
}

func NewWebhookProvisionerResource() resource.Resource {
	return &access.WebhookProvisionerResource{}
}

func NewAvailabilitySpecResource() resource.Resource {
	return &generic.AvailabilitySpecResource{}
}

func NewSlackAlertResource() resource.Resource {
	return &slack.SlackAlertResource{}
}

func NewAccessWorkflowResource() resource.Resource {
	return &access.AccessWorkflowResource{}
}

func NewGCPIntegrationResource() resource.Resource {
	return &gcp.GCPIntegrationResource{}
}

func NewAWSIDCIntegrationResource() resource.Resource {
	return &aws.AWSIDCIntegrationResource{}
}
func NewAWSRDSDatabaseResource() resource.Resource {
	return &aws.AWSRDSDatabaseResource{}
}

func NewSlackIntegrationResource() resource.Resource {
	return &slack.SlackIntegrationResource{}
}

func NewPagerDutyIntegrationResource() resource.Resource {
	return &pagerduty.PagerDutyIntegrationResource{}
}
func NewAWSIDCAccountAvailabilitiesResource() resource.Resource {
	return &aws.AWSIDCAccountAvailabilitiesResource{}
}
func NewAWSAccountSelectorResource() resource.Resource {
	return &aws.AWSAccountSelectorResource{}
}
func NewOpsGenieIntegrationResource() resource.Resource {
	return &opsgenie.OpsGenieIntegrationResource{}
}
func NewEntraIntegrationResource() resource.Resource {
	return &entra.EntraIntegrationResource{}
}
func NewEntraGroupSelectorResource() resource.Resource {
	return &entra.EntraGroupSelectorResource{}
}
func NewEntraGroupAvailabilitiesResource() resource.Resource {
	return &entra.EntraGroupAvailabilitiesResource{}
}

func NewAWSRDSSelectorResource() resource.Resource {
	return &aws.AWSRDSSelectorResource{}
}
func NewAWSRDSAvailabilitiesResource() resource.Resource {
	return &aws.AWSRDSAvailabilitiesResource{}
}
func NewAWSIDCGroupAvailabilitiesResource() resource.Resource {
	return &aws.AWSIDCGroupAvailabilitiesResource{}
}
func NewAWSIDCGroupSelectorResource() resource.Resource {
	return &aws.AWSIDCGroupSelectorResource{}
}
func NewOktaIntegrationResource() resource.Resource {
	return &okta.OktaIntegrationResource{}
}
func NewOktaGroupSelectorResource() resource.Resource {
	return &okta.OktaGroupSelectorResource{}
}
func NewOktaGroupAvailabilitiesResource() resource.Resource {
	return &okta.OktaGroupAvailabilitiesResource{}
}

func NewDataStaxIntegrationResource() resource.Resource {
	return &datastax.DataStaxIntegrationResource{}
}

func NewDataStaxOrganizationAvailabilitiesResource() resource.Resource {
	return &datastax.DataStaxOrganizationAvailabilitiesResource{}
}

func NewDataStaxOrganizationSelectorResource() resource.Resource {
	return &datastax.DataStaxOrganizationSelectorResource{}
}

func NewGCPBigQueryTableAvailabilitiesResource() resource.Resource {
	return &gcp.GCPBigQueryTableAvailabilitiesResource{}
}

func NewGCPBigQueryTableSelectorResource() resource.Resource {
	return &gcp.GCPBigQueryTableSelectorResource{}
}

func NewGCPBigQueryDatasetAvailabilitiesResource() resource.Resource {
	return &gcp.GCPBigQueryDatasetAvailabilitiesResource{}
}

func NewGCPBigQueryDatasetSelectorResource() resource.Resource {
	return &gcp.GCPBigQueryDatasetSelectorResource{}
}

func NewWebhookIntegrationResource() resource.Resource {
	return &webhook.WebhookIntegrationResource{}
}

func NewAuth0OrganizationAvailabilitiesResource() resource.Resource {
	return &auth0.Auth0OrganizationAvailabilitiesResource{}
}

func NewAuth0OrganizationSelectorResource() resource.Resource {
	return &auth0.Auth0OrganizationSelectorResource{}
}

func NewAuth0IntegrationResource() resource.Resource {
	return &auth0.Auth0IntegrationResource{}
}

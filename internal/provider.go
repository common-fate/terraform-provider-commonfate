package internal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	governance "github.com/common-fate/common-fate/governance/pkg/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	return &commonfateProvider{}
}

// commonfateProvider is the provider implementation.
type commonfateProvider struct {
	GovernanceAPIURL types.String `tfsdk:"governance_api_url"`
	AWSRegion        types.String `tfsdk:"aws_region"`
	AssumeRoleARN    types.String `tfsdk:"assume_role_arn"`
}

// commonfateProviderModel maps provider schema data to a Go type.
type commonfateProviderModel struct {
	GovernanceAPIURL types.String `tfsdk:"governance_api_url"`
	AWSRegion        types.String `tfsdk:"aws_region"`
	AssumeRoleARN    types.String `tfsdk:"assume_role_arn"`
}

// Metadata returns the provider type name.
func (p *commonfateProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "commonfate"
}

// GetSchema defines the provider-level schema for configuration data.
func (p *commonfateProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"governance_api_url": schema.StringAttribute{
				Required: true,
			},
			"aws_region": schema.StringAttribute{
				Required: true,
			},
			"assume_role_arn": schema.StringAttribute{
				Optional: true,
			},
		},
	}
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

	if config.GovernanceAPIURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("governance_api_url"),
			"Unknown Common Fate API Host",
			"The provider cannot create the Common Fate API client as there is an unknown configuration value for the Common Fate Governance API.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	governanceAPIURL := config.GovernanceAPIURL.ValueString()

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if governanceAPIURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("governance_api_url"),
			"Missing Common Fate API Host",
			"The provider cannot create the Common Fate API client as there is a missing or empty value for the Common Fate API host. "+
				"Set the host value in the configuration or use the COMMONFATE_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	region := config.AWSRegion.ValueString()

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if region == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("aws_region"),
			"Missing AWS region",
			"The provider cannot create the Common Fate API client as there is a missing or empty value for the Common Fate API region. "+
				"Make sure the aws_region provider variable is set.",
		)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
	)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client"),
			"error creating AWS config",
			err.Error(),
		)
		return
	}

	assumeRoleARN := config.AssumeRoleARN.ValueString()
	if assumeRoleARN != "" {
		tflog.Debug(ctx, "assuming role", map[string]interface{}{"role": assumeRoleARN})

		stsclient := sts.NewFromConfig(awsCfg)
		awsCfg, err = awsconfig.LoadDefaultConfig(
			ctx, awsconfig.WithRegion(region),
			awsconfig.WithCredentialsProvider(aws.NewCredentialsCache(
				stscreds.NewAssumeRoleProvider(
					stsclient,
					assumeRoleARN,
				)),
			),
		)
		if err != nil {
			resp.Diagnostics.AddError("error assuming role", err.Error())
			return
		}
	}

	creds, err := awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client"),
			"error fetching AWS credentials",
			err.Error(),
		)
		return
	}

	client, err := governance.NewClientWithResponses(governanceAPIURL, governance.WithRequestEditorFn(apiGatewayRequestSigner(creds, region)))
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client"),
			"error creating Common Fate API client",
			err.Error(),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Make the Common Fate client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Common Fate client", map[string]any{"success": true})
}

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

// apiGatewayRequestSigner uses the AWS SDK to sign the request with sigv4
// Docs are scarce for this however I found this good example repo which is a little old but has some gems in it
// https://github.com/smarty-archives/go-aws-auth
func apiGatewayRequestSigner(creds aws.Credentials, region string) governance.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) (err error) {
		signer := v4.NewSigner()
		h := sha256.New()
		var b []byte
		if req.Body != nil {
			b, err = io.ReadAll(req.Body)
			// after you read the body you need to replace it with a new readcloser!
			req.Body = io.NopCloser(bytes.NewReader(b))
			if err != nil {
				return err
			}
		}

		_, err = h.Write(b)
		if err != nil {
			return err
		}
		sha256_hash := hex.EncodeToString(h.Sum(nil))
		return signer.SignHTTP(ctx, creds, req, sha256_hash, "execute-api", region, time.Now())
	}
}

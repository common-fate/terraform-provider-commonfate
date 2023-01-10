package commonfate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	governance "github.com/common-fate/common-fate/governance/pkg/types"
	"github.com/common-fate/common-fate/pkg/cfaws"

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
	Host   types.String `tfsdk:"host"`
	Region types.String `tfsdk:"region"`

	Version types.String `tfsdk:"version"`
}

// commonfateProviderModel maps provider schema data to a Go type.
type commonfateProviderModel struct {
	Host   types.String `tfsdk:"host"`
	Region types.String `tfsdk:"region"`

	Version types.String `tfsdk:"version"`
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
				Required: true,
			},
			"region": {
				Type:     types.StringType,
				Required: true,
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

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("COMMONFATE_HOST")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
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

	region := os.Getenv("COMMONFATE_REGION")

	if !config.Region.IsNull() {
		region = config.Region.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if region == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Missing Common Fate API Host",
			"The provider cannot create the Common Fate API client as there is a missing or empty value for the Common Fate API region. "+
				"Set the host value in the configuration or use the COMMONFATE_REGION environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	awsCfg, err := cfaws.ConfigFromContextOrDefault(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client"),
			"error setting up client",
			"error",
		)
		return
	}
	creds, err := awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client"),
			"error setting up client",
			"error",
		)
		return
	}

	client, err := governance.NewClientWithResponses(host, governance.WithRequestEditorFn(apiGatewayRequestSigner(creds, region)))
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("client"),
			"error setting up client",
			"error",
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured commonfate client", map[string]any{"success": true})

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

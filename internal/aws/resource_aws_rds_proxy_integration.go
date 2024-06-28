package aws

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	integrationv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1/integrationv1alpha1connect"
	"github.com/common-fate/sdk/service/control/integration"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSRDSProxyIntegrationModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	IDCInstanceARN        types.String `tfsdk:"idc_instance_arn"`
	IDCRegion             types.String `tfsdk:"idc_region"`
	IDCProvisionerRoleARN types.String `tfsdk:"idc_provisioner_role_arn"`
	ProxyRegion           types.String `tfsdk:"proxy_region"`
	ProxyAccount          types.String `tfsdk:"proxy_account"`
}

type AWSRDSProxyIntegrationResource struct {
	client integrationv1alpha1connect.IntegrationServiceClient
}

var (
	_ resource.Resource                = &AWSRDSProxyIntegrationResource{}
	_ resource.ResourceWithConfigure   = &AWSRDSProxyIntegrationResource{}
	_ resource.ResourceWithImportState = &AWSRDSProxyIntegrationResource{}
)

// Metadata returns the data source type name.
func (r *AWSRDSProxyIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_rds_proxy_integration"
}

// Configure adds the provider configured client to the data source.
func (r *AWSRDSProxyIntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config_client.Context)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	client := integration.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *AWSRDSProxyIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers an AWS RDS Proxy integration`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the integration: use a short label which is descriptive of the organization you're connecting to",
				Required:            true,
			},
			"idc_instance_arn": schema.StringAttribute{
				MarkdownDescription: "The ARN of the IAM Identity Center SSO instance",
				Required:            true,
			},
			"idc_region": schema.StringAttribute{
				MarkdownDescription: "The AWS region that the SSO instance is hosted in",
				Required:            true,
			},
			"idc_provisioner_role_arn": schema.StringAttribute{
				MarkdownDescription: "The ARN of the role to assume in order to provision access in AWS RDS Proxy",
				Required:            true,
			},
			"proxy_region": schema.StringAttribute{
				MarkdownDescription: "The AWS region that the Proxy is hosted in",
				Required:            true,
			},
			"proxy_account": schema.StringAttribute{
				MarkdownDescription: "The AWS Account that the Proxy is hosted in",
				Required:            true,
			},
		},
		MarkdownDescription: `Registers an AWS RDS Proxy integration`,
	}
}

func (r *AWSRDSProxyIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AWSRDSProxyIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.CreateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.CreateIntegrationRequest{
		Name: data.Name.ValueString(),
		Config: &integrationv1alpha1.Config{
			Config: &integrationv1alpha1.Config_AwsRdsProxy{
				AwsRdsProxy: &integrationv1alpha1.AWSRDSProxy{
					IdcInstanceArn:        data.IDCInstanceARN.ValueString(),
					IdcRegion:             data.IDCRegion.ValueString(),
					IdcProvisionerRoleArn: data.IDCProvisionerRoleARN.ValueString(),
					ProxyAwsRegion:        data.ProxyRegion.ValueString(),
					ProxyAwsAccount:       data.ProxyAccount.ValueString(),
				},
			},
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: AWS RDS Proxy Integration",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	data.Id = types.StringValue(res.Msg.Integration.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AWSRDSProxyIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AWSRDSProxyIntegrationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.GetIntegration(ctx, connect.NewRequest(&integrationv1alpha1.GetIntegrationRequest{
		Id: state.Id.ValueString(),
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read AWS RDS Proxy Integration",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AWSRDSProxyIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AWSRDSProxyIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to update AWS RDS Proxy Integration",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}

	res, err := r.client.UpdateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.UpdateIntegrationRequest{
		Integration: &integrationv1alpha1.Integration{
			Id:   data.Id.ValueString(),
			Name: data.Name.ValueString(),
			Config: &integrationv1alpha1.Config{
				Config: &integrationv1alpha1.Config_AwsRdsProxy{
					AwsRdsProxy: &integrationv1alpha1.AWSRDSProxy{
						IdcInstanceArn:        data.IDCInstanceARN.ValueString(),
						IdcRegion:             data.IDCRegion.ValueString(),
						IdcProvisionerRoleArn: data.IDCProvisionerRoleARN.ValueString(),
						ProxyAwsRegion:        data.ProxyRegion.ValueString(),
						ProxyAwsAccount:       data.ProxyAccount.ValueString(),
					},
				},
			},
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update AWS RDS Proxy Integration",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	data.Id = types.StringValue(res.Msg.Integration.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSRDSProxyIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AWSRDSProxyIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete AWS RDS Proxy Integration",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteIntegration(ctx, connect.NewRequest(&integrationv1alpha1.DeleteIntegrationRequest{
		Id: data.Id.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete AWS RDS Proxy Integration",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *AWSRDSProxyIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

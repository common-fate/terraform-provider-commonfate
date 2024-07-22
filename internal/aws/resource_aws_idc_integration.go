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

type AWSIDCIntegrationModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	SSOInstanceARN     types.String `tfsdk:"sso_instance_arn"`
	IdentityStoreID    types.String `tfsdk:"identity_store_id"`
	SSORegion          types.String `tfsdk:"sso_region"`
	ReaderRoleARN      types.String `tfsdk:"reader_role_arn"`
	AuditRoleName      types.String `tfsdk:"audit_role_name"`
	ResourceRegions    types.Set    `tfsdk:"resource_regions"`
	SSOAccessPortalURL types.String `tfsdk:"sso_access_portal_url"`
	ProvisionerRoleARN types.String `tfsdk:"provisioner_role_arn"`
}

type AWSIDCIntegrationResource struct {
	client integrationv1alpha1connect.IntegrationServiceClient
}

var (
	_ resource.Resource                = &AWSIDCIntegrationResource{}
	_ resource.ResourceWithConfigure   = &AWSIDCIntegrationResource{}
	_ resource.ResourceWithImportState = &AWSIDCIntegrationResource{}
)

// Metadata returns the data source type name.
func (r *AWSIDCIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_idc_integration"
}

// Configure adds the provider configured client to the data source.
func (r *AWSIDCIntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AWSIDCIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers an AWS IAM Identity Center integration`,
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
			"sso_instance_arn": schema.StringAttribute{
				MarkdownDescription: "The ARN of the IAM Identity Center SSO instance",
				Required:            true,
			},
			"sso_region": schema.StringAttribute{
				MarkdownDescription: "The AWS region that the SSO instance is hosted in",
				Required:            true,
			},
			"identity_store_id": schema.StringAttribute{
				MarkdownDescription: "The IAM Identity Center identity store ID",
				Required:            true,
			},
			"reader_role_arn": schema.StringAttribute{
				MarkdownDescription: "The ARN of the role to assume in order to read AWS IAM Identity Center data",
				Required:            true,
			},
			"provisioner_role_arn": schema.StringAttribute{
				MarkdownDescription: "The ARN of the role to assume in order to provision access in AWS IAM Identity Center",
				Optional:            true,
			},
			"audit_role_name": schema.StringAttribute{
				MarkdownDescription: "The name of the role to assume in each AWS Account in order to read resources",
				Optional:            true,
			},
			"resource_regions": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The regions to read reasources from in each account",
				Optional:            true,
			},
			"sso_access_portal_url": schema.StringAttribute{
				MarkdownDescription: "The SSO access portal URL, e.g https://example.awsapps.com/start",
				Optional:            true,
			},
		},
		MarkdownDescription: `Registers an AWS IAM Identity Center integration`,
	}
}

func (r *AWSIDCIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AWSIDCIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	var resourceRegions []string
	diag := data.ResourceRegions.ElementsAs(ctx, &resourceRegions, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	res, err := r.client.CreateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.CreateIntegrationRequest{
		Name: data.Name.ValueString(),
		Config: &integrationv1alpha1.Config{
			Config: &integrationv1alpha1.Config_AwsIdc{
				AwsIdc: &integrationv1alpha1.AWSIDC{
					SsoInstanceArn:     data.SSOInstanceARN.ValueString(),
					IdentityStoreId:    data.IdentityStoreID.ValueString(),
					SsoRegion:          data.SSORegion.ValueString(),
					ReaderRoleArn:      data.ReaderRoleARN.ValueString(),
					ProvisionerRoleArn: data.ProvisionerRoleARN.ValueString(),
					AuditRoleName:      data.AuditRoleName.ValueString(),
					ResourceRegions:    resourceRegions,
					SsoAccessPortalUrl: data.SSOAccessPortalURL.ValueString(),
				},
			},
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: AWS IAM Identity Center Integration",
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
func (r *AWSIDCIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AWSIDCIntegrationModel

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
			"Failed to read AWS IAM Identity Center Integration",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AWSIDCIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AWSIDCIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to update AWS IAM Identity Center Integration",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}
	var resourceRegions []string
	diag := data.ResourceRegions.ElementsAs(ctx, &resourceRegions, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	res, err := r.client.UpdateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.UpdateIntegrationRequest{
		Integration: &integrationv1alpha1.Integration{
			Id:   data.Id.ValueString(),
			Name: data.Name.ValueString(),
			Config: &integrationv1alpha1.Config{
				Config: &integrationv1alpha1.Config_AwsIdc{
					AwsIdc: &integrationv1alpha1.AWSIDC{
						SsoInstanceArn:     data.SSOInstanceARN.ValueString(),
						IdentityStoreId:    data.IdentityStoreID.ValueString(),
						SsoRegion:          data.SSORegion.ValueString(),
						ReaderRoleArn:      data.ReaderRoleARN.ValueString(),
						ProvisionerRoleArn: data.ProvisionerRoleARN.ValueString(),
						AuditRoleName:      data.AuditRoleName.ValueString(),
						ResourceRegions:    resourceRegions,
						SsoAccessPortalUrl: data.SSOAccessPortalURL.ValueString(),
					},
				},
			},
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update AWS IAM Identity Center Integration",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	data.Id = types.StringValue(res.Msg.Integration.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSIDCIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AWSIDCIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete AWS IAM Identity Center Integration",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteIntegration(ctx, connect.NewRequest(&integrationv1alpha1.DeleteIntegrationRequest{
		Id: data.Id.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete AWS IAM Identity Center Integration",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *AWSIDCIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

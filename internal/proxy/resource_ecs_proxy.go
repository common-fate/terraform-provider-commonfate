package proxy

import (
	"context"
	"fmt"

	integrationv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1"

	config_client "github.com/common-fate/sdk/config"
	"github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1/integrationv1alpha1connect"
	"github.com/common-fate/sdk/service/control/integration"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ECSProxyModel struct {
	ID           types.String `tfsdk:"id"`
	AwsRegion    types.String `tfsdk:"aws_region"`
	AwsAccountId types.String `tfsdk:"aws_account_id"`

	ECSClusterName            types.String `tfsdk:"ecs_cluster_name"`
	ECSTaskDefinitionFamily   types.String `tfsdk:"ecs_task_definition_family"`
	ECSClusterReaderRoleARN   types.String `tfsdk:"ecs_cluster_reader_role_arn"`
	ECSClusterSecurityGroupId types.String `tfsdk:"ecs_cluster_security_group_id"`
	ECSClusterTaskRoleName types.String `tfsdk:"ecs_cluster_task_role_name"`

}

// AccessRuleResource is the data source implementation.
type ECSProxyResource struct {
	client integrationv1alpha1connect.IntegrationServiceClient
}

var (
	_ resource.Resource                = &ECSProxyResource{}
	_ resource.ResourceWithConfigure   = &ECSProxyResource{}
	_ resource.ResourceWithImportState = &ECSProxyResource{}
)

// Metadata returns the data source type name.
func (r *ECSProxyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ecs_proxy"
}

// Configure adds the provider configured client to the data source.
func (r *ECSProxyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ECSProxyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the proxy. Eg: prod-us-west-2",
				Required:            true,

			},

			"aws_region": schema.StringAttribute{
				MarkdownDescription: "The AWS region the proxy will be installed to.",
				Required:            true,
			},

			"aws_account_id": schema.StringAttribute{
				MarkdownDescription: "The AWS account the proxy is installed in.",
				Required:            true,
			},
			"ecs_cluster_name": schema.StringAttribute{
				MarkdownDescription: "The ECS cluster name of the proxy.",
				Required:            true,
			},
			"ecs_task_definition_family": schema.StringAttribute{
				MarkdownDescription: "The ECS task definition family of the proxy.",
				Required:            true,
			},
			"ecs_cluster_reader_role_arn": schema.StringAttribute{
				MarkdownDescription: "The ECS cluster reader role ARN of the proxy.",
				Required:            true,
			},
			"ecs_cluster_security_group_id": schema.StringAttribute{
				MarkdownDescription: "The ECS cluster security group ID.",
				Required:            true,
			},
			"ecs_cluster_task_role_name": schema.StringAttribute{
				MarkdownDescription: "The ECS cluster task role ARN.",
				Required:            true,
			},

		},
		MarkdownDescription: `Registers a proxy with Common Fate..`,
	}
}

func (r *ECSProxyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *ECSProxyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	createReq := integrationv1alpha1.RegisterProxyRequest{
		Id: data.ID.ValueString(),
		InstanceConfig: &integrationv1alpha1.RegisterProxyRequest_AwsEcsProxyInstanceConfig{
			AwsEcsProxyInstanceConfig: &integrationv1alpha1.AWSECSProxyInstanceConfig{
				EcsClusterName:          data.ECSClusterName.ValueString(),
				Account:                 data.AwsAccountId.ValueString(),
				Region:                  data.AwsRegion.ValueString(),
				EcsTaskDefinitionFamily: data.ECSTaskDefinitionFamily.ValueString(),
				EcsContainerName:        data.ECSClusterName.ValueString(),
				EcsClusterReaderRoleArn: data.ECSClusterReaderRoleARN.ValueString(),
				EcsClusterSecurityGroupId: data.ECSClusterSecurityGroupId.ValueString(),
				EcsClusterTaskRoleName: data.ECSClusterTaskRoleName.ValueString(),

			},
		},
	}

	res, err := r.client.RegisterProxy(ctx, connect.NewRequest(&createReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: RDS Proxy",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ECSProxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state ECSProxyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// read the state from the client
	res, err := r.client.GetProxy(ctx, connect.NewRequest(&integrationv1alpha1.GetProxyRequest{
		Id: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read RDS resource",
			err.Error(),
		)
		return
	}

	// refresh state

	state.ID = types.StringValue(res.Msg.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ECSProxyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *ECSProxyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	updateReq := integrationv1alpha1.UpdateProxyRequest{
		Id: data.ID.ValueString(),
		InstanceConfig: &integrationv1alpha1.UpdateProxyRequest_AwsEcsProxyInstanceConfig{
			AwsEcsProxyInstanceConfig: &integrationv1alpha1.AWSECSProxyInstanceConfig{
				EcsClusterName:          data.ECSClusterName.ValueString(),
				Account:                 data.AwsAccountId.ValueString(),
				Region:                  data.AwsRegion.ValueString(),
				EcsContainerName:        data.ECSClusterName.ValueString(),
				EcsClusterReaderRoleArn: data.ECSClusterReaderRoleARN.ValueString(),
				EcsClusterSecurityGroupId: data.ECSClusterSecurityGroupId.ValueString(),
				EcsClusterTaskRoleName: data.ECSClusterTaskRoleName.ValueString(),
			},
		},
	}

	res, err := r.client.UpdateProxy(ctx, connect.NewRequest(&updateReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource: ECS Proxy",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ECSProxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *ECSProxyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteProxy(ctx, connect.NewRequest(&integrationv1alpha1.DeleteProxyRequest{
		Id: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *ECSProxyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

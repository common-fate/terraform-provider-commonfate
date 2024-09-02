package proxy

import (
	"context"
	"fmt"

	integrationv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1"

	config_client "github.com/common-fate/sdk/config"
	"github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1/integrationv1alpha1connect"
	"github.com/common-fate/sdk/service/control/integration"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AccessRuleResource is the data source implementation.
type ECSProxyDatasource struct {
	client integrationv1alpha1connect.IntegrationServiceClient
}

var _ datasource.DataSource = &ECSProxyDatasource{}

// Metadata returns the data source type name.
func (r *ECSProxyDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ecs_proxy"
}

// Configure adds the provider configured client to the data source.
func (r *ECSProxyDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (r *ECSProxyDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the proxy. Eg: prod-us-west-2",
				Required:            true,
			},

			"aws_region": schema.StringAttribute{
				MarkdownDescription: "The AWS region the proxy is installed to.",
				Computed:            true,
			},

			"aws_account_id": schema.StringAttribute{
				MarkdownDescription: "The AWS account the proxy is installed in.",
				Computed:            true,
			},
			"ecs_cluster_name": schema.StringAttribute{
				MarkdownDescription: "The ECS cluster name of the proxy.",
				Computed:            true,
			},
			"ecs_task_definition_family": schema.StringAttribute{
				MarkdownDescription: "The ECS task definition family of the proxy.",
				Computed:            true,
			},
			"ecs_cluster_reader_role_arn": schema.StringAttribute{
				MarkdownDescription: "The ECS cluster reader role ARN of the proxy.",
				Computed:            true,
			},
			"ecs_cluster_security_group_id": schema.StringAttribute{
				MarkdownDescription: "The ECS cluster security group ID",
				Computed:            true,
			},
			"ecs_cluster_task_role_name": schema.StringAttribute{
				MarkdownDescription: "The ECS cluster task role ARN.",
				Computed:            true,
			},
		},
		MarkdownDescription: `.`,
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ECSProxyDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}

	var state ECSProxyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

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

	state = ECSProxyModel{
		ID:                        types.StringValue(res.Msg.Id),
		AwsRegion:                 types.StringValue(res.Msg.GetAwsEcsProxyInstanceConfig().Region),
		AwsAccountID:              types.StringValue(res.Msg.GetAwsEcsProxyInstanceConfig().Account),
		ECSClusterName:            types.StringValue(res.Msg.GetAwsEcsProxyInstanceConfig().EcsClusterName),
		ECSTaskDefinitionFamily:   types.StringValue(res.Msg.GetAwsEcsProxyInstanceConfig().EcsTaskDefinitionFamily),
		ECSClusterReaderRoleARN:   types.StringValue(res.Msg.GetAwsEcsProxyInstanceConfig().EcsClusterReaderRoleArn),
		ECSClusterSecurityGroupID: types.StringValue(res.Msg.GetAwsEcsProxyInstanceConfig().EcsClusterSecurityGroupId),
		ECSClusterTaskRoleName:    types.StringValue(res.Msg.GetAwsEcsProxyInstanceConfig().EcsClusterTaskRoleName),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

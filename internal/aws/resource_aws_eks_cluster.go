package aws

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	"github.com/common-fate/sdk/service/control/configsvc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSEKSClusterModel struct {
	ID            types.String `tfsdk:"id"`
	ARN           types.String `tfsdk:"arn"`
	Region        types.String `tfsdk:"region"`
	Name          types.String `tfsdk:"name"`
	AWSAccountID  types.String `tfsdk:"aws_account_id"`
	IntegrationID types.String `tfsdk:"integration_id"`
}

type AWSEKSClusterResource struct {
	client configv1alpha1connect.EKSClusterServiceClient
}

func NewAWSEKSClusterResource() resource.Resource {
	return &AWSEKSClusterResource{}
}

var (
	_ resource.Resource                = &AWSEKSClusterResource{}
	_ resource.ResourceWithConfigure   = &AWSEKSClusterResource{}
	_ resource.ResourceWithImportState = &AWSEKSClusterResource{}
)

// Metadata returns the data source type name.
func (r *AWSEKSClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_eks_cluster"
}

// Configure adds the provider configured client to the data source.
func (r *AWSEKSClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = configsvc.NewFromConfig(cfg).AWSEKSCluster()
}

// GetSchema defines the schema for the data source.
func (r *AWSEKSClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers an AWS EKS Cluster with Common Fate`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The EKS cluster ARN",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"arn": schema.StringAttribute{
				MarkdownDescription: "The EKS cluster ARN",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The AWS region the cluster is located in",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the cluster",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_account_id": schema.StringAttribute{
				MarkdownDescription: "The AWS account ID that the cluster belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"integration_id": schema.StringAttribute{
				MarkdownDescription: "The EKS integration ID used to provision access to the cluster",
				Required:            true,
			},
		},
		MarkdownDescription: `Registers an AWS EKS Cluster with Common Fate`,
	}
}

func (r *AWSEKSClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AWSEKSClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.CreateEKSCluster(ctx, connect.NewRequest(&configv1alpha1.CreateEKSClusterRequest{
		Name:          data.Name.ValueString(),
		Region:        data.Region.ValueString(),
		AwsAccountId:  data.AWSAccountID.ValueString(),
		IntegrationId: data.IntegrationID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: AWS EKS Cluster",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Cluster.Arn)
	data.ARN = types.StringValue(res.Msg.Cluster.Arn)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AWSEKSClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AWSEKSClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	res, err := r.client.GetEKSCluster(ctx, connect.NewRequest(&configv1alpha1.GetEKSClusterRequest{
		Arn: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read AWS EKS Cluster",
			err.Error(),
		)
		return
	}

	// refresh state
	state = AWSEKSClusterModel{
		ID:            types.StringValue(res.Msg.Cluster.Arn),
		ARN:           types.StringValue(res.Msg.Cluster.Arn),
		Region:        types.StringValue(res.Msg.Cluster.Region),
		Name:          types.StringValue(res.Msg.Cluster.Name),
		AWSAccountID:  types.StringValue(res.Msg.Cluster.AwsAccountId),
		IntegrationID: types.StringValue(res.Msg.Cluster.IntegrationId),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AWSEKSClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AWSEKSClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}

	res, err := r.client.UpdateEKSCluster(ctx, connect.NewRequest(&configv1alpha1.UpdateEKSClusterRequest{
		Arn:           data.ID.ValueString(),
		IntegrationId: data.IntegrationID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",

			"JSON Error: "+err.Error(),
		)

		return

	}

	data.ID = types.StringValue(res.Msg.Cluster.Arn)
	data.ARN = types.StringValue(res.Msg.Cluster.Arn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSEKSClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AWSEKSClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteEKSCluster(ctx, connect.NewRequest(&configv1alpha1.DeleteEKSClusterRequest{
		Arn: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Resource", err.Error(),
		)

		return
	}
}

func (r *AWSEKSClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

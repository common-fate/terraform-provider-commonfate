package proxy

import (
	"context"
	"fmt"

	integrationv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1"

	config_client "github.com/common-fate/sdk/config"
	"github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1/integrationv1alpha1connect"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EKSClusterModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Region       types.String `tfsdk:"region"`
	AWSAccountID types.String `tfsdk:"aws_account_id"`
	ProxyId      types.String `tfsdk:"proxy_id"`
}

type EKSClusterResource struct {
	client integrationv1alpha1connect.ProxyServiceClient
}

var (
	_ resource.Resource                = &EKSClusterResource{}
	_ resource.ResourceWithConfigure   = &EKSClusterResource{}
	_ resource.ResourceWithImportState = &EKSClusterResource{}
)

// Metadata returns the data source type name.
func (r *EKSClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy_eks_cluster"
}

// Configure adds the provider configured client to the data source.
func (r *EKSClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := integrationv1alpha1connect.NewProxyServiceClient(cfg.HTTPClient, cfg.APIURL)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *EKSClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Access Workflows are used to describe how long access should be applied. Created Workflows can be referenced in other resources created.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "A unique name for the resource",
				Required:            true,
			},

			"region": schema.StringAttribute{
				MarkdownDescription: "The region the database is in",
				Required:            true,
			},
			"aws_account_id": schema.StringAttribute{
				MarkdownDescription: "The AWS account id the database is in",
				Required:            true,
			},

			"proxy_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the proxy in the same account as the database.",
				Required:            true,
			},
		},
		MarkdownDescription: `Registers a EKS Cluster database with a Common Fate Proxy.`,
	}
}

func (r *EKSClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *EKSClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	resource := &integrationv1alpha1.AWSEKSCluster{
		Name:    data.Name.ValueString(),
		Region:  data.Region.ValueString(),
		Account: data.AWSAccountID.ValueString(),
	}

	createReq := integrationv1alpha1.CreateProxyEksClusterResourceRequest{
		ProxyId:    data.ProxyId.ValueString(),
		EksCluster: resource,
	}

	res, err := r.client.CreateProxyEksClusterResource(ctx, connect.NewRequest(&createReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: EKS Cluster Database",
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
func (r *EKSClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state EKSClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// read the state from the client
	res, err := r.client.GetProxyEksClusterResource(ctx, connect.NewRequest(&integrationv1alpha1.GetProxyEksClusterResourceRequest{
		Id: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read EKS Cluster resource",
			err.Error(),
		)
		return
	}

	// refresh state

	state.ID = types.StringValue(res.Msg.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EKSClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *EKSClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	resource := &integrationv1alpha1.AWSEKSCluster{
		Name:    data.Name.ValueString(),
		Region:  data.Region.ValueString(),
		Account: data.AWSAccountID.ValueString(),
	}

	updateReq := integrationv1alpha1.UpdateProxyEksClusterResourceRequest{
		Id:         data.ID.ValueString(),
		ProxyId:    data.ProxyId.ValueString(),
		EksCluster: resource,
	}

	res, err := r.client.UpdateProxyEksClusterResource(ctx, connect.NewRequest(&updateReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource: EKS Cluster Database",
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

func (r *EKSClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *EKSClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteProxyEksClusterResource(ctx, connect.NewRequest(&integrationv1alpha1.DeleteProxyEksClusterResourceRequest{
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

func (r *EKSClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

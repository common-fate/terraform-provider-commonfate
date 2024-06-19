package aws

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	"github.com/common-fate/sdk/service/control/config/awsrdsdatabase"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSRDSDatabaseModel struct {
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	DatabaseUser         types.String `tfsdk:"database_user"`
	ProxyInstanceAccount types.String `tfsdk:"proxy_instance_account"`
	ProxyInstanceRegion  types.String `tfsdk:"proxy_instance_region"`
}

type AWSRDSDatabaseResource struct {
	client configv1alpha1connect.AWSRDSDatabaseServiceClient
}

var (
	_ resource.Resource                = &AWSRDSDatabaseResource{}
	_ resource.ResourceWithConfigure   = &AWSRDSDatabaseResource{}
	_ resource.ResourceWithImportState = &AWSRDSDatabaseResource{}
)

// Metadata returns the data source type name.
func (r *AWSRDSDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_rds_database"
}

// Configure adds the provider configured client to the data source.
func (r *AWSRDSDatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := awsrdsdatabase.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *AWSRDSDatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers an AWS RDS integration`,
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
			"database_user": schema.StringAttribute{
				MarkdownDescription: "The database user/role",
				Required:            true,
			},
			"proxy_instance_account": schema.StringAttribute{
				MarkdownDescription: "The AWS account ID where the proxy is deployed",
				Required:            true,
			},
			"proxy_instance_region": schema.StringAttribute{
				MarkdownDescription: "The AWS region where the proxy is deployed",
				Required:            true,
			},
		},
		MarkdownDescription: `Registers an AWS RDS integration`,
	}
}

func (r *AWSRDSDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AWSRDSDatabaseModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.CreateAWSRDSDatabase(ctx, connect.NewRequest(&configv1alpha1.CreateAWSRDSDatabaseRequest{
		Name:                 data.Name.ValueString(),
		DatabaseUser:         data.DatabaseUser.ValueString(),
		ProxyInstanceAccount: data.ProxyInstanceAccount.ValueString(),
		ProxyInstanceRegion:  data.ProxyInstanceRegion.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: AWS RDS Database",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	data.Id = types.StringValue(res.Msg.AwsRdsDatabase.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AWSRDSDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AWSRDSDatabaseModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.GetAWSRDSDatabase(ctx, connect.NewRequest(&configv1alpha1.GetAWSRDSDatabaseRequest{
		Id: state.Id.ValueString(),
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read AWS RDS Database",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AWSRDSDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AWSRDSDatabaseModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to update AWS RDS Database",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}

	res, err := r.client.UpdateAWSRDSDatabase(ctx, connect.NewRequest(&configv1alpha1.UpdateAWSRDSDatabaseRequest{
		AwsRdsDatabase: &configv1alpha1.AWSRDSDatabase{
			Id:                   data.Id.ValueString(),
			Name:                 data.Name.ValueString(),
			DatabaseUser:         data.DatabaseUser.ValueString(),
			ProxyInstanceAccount: data.ProxyInstanceAccount.ValueString(),
			ProxyInstanceRegion:  data.ProxyInstanceRegion.ValueString(),
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update AWS RDS Database",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	data.Id = types.StringValue(res.Msg.AwsRdsDatabase.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSRDSDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AWSRDSDatabaseModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete AWS RDS Database",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteAWSRDSDatabase(ctx, connect.NewRequest(&configv1alpha1.DeleteAWSRDSDatabaseRequest{
		Id: data.Id.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete AWS RDS Database",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *AWSRDSDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package aws

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	"github.com/common-fate/sdk/service/control/config/awsrdsdatabaseuser"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSRDSDatabaseUserModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseID types.String `tfsdk:"database_id"`
}

type AWSRDSDatabaseUserResource struct {
	client configv1alpha1connect.AWSRDSDatabaseUserServiceClient
}

var (
	_ resource.Resource                = &AWSRDSDatabaseUserResource{}
	_ resource.ResourceWithConfigure   = &AWSRDSDatabaseUserResource{}
	_ resource.ResourceWithImportState = &AWSRDSDatabaseUserResource{}
)

// Metadata returns the data source type name.
func (r *AWSRDSDatabaseUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_rds_database_user"
}

// Configure adds the provider configured client to the data source.
func (r *AWSRDSDatabaseUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := awsrdsdatabaseuser.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *AWSRDSDatabaseUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers an AWS RDS user`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user",
				Required:            true,
			},
			"database_id": schema.StringAttribute{
				MarkdownDescription: "The database id",
				Required:            true,
			},
		},
		MarkdownDescription: `Registers an AWS RDS user`,
	}
}

func (r *AWSRDSDatabaseUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AWSRDSDatabaseUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.CreateAWSRDSDatabaseUser(ctx, connect.NewRequest(&configv1alpha1.CreateAWSRDSDatabaseUserRequest{
		Name:     data.Name.ValueString(),
		Database: data.DatabaseID.ValueString(),
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
func (r *AWSRDSDatabaseUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AWSRDSDatabaseUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.GetAWSRDSDatabaseUser(ctx, connect.NewRequest(&configv1alpha1.GetAWSRDSDatabaseUserRequest{
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

func (r *AWSRDSDatabaseUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AWSRDSDatabaseUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to update AWS RDS Database",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}

	res, err := r.client.UpdateAWSRDSDatabaseUser(ctx, connect.NewRequest(&configv1alpha1.UpdateAWSRDSDatabaseUserRequest{
		AwsRdsDatabase: &configv1alpha1.AWSRDSDatabaseUser{
			Id:       data.Id.ValueString(),
			Name:     data.Name.ValueString(),
			Database: data.DatabaseID.ValueString(),
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

func (r *AWSRDSDatabaseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AWSRDSDatabaseUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete AWS RDS Database",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteAWSRDSDatabaseUser(ctx, connect.NewRequest(&configv1alpha1.DeleteAWSRDSDatabaseUserRequest{
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

func (r *AWSRDSDatabaseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

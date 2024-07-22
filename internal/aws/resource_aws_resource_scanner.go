package aws

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	"github.com/common-fate/sdk/service/control/configsvc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSResourceScannerModel struct {
	ID                  types.String `tfsdk:"id"`
	IntegrationID       types.String `tfsdk:"integration_id"`
	Regions             types.Set    `tfsdk:"regions"`
	ResourceTypes       types.Set    `tfsdk:"resource_types"`
	FilterForAccountIDs types.Set    `tfsdk:"filter_for_account_ids"`
	RoleName            types.String `tfsdk:"role_name"`
}

type AWSResourceScannerResource struct {
	client configv1alpha1connect.AWSResourceScannerServiceClient
}

func NewAWSResourceScannerResource() resource.Resource {
	return &AWSResourceScannerResource{}
}

var (
	_ resource.Resource                = &AWSResourceScannerResource{}
	_ resource.ResourceWithConfigure   = &AWSResourceScannerResource{}
	_ resource.ResourceWithImportState = &AWSResourceScannerResource{}
)

// Metadata returns the data source type name.
func (r *AWSResourceScannerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_resource_scanner"
}

// Configure adds the provider configured client to the data source.
func (r *AWSResourceScannerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := configsvc.NewFromConfig(cfg).AWSResourceScanner()

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *AWSResourceScannerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers an AWS Resource Scanner`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"integration_id": schema.StringAttribute{
				MarkdownDescription: "The AWS integration ID to associate the resource scanner with",
				Required:            true,
			},
			"regions": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The regions to read reasources from in each account",
				Required:            true,
			},
			"resource_types": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Resource types to scan for. If empty, Common Fate will attempt to scan for all supported resource types. Resource types should adhere to the Cedar format, for example 'AWS::S3::Bucket'.",
				Optional:            true,
			},
			"filter_for_account_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "If provided, only accounts matching the specified ID will be scanned",
				Optional:            true,
			},
			"role_name": schema.StringAttribute{
				MarkdownDescription: "The name of the role to assume in each AWS Account in order to read resources. Defaults to 'common-fate-audit' if not provided.",
				Optional:            true,
			},
		},
		MarkdownDescription: `Registers an AWS Resource Scanner`,
	}
}

func (r *AWSResourceScannerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AWSResourceScannerModel

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
	diag := data.Regions.ElementsAs(ctx, &resourceRegions, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var resourceTypes []string
	diag = data.ResourceTypes.ElementsAs(ctx, &resourceTypes, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var filterForAccountIDs []string
	diag = data.FilterForAccountIDs.ElementsAs(ctx, &filterForAccountIDs, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	res, err := r.client.CreateAWSResourceScanner(ctx, connect.NewRequest(&configv1alpha1.CreateAWSResourceScannerRequest{
		IntegrationId:       data.IntegrationID.ValueString(),
		Regions:             resourceRegions,
		ResourceTypes:       resourceTypes,
		FilterForAccountIds: filterForAccountIDs,
		RoleName:            data.RoleName.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: AWS Resource Scanner",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	data.ID = types.StringValue(res.Msg.ResourceScanner.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AWSResourceScannerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AWSResourceScannerModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.GetAWSResourceScanner(ctx, connect.NewRequest(&configv1alpha1.GetAWSResourceScannerRequest{
		Id: state.ID.ValueString(),
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read AWS Resource Scanner",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AWSResourceScannerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AWSResourceScannerModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to update AWS Resource Scanner",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}

	var resourceRegions []string
	diag := data.Regions.ElementsAs(ctx, &resourceRegions, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var resourceTypes []string
	diag = data.ResourceTypes.ElementsAs(ctx, &resourceTypes, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var filterForAccountIDs []string
	diag = data.FilterForAccountIDs.ElementsAs(ctx, &filterForAccountIDs, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	res, err := r.client.UpdateAWSResourceScanner(ctx, connect.NewRequest(&configv1alpha1.UpdateAWSResourceScannerRequest{
		ResourceScanner: &configv1alpha1.AWSResourceScanner{
			Id:                  data.ID.ValueString(),
			IntegrationId:       data.IntegrationID.ValueString(),
			Regions:             resourceRegions,
			ResourceTypes:       resourceTypes,
			FilterForAccountIds: filterForAccountIDs,
			RoleName:            data.RoleName.ValueString(),
		},
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update AWS Resource Scanner",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	data.ID = types.StringValue(res.Msg.ResourceScanner.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSResourceScannerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AWSResourceScannerModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete AWS Resource Scanner",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteAWSResourceScanner(ctx, connect.NewRequest(&configv1alpha1.DeleteAWSResourceScannerRequest{
		Id: data.ID.ValueString(),
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete AWS Resource Scanner",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *AWSResourceScannerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package internal

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	gcp_handler "github.com/common-fate/sdk/service/control/config/gcporg"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GCPOrganizationModel struct {
	Id                   types.String `tfsdk:"id"`
	WorkloadIdentityRole types.String `tfsdk:"workload_identity_role"`
}

// AccessRuleResource is the data source implementation.
type GCPOrganizationResource struct {
	client configv1alpha1connect.GCPOrganizationServiceClient
}

var (
	_ resource.Resource                = &GCPOrganizationResource{}
	_ resource.ResourceWithConfigure   = &GCPOrganizationResource{}
	_ resource.ResourceWithImportState = &GCPOrganizationResource{}
)

// Metadata returns the data source type name.
func (r *GCPOrganizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_organization"
}

// Configure adds the provider configured client to the data source.
func (r *GCPOrganizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := gcp_handler.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *GCPOrganizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Creates a access policy that Common Fate will use to use in the approval engine.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the GCP organization",
				Required:            true,
			},
			"workload_identity_role": schema.StringAttribute{
				MarkdownDescription: "aws role to be used to be able to pull resources from GCP through a federated workload identity.",
				Required:            true,
			},
		},
		MarkdownDescription: `Creates a access policy that Common Fate will use to use in the approval engine.`,
	}
}

func (r *GCPOrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *GCPOrganizationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.CreateGCPOrganization(ctx, connect.NewRequest(&configv1alpha1.CreateGCPOrganizationRequest{
		Id:                   data.Id.ValueString(),
		WorkloadIdentityRole: data.WorkloadIdentityRole.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: Access Policy",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	data.Id = types.StringValue(res.Msg.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *GCPOrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state GCPOrganizationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.ReadGCPOrganization(ctx, connect.NewRequest(&configv1alpha1.ReadGCPOrganizationRequest{
		Id: state.Id.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read GCP Connection",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GCPOrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data GCPOrganizationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.UpdateGCPOrganization(ctx, connect.NewRequest(&configv1alpha1.UpdateGCPOrganizationRequest{
		Id:                   data.Id.ValueString(),
		WorkloadIdentityRole: data.WorkloadIdentityRole.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	data.Id = types.StringValue(res.Msg.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GCPOrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *GCPOrganizationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteGCPOrganization(ctx, connect.NewRequest(&configv1alpha1.DeleteGCPOrganizationRequest{
		Id: data.Id.ValueString(),
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

func (r *GCPOrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package gcp

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	entityv1alpha1 "github.com/common-fate/sdk/gen/commonfate/entity/v1alpha1"
	"github.com/common-fate/sdk/service/control/configsvc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GCPFolderAvailabilities struct {
	ID                  types.String `tfsdk:"id"`
	WorkflowID          types.String `tfsdk:"workflow_id"`
	Role                types.String `tfsdk:"gcp_role"`
	FolderSelectorID    types.String `tfsdk:"gcp_folder_selector_id"`
	WorkspaceCustomerID types.String `tfsdk:"google_workspace_customer_id"`
	RolePriority        types.Int64  `tfsdk:"role_priority"`
}

type GCPFolderAvailabilitiesResource struct {
	client *configsvc.Client
}

var (
	_ resource.Resource                = &GCPFolderAvailabilitiesResource{}
	_ resource.ResourceWithConfigure   = &GCPFolderAvailabilitiesResource{}
	_ resource.ResourceWithImportState = &GCPFolderAvailabilitiesResource{}
)

// Metadata returns the data source type name.
func (r *GCPFolderAvailabilitiesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_folder_availabilities"
}

// Configure adds the provider configured client to the data source.
func (r *GCPFolderAvailabilitiesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := configsvc.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *GCPFolderAvailabilitiesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: "A specifier to make GCP folders available for selection under a particular Access Workflow",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"workflow_id": schema.StringAttribute{
				MarkdownDescription: "The Access Workflow ID",
				Required:            true,
			},

			"gcp_role": schema.StringAttribute{
				MarkdownDescription: "The GCP role to make available",
				Required:            true,
			},

			"gcp_folder_selector_id": schema.StringAttribute{
				MarkdownDescription: "The target to make available. Should be a Selector entity.",
				Required:            true,
			},

			"google_workspace_customer_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Google Workspace customer associated with the folders",
				Required:            true,
			},
			"role_priority": schema.Int64Attribute{
				MarkdownDescription: "The priority that governs which role will be suggested to use in the web app when requesting access. The availability spec with the highest priority will have its role suggested first in the UI",
				Optional:            true,
			},
		},
		MarkdownDescription: `A specifier to make GCP folders available for selection under a particular Access Workflow`,
	}
}

func (r *GCPFolderAvailabilitiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *GCPFolderAvailabilities

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	input := &configv1alpha1.CreateAvailabilitySpecRequest{
		Role: &entityv1alpha1.EID{
			Type: "GCP::Role",
			Id:   data.Role.ValueString(),
		},
		WorkflowId: data.WorkflowID.ValueString(),
		Target: &entityv1alpha1.EID{
			Type: "Access::Selector",
			Id:   data.FolderSelectorID.ValueString(),
		},
		IdentityDomain: &entityv1alpha1.EID{
			Type: "Google::Workspace::Customer",
			Id:   data.WorkspaceCustomerID.ValueString(),
		},
	}
	if !data.RolePriority.IsNull() {
		priority := data.RolePriority.ValueInt64()
		input.RolePriority = &priority
	}

	res, err := r.client.AvailabilitySpec().CreateAvailabilitySpec(ctx, connect.NewRequest(input))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create GCP Folder Availabilities",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.AvailabilitySpec.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *GCPFolderAvailabilitiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state GCPFolderAvailabilities

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	res, err := r.client.AvailabilitySpec().GetAvailabilitySpec(ctx, connect.NewRequest(&configv1alpha1.GetAvailabilitySpecRequest{
		Id: state.ID.ValueString(),
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read GCP Folder Availabilities",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(res.Msg.AvailabilitySpec.Id)
	state.WorkflowID = types.StringValue(res.Msg.AvailabilitySpec.WorkflowId)
	state.Role = types.StringValue(res.Msg.AvailabilitySpec.Role.Id)
	state.FolderSelectorID = types.StringValue(res.Msg.AvailabilitySpec.Target.Id)

	if res.Msg.AvailabilitySpec.IdentityDomain != nil {
		state.WorkspaceCustomerID = types.StringValue(res.Msg.AvailabilitySpec.IdentityDomain.Id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GCPFolderAvailabilitiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data GCPFolderAvailabilities

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to read plan data into model",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	input := &configv1alpha1.AvailabilitySpec{
		Id: data.ID.ValueString(),
		Role: &entityv1alpha1.EID{
			Type: "GCP::Role",
			Id:   data.Role.ValueString(),
		},
		WorkflowId: data.WorkflowID.ValueString(),
		Target: &entityv1alpha1.EID{
			Type: "Access::Selector",
			Id:   data.FolderSelectorID.ValueString(),
		},
		IdentityDomain: &entityv1alpha1.EID{
			Type: "Google::Workspace::Customer",
			Id:   data.WorkspaceCustomerID.ValueString(),
		},
	}
	if !data.RolePriority.IsNull() {
		priority := data.RolePriority.ValueInt64()
		input.RolePriority = &priority
	}

	res, err := r.client.AvailabilitySpec().UpdateAvailabilitySpec(ctx, connect.NewRequest(&configv1alpha1.UpdateAvailabilitySpecRequest{
		AvailabilitySpec: input,
	}))
	if connectErr, ok := err.(*connect.Error); ok {
		if connectErr.Code() == connect.CodeNotFound {
			resp.Diagnostics.AddError(
				"GCP Folder Availability Not Found",
				"The requested GCP Folder Availability no longer exists. "+
					"It may have been deleted or otherwise removed.\n"+
					"Please create a new Availability.",
			)

			return
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update GCP Folder Availabilities",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.AvailabilitySpec.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GCPFolderAvailabilitiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *GCPFolderAvailabilities

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete GCP Folder Availabilities",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.AvailabilitySpec().DeleteAvailabilitySpec(ctx, connect.NewRequest(&configv1alpha1.DeleteAvailabilitySpecRequest{
		Id: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete GCP Folder Availabilities",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *GCPFolderAvailabilitiesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

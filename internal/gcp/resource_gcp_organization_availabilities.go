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

type GCPOrganizationAvailabilities struct {
	ID                  types.String `tfsdk:"id"`
	WorkflowID          types.String `tfsdk:"workflow_id"`
	SelectorID          types.String `tfsdk:"gcp_organization_selector_id"`
	RoleID              types.String `tfsdk:"gcp_role"`
	WorkspaceCustomerID types.String `tfsdk:"google_workspace_customer_id"`
	Priority            types.Int64  `tfsdk:"priority"`
}

type GCPOrganizationAvailabilitiesResource struct {
	client *configsvc.Client
}

var (
	_ resource.Resource                = &GCPOrganizationAvailabilitiesResource{}
	_ resource.ResourceWithConfigure   = &GCPOrganizationAvailabilitiesResource{}
	_ resource.ResourceWithImportState = &GCPOrganizationAvailabilitiesResource{}
)

// Metadata returns the data source type name.
func (r *GCPOrganizationAvailabilitiesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_organization_availabilities"
}

// Configure adds the provider configured client to the data source.
func (r *GCPOrganizationAvailabilitiesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *GCPOrganizationAvailabilitiesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: "A specifier to make GCP Organization roles available for selection under a particular Access Workflow",
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

			"gcp_organization_selector_id": schema.StringAttribute{
				MarkdownDescription: "The target to make available. Should be a Selector entity.",
				Required:            true,
			},

			"google_workspace_customer_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Google Workspace customer associated with the projects",
				Required:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority that governs which role will be suggested to use in the web app when requesting access. The availability spec with the highest priority will have its role suggested first in the UI",
				Optional:            true,
			},
		},
		MarkdownDescription: `A specifier to make GCP Organization roles available for selection under a particular Access Workflow`,
	}
}

func (r *GCPOrganizationAvailabilitiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *GCPOrganizationAvailabilities

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
			Id:   data.RoleID.ValueString(),
		},
		WorkflowId: data.WorkflowID.ValueString(),
		Target: &entityv1alpha1.EID{
			Type: "Access::Selector",
			Id:   data.SelectorID.ValueString(),
		},
		IdentityDomain: &entityv1alpha1.EID{
			Type: "Google::Workspace::Customer",
			Id:   data.WorkspaceCustomerID.ValueString(),
		},
	}
	if !data.Priority.IsNull() {
		priority := data.Priority.ValueInt64()
		input.Priority = &priority
	}

	res, err := r.client.AvailabilitySpec().CreateAvailabilitySpec(ctx, connect.NewRequest(input))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create GCP Organization Availabilities",
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
func (r *GCPOrganizationAvailabilitiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state GCPOrganizationAvailabilities

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
			"Failed to read GCP Organization Availabilities",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(res.Msg.AvailabilitySpec.Id)
	state.WorkflowID = types.StringValue(res.Msg.AvailabilitySpec.WorkflowId)
	state.RoleID = types.StringValue(res.Msg.AvailabilitySpec.Role.Id)
	state.SelectorID = types.StringValue(res.Msg.AvailabilitySpec.Target.Id)

	if res.Msg.AvailabilitySpec.IdentityDomain != nil {
		state.WorkspaceCustomerID = types.StringValue(res.Msg.AvailabilitySpec.IdentityDomain.Id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GCPOrganizationAvailabilitiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data GCPOrganizationAvailabilities

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
			Id:   data.RoleID.ValueString(),
		},
		WorkflowId: data.WorkflowID.ValueString(),
		Target: &entityv1alpha1.EID{
			Type: "Access::Selector",
			Id:   data.SelectorID.ValueString(),
		},
		IdentityDomain: &entityv1alpha1.EID{
			Type: "Google::Workspace::Customer",
			Id:   data.WorkspaceCustomerID.ValueString(),
		},
	}
	if !data.Priority.IsNull() {
		priority := data.Priority.ValueInt64()
		input.Priority = &priority
	}

	res, err := r.client.AvailabilitySpec().UpdateAvailabilitySpec(ctx, connect.NewRequest(&configv1alpha1.UpdateAvailabilitySpecRequest{
		AvailabilitySpec: input,
	}))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update GCP Organization Availabilities",
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

func (r *GCPOrganizationAvailabilitiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *GCPOrganizationAvailabilities

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete GCP Organization Availabilities",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.AvailabilitySpec().DeleteAvailabilitySpec(ctx, connect.NewRequest(&configv1alpha1.DeleteAvailabilitySpecRequest{
		Id: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete GCP Organization Availabilities",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *GCPOrganizationAvailabilitiesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package gcp

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	"github.com/common-fate/sdk/service/control/config/gcprolegroup"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GCPRoleGroupModel struct {
	ID      types.String   `tfsdk:"id"`
	RoleIDs []types.String `tfsdk:"role_ids"`
	Name    types.String   `tfsdk:"name"`
	OrgID   types.String   `tfsdk:"gcp_organization_id"`
}

type GCPRoleGroupResource struct {
	client configv1alpha1connect.GCPRoleGroupServiceClient
}

var (
	_ resource.Resource                = &GCPRoleGroupResource{}
	_ resource.ResourceWithConfigure   = &GCPRoleGroupResource{}
	_ resource.ResourceWithImportState = &GCPRoleGroupResource{}
)

// Metadata returns the data source type name.
func (r *GCPRoleGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_role_group"
}

// Configure adds the provider configured client to the data source.
func (r *GCPRoleGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = gcprolegroup.NewFromConfig(cfg)
}

// GetSchema defines the schema for the data source.
func (r *GCPRoleGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Defines GCP role group resource`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The GCP role group resource ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_ids": schema.ListAttribute{
				MarkdownDescription: "The list of GCP role IDs to include in the role group",
				Required:            true,
				ElementType:         types.StringType,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The unique name of the selector. Call this something memorable and relevant to the resources being selected. For example: `prod-data-eng`",
				Optional:            true,
			},

			"gcp_organization_id": schema.StringAttribute{
				MarkdownDescription: "The GCP organization ID",
				Required:            true,
			},
		},
		MarkdownDescription: `Defines GCP role group resource`,
	}
}

func (r *GCPRoleGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *GCPRoleGroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	createGCPRoleGroup := &configv1alpha1.CreateGCPRoleGroupRequest{
		Name:           data.Name.ValueString(),
		OrganizationId: data.OrgID.ValueString(),
	}

	for _, r := range data.RoleIDs {
		createGCPRoleGroup.RoleIds = append(createGCPRoleGroup.RoleIds, r.ValueString())
	}

	res, err := r.client.CreateGCPRoleGroup(ctx, connect.NewRequest(createGCPRoleGroup))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: GCPRoleGroup",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.RoleGroup.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *GCPRoleGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state GCPRoleGroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	res, err := r.client.GetGCPRoleGroup(ctx, connect.NewRequest(&configv1alpha1.GetGCPRoleGroupRequest{
		Id: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read GCPRoleGroup",
			err.Error(),
		)
		return
	}

	// refresh state
	state = GCPRoleGroupModel{
		ID:    types.StringValue(res.Msg.RoleGroup.Id),
		Name:  types.StringValue(res.Msg.RoleGroup.Name),
		OrgID: types.StringValue(res.Msg.RoleGroup.OrganizationId),
	}

	for _, r := range res.Msg.RoleGroup.RoleIds {
		state.RoleIDs = append(state.RoleIDs, types.StringValue(r))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GCPRoleGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data GCPRoleGroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	updateGCPRoleGroup := &configv1alpha1.UpdateGCPRoleGroupRequest{
		RoleGroup: &configv1alpha1.GCPRoleGroup{
			Id:             data.ID.ValueString(),
			Name:           data.Name.ValueString(),
			OrganizationId: data.OrgID.ValueString()},
	}

	for _, roleID := range data.RoleIDs {
		updateGCPRoleGroup.RoleGroup.RoleIds = append(updateGCPRoleGroup.RoleGroup.RoleIds, roleID.ValueString())
	}

	res, err := r.client.UpdateGCPRoleGroup(ctx, connect.NewRequest(updateGCPRoleGroup))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",

			"JSON Error: "+err.Error(),
		)

		return

	}

	data.ID = types.StringValue(res.Msg.RoleGroup.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GCPRoleGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *GCPRoleGroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteGCPRoleGroup(ctx, connect.NewRequest(&configv1alpha1.DeleteGCPRoleGroupRequest{
		Id: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Resource", err.Error(),
		)

		return
	}
}

func (r *GCPRoleGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package datastax

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/common-fate/grab"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	entityv1alpha1 "github.com/common-fate/sdk/gen/commonfate/entity/v1alpha1"
	"github.com/common-fate/sdk/service/control/configsvc"
	"github.com/common-fate/terraform-provider-commonfate/pkg/diags"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DataStaxOrganizationSelector struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	OrgID types.String `tfsdk:"datastax_organization_id"`
}

func (s DataStaxOrganizationSelector) ToAPI() *configv1alpha1.Selector {
	return &configv1alpha1.Selector{
		Id:           s.ID.ValueString(),
		Name:         s.Name.ValueString(),
		ResourceType: "DataStax::Organization",
		BelongingTo: &entityv1alpha1.EID{
			Type: "DataStax::Organization",
			Id:   s.OrgID.ValueString(),
		},
		When: "true",
	}
}

// AccessRuleResource is the data source implementation.
type DataStaxOrganizationSelectorResource struct {
	client *configsvc.Client
}

var (
	_ resource.Resource                = &DataStaxOrganizationSelectorResource{}
	_ resource.ResourceWithConfigure   = &DataStaxOrganizationSelectorResource{}
	_ resource.ResourceWithImportState = &DataStaxOrganizationSelectorResource{}
)

// Metadata returns the data source type name.
func (r *DataStaxOrganizationSelectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datastax_organization_selector"
}

// Configure adds the provider configured client to the data source.
func (r *DataStaxOrganizationSelectorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *DataStaxOrganizationSelectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: "A Selector to match a DataStax organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the selector",
				Required:            true,
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "The unique name of the selector. Call this something memorable and relevant to the resources being selected. For example: `prod-data-eng`",
				Optional:            true,
			},

			"datastax_organization_id": schema.StringAttribute{
				MarkdownDescription: "The DataStax organization ID",
				Required:            true,
			},
		},
		MarkdownDescription: `A Selector to match a DataStax organization.`,
	}
}

func (r *DataStaxOrganizationSelectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *DataStaxOrganizationSelector

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.Selector().CreateSelector(ctx, connect.NewRequest(&configv1alpha1.CreateSelectorRequest{
		Selector: data.ToAPI(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: Access Selector",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	diags.ToTerraform(res.Msg.Diagnostics, &resp.Diagnostics)

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Selector.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *DataStaxOrganizationSelectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state DataStaxOrganizationSelector

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	res, err := r.client.Selector().GetSelector(ctx, connect.NewRequest(&configv1alpha1.GetSelectorRequest{
		Id: state.ID.ValueString(),
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read Selector",
			err.Error(),
		)
		return
	}

	state.Name = types.StringPointerValue(grab.If(res.Msg.Selector.Name == "", nil, &res.Msg.Selector.Name))
	state.OrgID = types.StringValue(res.Msg.Selector.BelongingTo.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DataStaxOrganizationSelectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data DataStaxOrganizationSelector

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.Selector().UpdateSelector(ctx, connect.NewRequest(&configv1alpha1.UpdateSelectorRequest{
		Selector: data.ToAPI(),
	}))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: Access Selector",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	diags.ToTerraform(res.Msg.Diagnostics, &resp.Diagnostics)

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Selector.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataStaxOrganizationSelectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *DataStaxOrganizationSelector

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource deletion response.",
		)

		return
	}

	_, err := r.client.Selector().DeleteSelector(ctx, connect.NewRequest(&configv1alpha1.DeleteSelectorRequest{
		Id: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource deletion response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *DataStaxOrganizationSelectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

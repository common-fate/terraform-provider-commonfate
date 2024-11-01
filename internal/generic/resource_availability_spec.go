package generic

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	"github.com/common-fate/sdk/service/control/configsvc"
	"github.com/common-fate/terraform-provider-commonfate/pkg/eid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AvailabilitySpec struct {
	ID             types.String `tfsdk:"id"`
	WorkflowID     types.String `tfsdk:"workflow_id"`
	Role           eid.EID      `tfsdk:"role"`
	Target         eid.EID      `tfsdk:"target"`
	IdentityDomain *eid.EID     `tfsdk:"identity_domain"`
	RolePriority   types.Int64  `tfsdk:"role_priority"`
}

// AccessRuleResource is the data source implementation.
type AvailabilitySpecResource struct {
	client *configsvc.Client
}

var (
	_ resource.Resource                = &AvailabilitySpecResource{}
	_ resource.ResourceWithConfigure   = &AvailabilitySpecResource{}
	_ resource.ResourceWithImportState = &AvailabilitySpecResource{}
)

// Metadata returns the data source type name.
func (r *AvailabilitySpecResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_availability_spec"
}

// Configure adds the provider configured client to the data source.
func (r *AvailabilitySpecResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AvailabilitySpecResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: "A specifier to make resources available for selection under a particular Access Workflow",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The helpers Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"workflow_id": schema.StringAttribute{
				MarkdownDescription: "The Access Workflow ID",
				Required:            true,
			},

			"role": schema.SingleNestedAttribute{
				MarkdownDescription: "The role to make available",
				Required:            true,
				Attributes:          eid.EIDAttrs,
			},

			"target": schema.SingleNestedAttribute{
				MarkdownDescription: "The target to make available. Should be a Selector entity.",
				Required:            true,
				Attributes:          eid.EIDAttrs,
			},

			"identity_domain": schema.SingleNestedAttribute{
				MarkdownDescription: "The identity domain associated with the integration",
				Optional:            true,
				Attributes:          eid.EIDAttrs,
			},
			"role_priority": schema.Int64Attribute{
				MarkdownDescription: "The priority that governs which role will be suggested to use in the web app when requesting access. The availability spec with the highest priority will have its role suggested first in the UI",
				Optional:            true,
			},
		},
		MarkdownDescription: `A specifier to make resources available for selection under a particular Access Workflow`,
	}
}

func (r *AvailabilitySpecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AvailabilitySpec

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
		Role:       data.Role.ToAPI(),
		WorkflowId: data.WorkflowID.ValueString(),
		Target:     data.Target.ToAPI(),
	}

	if data.IdentityDomain != nil {
		input.IdentityDomain = data.IdentityDomain.ToAPI()
	}
	if !data.RolePriority.IsNull() {
		priority := data.RolePriority.ValueInt64()
		input.RolePriority = &priority
	}

	res, err := r.client.AvailabilitySpec().CreateAvailabilitySpec(ctx, connect.NewRequest(input))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: AvailabilitySpec",
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
func (r *AvailabilitySpecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AvailabilitySpec

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
			"Failed to read Availability Spec",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(res.Msg.AvailabilitySpec.Id)
	state.WorkflowID = types.StringValue(res.Msg.AvailabilitySpec.WorkflowId)
	state.Role = eid.EIDFromAPI(res.Msg.AvailabilitySpec.Role)
	state.Target = eid.EIDFromAPI(res.Msg.AvailabilitySpec.Target)
	state.IdentityDomain = eid.EIDPtrFromAPI(res.Msg.AvailabilitySpec.IdentityDomain)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AvailabilitySpecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AvailabilitySpec

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	input := &configv1alpha1.AvailabilitySpec{
		Id:         data.ID.ValueString(),
		Role:       data.Role.ToAPI(),
		WorkflowId: data.WorkflowID.ValueString(),
		Target:     data.Target.ToAPI(),
	}

	if data.IdentityDomain != nil {
		input.IdentityDomain = data.IdentityDomain.ToAPI()
	}
	if !data.RolePriority.IsNull() {
		priority := data.RolePriority.ValueInt64()
		input.RolePriority = &priority
	}

	res, err := r.client.AvailabilitySpec().UpdateAvailabilitySpec(ctx, connect.NewRequest(&configv1alpha1.UpdateAvailabilitySpecRequest{
		AvailabilitySpec: input,
	}))
	if connectErr, ok := err.(*connect.Error); ok && connectErr.Code() == connect.CodeNotFound {

		resp.Diagnostics.AddError(
			"Resource Availability Not Found",
			"The requested Resource Availability no longer exists. "+
				"It may have been deleted or otherwise removed.\n"+
				"Please create a new Availability.",
		)

		return

	} else if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource Availability Spec",
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

func (r *AvailabilitySpecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AvailabilitySpec

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.AvailabilitySpec().DeleteAvailabilitySpec(ctx, connect.NewRequest(&configv1alpha1.DeleteAvailabilitySpecRequest{
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

func (r *AvailabilitySpecResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

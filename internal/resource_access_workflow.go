package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	accessworkflow_handler "github.com/common-fate/sdk/service/control/config/accessworkflow"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/durationpb"
)

type AccessWorkflowModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Duration types.String `tfsdk:"duration"`
	Priority types.Int64  `tfsdk:"priority"`
}

// AccessRuleResource is the data source implementation.
type AccessWorkflowResource struct {
	client configv1alpha1connect.AccessWorkflowServiceClient
}

var (
	_ resource.Resource                = &AccessWorkflowResource{}
	_ resource.ResourceWithConfigure   = &AccessWorkflowResource{}
	_ resource.ResourceWithImportState = &AccessWorkflowResource{}
)

// Metadata returns the data source type name.
func (r *AccessWorkflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_workflow"
}

// Configure adds the provider configured client to the data source.
func (r *AccessWorkflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := accessworkflow_handler.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *AccessWorkflowResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Approval Workflows are used to describe where to send Approval messages to.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal approval workflow ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A unique name for the workflow so you know how to identify it.",
				Optional:            true,
			},
			"duration": schema.StringAttribute{
				MarkdownDescription: "The duration of the access workflow.",
				Required:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority that governs whether the policy will be used. If a different policy with a higher priority and the same role exists that one will be used over another.",
				Required:            true,
			},
		},
		MarkdownDescription: `Create`,
	}
}

func (r *AccessWorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AccessWorkflowModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}
	duration, err := time.ParseDuration(data.Duration.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: JIT Policy",
			"There was an error handling the policy duration "+
				"JSON Error: "+err.Error(),
		)

		return
	}

	res, err := r.client.CreateAccessWorkflow(ctx, connect.NewRequest(&configv1alpha1.CreateAccessWorkflowRequest{
		Name:           data.Name.ValueString(),
		AccessDuration: durationpb.New(duration),
		Priority:       int32(data.Priority.ValueInt64()),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: Approval Workflow",
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
func (r *AccessWorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AccessWorkflowModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.ReadAccessWorkflow(ctx, connect.NewRequest(&configv1alpha1.ReadAccessWorkflowRequest{
		Id: state.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Approval Workflow",
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AccessWorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AccessWorkflowModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.UpdateAccessWorkflow(ctx, connect.NewRequest(&configv1alpha1.UpdateAccessWorkflowRequest{
		Id: data.ID.ValueString(),

		Name: data.Name.ValueString(),
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

	data.ID = types.StringValue(res.Msg.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessWorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AccessWorkflowModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteAccessWorkflow(ctx, connect.NewRequest(&configv1alpha1.DeleteAccessWorkflowRequest{
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

func (r *AccessWorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

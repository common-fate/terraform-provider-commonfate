package internal

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	slack_alert_handler "github.com/common-fate/sdk/service/control/config/slackalert"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SlackAlertModel struct {
	ID           types.String `tfsdk:"id"`
	WorkflowId   types.String `tfsdk:"workflow_id"`
	SlackChannel types.String `tfsdk:"slack_channel"`
}

// AccessRuleResource is the data source implementation.
type SlackAlertResource struct {
	client configv1alpha1connect.SlackAlertServiceClient
}

var (
	_ resource.Resource                = &SlackAlertResource{}
	_ resource.ResourceWithConfigure   = &SlackAlertResource{}
	_ resource.ResourceWithImportState = &SlackAlertResource{}
)

// Metadata returns the data source type name.
func (r *SlackAlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slack_alert"
}

// Configure adds the provider configured client to the data source.
func (r *SlackAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := slack_alert_handler.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *SlackAlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Creates a access SlackAlert with the specified cedar SlackAlert used that Common Fate will use to use in the approval engine when users request access using Common Fate.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate SlackAlert ID",
				Required:            true,
			},

			"workflow_id": schema.StringAttribute{
				MarkdownDescription: "The Cedar SlackAlert to define permissions as policies in your Common Fate instance.",
				Required:            true,
			},
			"slack_channel": schema.StringAttribute{
				MarkdownDescription: "If slack is connected, it will send notifications to this slack channel. Must be the ID of the channel and not the name. See below on how to find this ID.",
				Required:            true,
			},
		},
		MarkdownDescription: `Creates a access SlackAlert with the specified cedar SlackAlert used that Common Fate will use to use in the approval engine when users request access using Common Fate.`,
	}
}

func (r *SlackAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *SlackAlertModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.CreateSlackAlert(ctx, connect.NewRequest(&configv1alpha1.CreateSlackAlertRequest{
		WorkflowId:   data.WorkflowId.ValueString(),
		SlackChannel: data.SlackChannel.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: Access SlackAlert",
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
func (r *SlackAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state SlackAlertModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.ReadSlackAlert(ctx, connect.NewRequest(&configv1alpha1.ReadSlackAlertRequest{
		Id: state.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read access SlackAlert",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SlackAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data SlackAlertModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.UpdateSlackAlert(ctx, connect.NewRequest(&configv1alpha1.UpdateSlackAlertRequest{
		Id:           data.ID.ValueString(),
		WorkflowId:   data.WorkflowId.ValueString(),
		SlackChannel: data.SlackChannel.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",

			"JSON Error: "+err.Error(),
		)

		return

	}

	data.ID = types.StringValue(res.Msg.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SlackAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *SlackAlertModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	//TODO: call api to remove the identity source
	_, err := r.client.DeleteSlackAlert(ctx, connect.NewRequest(&configv1alpha1.DeleteSlackAlertRequest{
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

func (r *SlackAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

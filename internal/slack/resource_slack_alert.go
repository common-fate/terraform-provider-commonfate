package slack

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/common-fate/grab"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	slack_alert_handler "github.com/common-fate/sdk/service/control/config/slackalert"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/durationpb"
)

type SlackAlertModel struct {
	ID                               types.String `tfsdk:"id"`
	WorkflowID                       types.String `tfsdk:"workflow_id"`
	SlackIntegrationID               types.String `tfsdk:"integration_id"`
	SlackChannelID                   types.String `tfsdk:"slack_channel_id"`
	SlackWorkspaceID                 types.String `tfsdk:"slack_workspace_id"`
	UseWebConsoleForApprovalAction   types.Bool   `tfsdk:"use_web_console_for_approval_action"`
	SendDirectMessagesToApprovers    types.Bool   `tfsdk:"send_direct_message_to_approvers"`
	DisableInteractivityHandlers     types.Bool   `tfsdk:"disable_interactivity_handlers"`
	NotifyExpiryInSeconds            types.Int64  `tfsdk:"notify_expiry_in_seconds"`
	SendAlertForAutoApprovedRequests types.Bool   `tfsdk:"send_alert_for_autoapproved_reqeusts"`
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
		Description: `Links a Slack message being send to a particular channel or workspace based on actions made against a workflow. Slack Integration resource is required before creaing a Slack Alert.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal approval workflow ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow_id": schema.StringAttribute{
				MarkdownDescription: "The Access Workflow ID.",
				Required:            true,
			},
			"integration_id": schema.StringAttribute{
				MarkdownDescription: "The ID for the integration set up to slack.",
				Optional:            true,
			},
			"slack_channel_id": schema.StringAttribute{
				MarkdownDescription: "If Slack is connected, it will send notifications to this slack channel. Must be the ID of the channel and not the name. See below on how to find this ID.",
				Optional:            true,
			},
			"slack_workspace_id": schema.StringAttribute{
				MarkdownDescription: "The Slack Workspace ID. In Slack URLs, such as `https://app.slack.com/client/TXXXXXXX/CXXXXXXX` it is the string beginning with T.",
				Optional:            true,
			},
			"use_web_console_for_approval_action": schema.BoolAttribute{
				MarkdownDescription: "Optionally, configure the access request review buttons to be links to the web console, rather than performing the action in Slack.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"send_direct_message_to_approvers": schema.BoolAttribute{
				MarkdownDescription: "If Slack is connected, it will send notifications to the requesting user. Cannot be used in conjunction with 'slack_channel_id'",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"disable_interactivity_handlers": schema.BoolAttribute{
				MarkdownDescription: "Disables all webhook handlers for the Slack integration.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"notify_expiry_in_seconds": schema.Int64Attribute{
				MarkdownDescription: "The duration before access expiration at which Slack will notify the user about the upcoming expiration.",
				Optional:            true,
			},
			"send_alert_for_autoapproved_reqeusts": schema.BoolAttribute{
				MarkdownDescription: "Enabled sending notifications for all requests, including autoapproved requests.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
		MarkdownDescription: `Links a Slack message being send to a particular channel or workspace based on actions made against a workflow.`,
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

	if !data.SlackChannelID.IsNull() && data.SendDirectMessagesToApprovers.ValueBool() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"Cannot use `slack_channel_id` and `send_direct_message_to_approvers` together.",
		)

		return
	}

	if data.SlackChannelID.IsNull() && !data.SendDirectMessagesToApprovers.ValueBool() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"Must have `slack_channel_id` or `send_direct_message_to_approvers` set.",
		)

		return
	}

	createSlackAlert := &configv1alpha1.CreateSlackAlertRequest{
		WorkflowId:                    data.WorkflowID.ValueString(),
		SlackWorkspaceId:              data.SlackWorkspaceID.ValueString(),
		UseWebConsoleForApproveAction: data.UseWebConsoleForApprovalAction.ValueBool(),
		SendDirectMessagesToApprovers: data.SendDirectMessagesToApprovers.ValueBool(),
		DisableInteractivityHandlers:  data.DisableInteractivityHandlers.ValueBool(),
		SendForAutoApprovedRequests:   data.SendAlertForAutoApprovedRequests.ValueBool(),
	}

	if !data.SlackChannelID.IsNull() {
		createSlackAlert.SlackChannelId = data.SlackChannelID.ValueString()
	}

	if data.SlackIntegrationID.ValueString() != "" {
		createSlackAlert.IntegrationId = data.SlackIntegrationID.ValueStringPointer()
	}

	if !data.NotifyExpiryInSeconds.IsNull() {
		notifyExpiry := time.Second * time.Duration(data.NotifyExpiryInSeconds.ValueInt64())
		createSlackAlert.NotifyExpiryInSeconds = durationpb.New(notifyExpiry)
	}
	res, err := r.client.CreateSlackAlert(ctx, connect.NewRequest(createSlackAlert))

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
	data.ID = types.StringValue(res.Msg.Alert.Id)

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
	res, err := r.client.GetSlackAlert(ctx, connect.NewRequest(&configv1alpha1.GetSlackAlertRequest{
		Id: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read SlackAlert",
			err.Error(),
		)
		return
	}

	//refresh state
	state = SlackAlertModel{
		ID:         types.StringValue(res.Msg.Alert.Id),
		WorkflowID: types.StringValue(res.Msg.Alert.WorkflowId),
		// ensure that if channel ID is not provided that it doesn't show an update is going to happen
		SlackChannelID:                   types.StringPointerValue(grab.If(res.Msg.Alert.SlackChannelId == "", nil, &res.Msg.Alert.SlackChannelId)),
		SlackWorkspaceID:                 types.StringValue(res.Msg.Alert.SlackWorkspaceId),
		SlackIntegrationID:               types.StringPointerValue(res.Msg.Alert.IntegrationId),
		UseWebConsoleForApprovalAction:   types.BoolPointerValue(&res.Msg.Alert.UseWebConsoleForApproveAction),
		SendDirectMessagesToApprovers:    types.BoolPointerValue(&res.Msg.Alert.SendDirectMessagesToApprovers),
		DisableInteractivityHandlers:     types.BoolPointerValue(&res.Msg.Alert.DisableInteractivityHandlers),
		SendAlertForAutoApprovedRequests: types.BoolPointerValue(&res.Msg.Alert.SendForAutoApprovedRequests),
	}

	if res.Msg.Alert.NotifyExpiryInSeconds != nil {
		state.NotifyExpiryInSeconds = types.Int64Value(res.Msg.Alert.NotifyExpiryInSeconds.Seconds)
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

	if !data.SlackChannelID.IsNull() && data.SendDirectMessagesToApprovers.ValueBool() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"Cannot use `slack_channel_id` and `send_direct_message_to_approvers` together.",
		)

		return
	}

	if data.SlackChannelID.IsNull() && !data.SendDirectMessagesToApprovers.ValueBool() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"Must have `slack_channel_id` or `send_direct_message_to_approvers` set.",
		)

		return
	}

	updateSlackAlert := &configv1alpha1.UpdateSlackAlertRequest{
		Alert: &configv1alpha1.SlackAlert{
			Id:                            data.ID.ValueString(),
			WorkflowId:                    data.WorkflowID.ValueString(),
			SlackWorkspaceId:              data.SlackWorkspaceID.ValueString(),
			UseWebConsoleForApproveAction: data.UseWebConsoleForApprovalAction.ValueBool(),
			SendDirectMessagesToApprovers: data.SendDirectMessagesToApprovers.ValueBool(),
			DisableInteractivityHandlers:  data.DisableInteractivityHandlers.ValueBool(),
			SendForAutoApprovedRequests:   data.SendAlertForAutoApprovedRequests.ValueBool(),
		},
	}

	if !data.SlackChannelID.IsNull() {
		updateSlackAlert.Alert.SlackChannelId = data.SlackChannelID.ValueString()
	}

	if data.SlackIntegrationID.ValueString() != "" {
		updateSlackAlert.Alert.IntegrationId = data.SlackIntegrationID.ValueStringPointer()
	}

	if !data.NotifyExpiryInSeconds.IsNull() {
		notifyExpiry := time.Second * time.Duration(data.NotifyExpiryInSeconds.ValueInt64())
		updateSlackAlert.Alert.NotifyExpiryInSeconds = durationpb.New(notifyExpiry)
	}

	res, err := r.client.UpdateSlackAlert(ctx, connect.NewRequest(updateSlackAlert))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",

			"JSON Error: "+err.Error(),
		)

		return

	}

	data.ID = types.StringValue(res.Msg.Alert.Id)

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

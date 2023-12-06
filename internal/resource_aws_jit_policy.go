package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	jit_handler "github.com/common-fate/sdk/service/control/config/jitworkflow"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/durationpb"
)

type AWSJITPolicy struct {
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Priority           types.Int64    `tfsdk:"priority"`
	Duration           types.String   `tfsdk:"duration"`
	NotifySlackChannel types.String   `tfsdk:"notify_slack_channel"`
	Role               types.String   `tfsdk:"role"`
	MatchAccountIds    []types.String `tfsdk:"match_account_ids"`
	MatchOrgUnits      []types.String `tfsdk:"match_org_units"`
}

// AccessRuleResource is the data source implementation.
type AWSJITPolicyResource struct {
	client configv1alpha1connect.JITWorkflowServiceClient
}

var (
	_ resource.Resource                = &AWSJITPolicyResource{}
	_ resource.ResourceWithConfigure   = &AWSJITPolicyResource{}
	_ resource.ResourceWithImportState = &AWSJITPolicyResource{}
)

// Metadata returns the data source type name.
func (r *AWSJITPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_jit_policy"
}

// Configure adds the provider configured client to the data source.
func (r *AWSJITPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := jit_handler.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *AWSJITPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Used to create a Just-In-Time (JIT) workflows for Google Cloud Platform (AWS). Common Fate's policy engine will use the created workflows and use the best fit for each individual access request made.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority that governs whether the policy will be used. If a different policy with a higher priority and the same role exists that one will be used over another.",
				Required:            true,
			},
			"duration": schema.StringAttribute{
				MarkdownDescription: "The duration of the access that will be applied for the user requesting the resource.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The unique name of the workflow. Call this something memorable and relevant to the policy being created. For example: `prod-access`",
				Optional:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The target role that will be applied to the resource.",
				Required:            true,
			},
			"notify_slack_channel": schema.StringAttribute{
				MarkdownDescription: "If slack is connected, it will send notifications to this slack channel. Must be the ID of the channel and not the name. See below on how to find this ID.",
				Optional:            true,
			},
			"match_account_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A list of AWS Account ID's that you want to assign access to. Common Fate will automatically match these projects with the role attached to the policy.",
				Optional:            true,
				//todo: possibly add a validator here
			},
			"match_org_units": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A list of AWS org units that you want to give access to. Will include the org unit and its child Accounts. Common Fate will automatically match these accounts with the role attached to the policy.",
				Optional:            true,
				//todo: possibly add a validator here
			},
		},
		MarkdownDescription: `Used to create a Just-In-Time (JIT) workflows for AWS. Common Fate's policy engine will use the created workflows and use the best fit for each individual access request made.`,
	}
}

func (r *AWSJITPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AWSJITPolicy

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

	projectFilters := []string{}
	for _, projectFilter := range data.MatchAccountIds {
		projectFilters = append(projectFilters, projectFilter.ValueString())
	}

	folderFilters := []string{}
	for _, folderFilter := range data.MatchOrgUnits {
		folderFilters = append(folderFilters, folderFilter.ValueString())
	}

	res, err := r.client.CreateJITWorkflow(ctx, connect.NewRequest(&configv1alpha1.CreateJITWorkflowRequest{

		Priority:       data.Priority.ValueInt64(),
		Name:           data.Name.ValueString(),
		AccessDuration: durationpb.New(duration),
		Filters: []*configv1alpha1.Filter{
			{
				Filter: &configv1alpha1.Filter_AwsAccount{
					AwsAccount: &configv1alpha1.AWSAccountFilter{
						MatchAccountIds:         projectFilters,
						MatchAccountsInOrgUnits: folderFilters,
					},
				},
			},
		},
		Alerts: []*configv1alpha1.Alert{
			{
				Alert: &configv1alpha1.Alert_SlackChannel{
					SlackChannel: &configv1alpha1.SlackChannelAlert{
						ChannelId: data.NotifySlackChannel.ValueString(),
					},
				},
			},
		},
		Role: data.Role.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: JIT Policy",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Workflow.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AWSJITPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AWSJITPolicy

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.ReadJITWorkflow(ctx, connect.NewRequest(&configv1alpha1.GetJITWorkflowRequest{
		Id: state.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read access policy",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AWSJITPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AWSJITPolicy

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

	projectFilters := []string{}
	for _, projectFilter := range data.MatchAccountIds {
		projectFilters = append(projectFilters, projectFilter.ValueString())
	}

	folderFilters := []string{}
	for _, folderFilter := range data.MatchOrgUnits {
		folderFilters = append(folderFilters, folderFilter.ValueString())
	}

	res, err := r.client.UpdateJITWorkflow(ctx, connect.NewRequest(&configv1alpha1.UpdateJITWorkflowRequest{
		Workflow: &configv1alpha1.JITWorkflow{
			Id:             data.ID.ValueString(),
			Priority:       data.Priority.ValueInt64(),
			Name:           data.Name.ValueString(),
			AccessDuration: durationpb.New(duration),
			Filters: []*configv1alpha1.Filter{
				{
					Filter: &configv1alpha1.Filter_AwsAccount{
						AwsAccount: &configv1alpha1.AWSAccountFilter{
							MatchAccountIds:         projectFilters,
							MatchAccountsInOrgUnits: folderFilters,
						},
					},
				},
			},
			Alerts: []*configv1alpha1.Alert{
				{
					Alert: &configv1alpha1.Alert_SlackChannel{
						SlackChannel: &configv1alpha1.SlackChannelAlert{
							ChannelId: data.NotifySlackChannel.ValueString(),
						},
					},
				},
			},
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: JIT Policy",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Workflow.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSJITPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AWSJITPolicy

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
	_, err := r.client.DeleteJITWorkflow(ctx, connect.NewRequest(&configv1alpha1.DeleteJITWorkflowRequest{
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

func (r *AWSJITPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

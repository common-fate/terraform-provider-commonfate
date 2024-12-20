package access

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	config_client "github.com/common-fate/sdk/config"
	accessv1alpha1 "github.com/common-fate/sdk/gen/commonfate/access/v1alpha1"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	accessworkflow_handler "github.com/common-fate/sdk/service/control/config/accessworkflow"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Validations struct {
	HasReason     types.Bool        `tfsdk:"has_reason"`
	ReasonRegex   []RegexValidation `tfsdk:"reason_regex"`
	HasJiraTicket types.Bool        `tfsdk:"has_jira_ticket"`
}

type RegexValidation struct {
	RegexPattern types.String `tfsdk:"regex_pattern"`
	ErrorMessage types.String `tfsdk:"error_message"`
}

type ExtensionConditions struct {
	MaxExtensions     types.Int64 `tfsdk:"maximum_number_of_extensions"`
	ExtensionDuration types.Int64 `tfsdk:"extension_duration_seconds"`
}
type ApprovalStep struct {
	Name types.String `tfsdk:"name"`
	When types.String `tfsdk:"when"`
}
type AccessWorkflowModel struct {
	ID                        types.String         `tfsdk:"id"`
	Name                      types.String         `tfsdk:"name"`
	AccessDuration            types.Int64          `tfsdk:"access_duration_seconds"`
	TryExtendAfter            types.Int64          `tfsdk:"try_extend_after_seconds"`
	Priority                  types.Int64          `tfsdk:"priority"`
	ActivationExpiry          types.Int64          `tfsdk:"activation_expiry"`
	RequestedToApprovedExpiry types.Int64          `tfsdk:"requested_to_approved_expiry"`
	RequestedToActivateExpiry types.Int64          `tfsdk:"requested_to_activate_expiry"`
	DefaultDuration           types.Int64          `tfsdk:"default_duration_seconds"`
	Validation                *Validations         `tfsdk:"validation"`
	ExtensionConditions       *ExtensionConditions `tfsdk:"extension_conditions"`
	ApprovalSteps             []ApprovalStep       `tfsdk:"approval_steps"`
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
		Description: `Access Workflows are used to describe how long access should be applied. Created Workflows can be referenced in other resources created.
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
			"access_duration_seconds": schema.Int64Attribute{
				MarkdownDescription: "The maximum allowable duration for the access workflow",
				Required:            true,
			},
			"try_extend_after_seconds": schema.Int64Attribute{
				MarkdownDescription: "The amount of time after access is activated that extending access can be attempted. As a starting point we recommend setting this to half of the `access_duration_seconds`.",
				Optional:            true,
				DeprecationMessage:  "This field is no longer supported. Use extension_conditions to configure access workflow extensions",
				Default:             int64default.StaticInt64(0),
				Computed:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority that governs whether the policy will be used. If a different policy with a higher priority and the same role exists that one will be used over another.",
				Optional:            true,
			},
			"activation_expiry": schema.Int64Attribute{
				MarkdownDescription: "The amount of time after access is approved to be activated before the request will be expired",
				Optional:            true,
			},
			"requested_to_approved_expiry": schema.Int64Attribute{
				MarkdownDescription: "The amount of time after a request is made and approved before the request will be expired",
				Optional:            true,
			},
			"requested_to_activate_expiry": schema.Int64Attribute{
				MarkdownDescription: "The amount of time after a request is made and activated before the request will be expired",
				Optional:            true,
			},
			"default_duration_seconds": schema.Int64Attribute{
				MarkdownDescription: "The default duration of the access workflow",
				Optional:            true,
			},
			"validation": schema.SingleNestedAttribute{
				MarkdownDescription: "Validation requirements to be set with this workflow",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"has_reason": schema.BoolAttribute{
						MarkdownDescription: "Whether a reason is required for this workflow",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"reason_regex": schema.ListNestedAttribute{
						MarkdownDescription: "Regex validation requirements for the reason",
						Optional:            true,

						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"regex_pattern": schema.StringAttribute{
									MarkdownDescription: "The regex pattern that the reason should match on.",
									Required:            true,
								},
								"error_message": schema.StringAttribute{
									MarkdownDescription: "The custom error message to show if the reason doesn't match the regex pattern.",
									Required:            true,
								},
							},
						},
					},
					"has_jira_ticket": schema.BoolAttribute{
						MarkdownDescription: "Whether a jira ticket is required for this workflow",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"extension_conditions": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for extending access",
				Optional:            true,

				Attributes: map[string]schema.Attribute{
					"maximum_number_of_extensions": schema.Int64Attribute{
						MarkdownDescription: "The maximum number of allowed extensions (set to 0 to disable extensions). If not set, it defaults to 0.",
						Required:            true,
					},
					"extension_duration_seconds": schema.Int64Attribute{
						MarkdownDescription: "Specifies the duration for each extension. Defaults to the value of access_duration_seconds if not provided.",
						Required:            true,
					},
				},
			},
			"approval_steps": schema.ListNestedAttribute{
				MarkdownDescription: "Define the requirements for grant approval, each step must be completed by a distict principal, steps can be completed in any order.",
				Optional:            true,

				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the approval step.",
							Required:            true,
						},
						"when": schema.StringAttribute{
							MarkdownDescription: "The Cedar when expression to evaluate a review for a match.",
							Required:            true,
						},
					},
				},
			},
		},
		MarkdownDescription: `Access Workflows are used to describe how long access should be applied. Created Workflows can be referenced in other resources created.`,
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

	accessDuration := time.Second * time.Duration(data.AccessDuration.ValueInt64())
	tryExtendAfter := time.Second * time.Duration(data.TryExtendAfter.ValueInt64())

	createReq := &configv1alpha1.CreateAccessWorkflowRequest{
		Name:           data.Name.ValueString(),
		AccessDuration: durationpb.New(accessDuration),
		TryExtendAfter: durationpb.New(tryExtendAfter),
		Priority:       int32(data.Priority.ValueInt64()),
	}

	if !data.ActivationExpiry.IsNull() {
		activationExpiry := time.Second * time.Duration(data.ActivationExpiry.ValueInt64())

		createReq.ActivationExpiry = durationpb.New(activationExpiry)
	}

	if !data.RequestedToActivateExpiry.IsNull() {
		RequestedToActivateExpiry := time.Second * time.Duration(data.RequestedToActivateExpiry.ValueInt64())

		createReq.RequestToActiveExpiry = durationpb.New(RequestedToActivateExpiry)
	}

	if !data.RequestedToApprovedExpiry.IsNull() {
		RequestedToApprovedExpiry := time.Second * time.Duration(data.RequestedToApprovedExpiry.ValueInt64())

		createReq.RequestToApproveExpiry = durationpb.New(RequestedToApprovedExpiry)
	}

	if data.Validation != nil {

		var regexValidations []*accessv1alpha1.RegexValidation

		if data.Validation.ReasonRegex != nil {
			for _, r := range data.Validation.ReasonRegex {
				regexValidations = append(regexValidations, &accessv1alpha1.RegexValidation{
					RegexPattern: r.RegexPattern.ValueString(),
					ErrorMessage: r.ErrorMessage.ValueString(),
				})
			}
		}

		createReq.Validation = &configv1alpha1.ValidationConfig{
			HasReason:     data.Validation.HasReason.ValueBool(),
			ReasonRegex:   regexValidations,
			HasJiraTicket: data.Validation.HasJiraTicket.ValueBool(),
		}
	}

	// set default duration to access duration by default
	if !data.DefaultDuration.IsNull() {
		defaultDuration := time.Second * time.Duration(data.DefaultDuration.ValueInt64())
		if defaultDuration > accessDuration {
			resp.Diagnostics.AddError(
				"Invalid Default Duration",
				"The default duration must be less than the maximum access duration. "+
					"Please adjust the Default Duration to be less than Access Duration.\n\n"+
					"Default Duration: "+defaultDuration.String()+", Access Duration: "+accessDuration.String(),
			)
			return
		}
		createReq.DefaultDuration = durationpb.New(defaultDuration)
	}

	if data.ExtensionConditions != nil {
		cond := accessv1alpha1.ExtensionConditions{}
		if !data.ExtensionConditions.ExtensionDuration.IsNull() {
			cond.ExtensionDurationSeconds = durationpb.New(time.Second * time.Duration(data.ExtensionConditions.ExtensionDuration.ValueInt64()))
		}
		if !data.ExtensionConditions.MaxExtensions.IsNull() {
			cond.MaximumNumberOfExtensions = int32(data.ExtensionConditions.MaxExtensions.ValueInt64())
		}
		createReq.ExtensionConditions = &cond
	}

	for _, step := range data.ApprovalSteps {
		createReq.ApprovalSteps = append(createReq.ApprovalSteps, &configv1alpha1.ApprovalStep{
			Name: step.Name.ValueString(),
			When: step.When.ValueString(),
		})
	}

	res, err := r.client.CreateAccessWorkflow(ctx, connect.NewRequest(createReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: Approval Workflow",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// ad, tea := HandleDurationStrings(res.Msg.Workflow.AccessDuration, res.Msg.Workflow.TryExtendAfter)

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Workflow.Id)
	// data.AccessDuration = data.TryExtendAfter
	// data.TryExtendAfter = data.TryExtendAfter

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

	// read the state from the client
	res, err := r.client.GetAccessWorkflow(ctx, connect.NewRequest(&configv1alpha1.GetAccessWorkflowRequest{
		Id: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read Access Workflow",
			err.Error(),
		)
		return
	}

	// refresh state

	state.ID = types.StringValue(res.Msg.Workflow.Id)
	state.Name = types.StringValue(res.Msg.Workflow.Name)
	state.AccessDuration = types.Int64Value(res.Msg.Workflow.AccessDuration.Seconds)
	state.Priority = types.Int64Value(int64(res.Msg.Workflow.Priority))
	state.TryExtendAfter = types.Int64Value(res.Msg.Workflow.TryExtendAfter.Seconds)

	if res.Msg.Workflow.DefaultDuration != nil {
		state.DefaultDuration = types.Int64Value(res.Msg.Workflow.DefaultDuration.Seconds)
	}

	if res.Msg.Workflow.ActivationExpiry != nil {
		state.ActivationExpiry = types.Int64Value(res.Msg.Workflow.ActivationExpiry.Seconds)

	}

	if res.Msg.Workflow.RequestToActiveExpiry != nil {
		state.RequestedToActivateExpiry = types.Int64Value(res.Msg.Workflow.RequestToActiveExpiry.Seconds)

	}

	if res.Msg.Workflow.RequestToApproveExpiry != nil {
		state.RequestedToApprovedExpiry = types.Int64Value(res.Msg.Workflow.RequestToApproveExpiry.Seconds)

	}
	if res.Msg.Workflow.Validation != nil {
		var regexValidations []RegexValidation

		for _, r := range res.Msg.Workflow.Validation.ReasonRegex {
			regexValidations = append(regexValidations, RegexValidation{
				RegexPattern: types.StringValue(r.RegexPattern),
				ErrorMessage: types.StringValue(r.ErrorMessage),
			})
		}

		state.Validation = &Validations{
			HasReason:     types.BoolValue(res.Msg.Workflow.Validation.HasReason),
			ReasonRegex:   regexValidations,
			HasJiraTicket: types.BoolValue(res.Msg.Workflow.Validation.HasJiraTicket),
		}
	}

	if res.Msg.Workflow.ExtensionConditions != nil {
		state.ExtensionConditions = &ExtensionConditions{
			ExtensionDuration: types.Int64Value(res.Msg.Workflow.ExtensionConditions.ExtensionDurationSeconds.Seconds),
			MaxExtensions:     types.Int64Value(int64(res.Msg.Workflow.ExtensionConditions.MaximumNumberOfExtensions)),
		}
	}

	state.ApprovalSteps = nil
	for _, step := range res.Msg.Workflow.ApprovalSteps {
		state.ApprovalSteps = append(state.ApprovalSteps, ApprovalStep{
			Name: types.StringValue(step.Name),
			When: types.StringValue(step.When),
		})
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

	accessDuration := time.Second * time.Duration(data.AccessDuration.ValueInt64())
	tryExtendAfter := time.Second * time.Duration(data.TryExtendAfter.ValueInt64())

	updateReq := &configv1alpha1.UpdateAccessWorkflowRequest{
		Workflow: &configv1alpha1.AccessWorkflow{
			Id:             data.ID.ValueString(),
			Name:           data.Name.ValueString(),
			AccessDuration: durationpb.New(accessDuration),
			TryExtendAfter: durationpb.New(tryExtendAfter),
			Priority:       int32(data.Priority.ValueInt64()),
		},
	}

	if !data.ActivationExpiry.IsNull() {
		activationExpiry := time.Second * time.Duration(data.ActivationExpiry.ValueInt64())

		updateReq.Workflow.ActivationExpiry = durationpb.New(activationExpiry)
	}

	if !data.RequestedToActivateExpiry.IsNull() {
		RequestedToActivateExpiry := time.Second * time.Duration(data.RequestedToActivateExpiry.ValueInt64())

		updateReq.Workflow.RequestToActiveExpiry = durationpb.New(RequestedToActivateExpiry)
	}

	if !data.RequestedToApprovedExpiry.IsNull() {
		RequestedToApprovedExpiry := time.Second * time.Duration(data.RequestedToApprovedExpiry.ValueInt64())

		updateReq.Workflow.RequestToApproveExpiry = durationpb.New(RequestedToApprovedExpiry)
	}

	if data.Validation != nil {
		var regexValidations []*accessv1alpha1.RegexValidation

		for _, r := range data.Validation.ReasonRegex {
			regexValidations = append(regexValidations, &accessv1alpha1.RegexValidation{
				RegexPattern: r.RegexPattern.ValueString(),
				ErrorMessage: r.ErrorMessage.ValueString(),
			})
		}

		updateReq.Workflow.Validation = &configv1alpha1.ValidationConfig{
			HasReason:     data.Validation.HasReason.ValueBool(),
			ReasonRegex:   regexValidations,
			HasJiraTicket: data.Validation.HasJiraTicket.ValueBool(),
		}
	}

	// set default duration to access duration by default
	if !data.DefaultDuration.IsNull() {

		defaultDuration := time.Second * time.Duration(data.DefaultDuration.ValueInt64())
		if defaultDuration > accessDuration {
			resp.Diagnostics.AddError(
				"Invalid Default Duration",
				"The default duration must be less than the maximum access duration. "+
					"Please adjust the Default Duration to be less than Access Duration.\n\n"+
					"Default Duration: "+defaultDuration.String()+", Access Duration: "+accessDuration.String(),
			)
			return
		}
		updateReq.Workflow.DefaultDuration = durationpb.New(defaultDuration)
	}

	if data.ExtensionConditions != nil {
		cond := accessv1alpha1.ExtensionConditions{}
		if !data.ExtensionConditions.ExtensionDuration.IsNull() {
			cond.ExtensionDurationSeconds = durationpb.New(time.Second * time.Duration(data.ExtensionConditions.ExtensionDuration.ValueInt64()))
		}
		if !data.ExtensionConditions.MaxExtensions.IsNull() {
			cond.MaximumNumberOfExtensions = int32(data.ExtensionConditions.MaxExtensions.ValueInt64())
		}
		updateReq.Workflow.ExtensionConditions = &cond
	}

	for _, step := range data.ApprovalSteps {
		updateReq.Workflow.ApprovalSteps = append(updateReq.Workflow.ApprovalSteps, &configv1alpha1.ApprovalStep{
			Name: step.Name.ValueString(),
			When: step.When.ValueString(),
		})
	}

	res, err := r.client.UpdateAccessWorkflow(ctx, connect.NewRequest(updateReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Workflow.Id)
	data.AccessDuration = types.Int64Value(res.Msg.Workflow.AccessDuration.Seconds)
	data.TryExtendAfter = types.Int64Value(res.Msg.Workflow.TryExtendAfter.Seconds)

	if res.Msg.Workflow.ActivationExpiry == nil {
		data.ActivationExpiry = types.Int64Null()
	} else {
		data.ActivationExpiry = types.Int64Value(res.Msg.Workflow.ActivationExpiry.Seconds)

	}

	if res.Msg.Workflow.DefaultDuration == nil {
		data.DefaultDuration = types.Int64Null()
	} else {
		data.DefaultDuration = types.Int64Value(res.Msg.Workflow.DefaultDuration.Seconds)
	}

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

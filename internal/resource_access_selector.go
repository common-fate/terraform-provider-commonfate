package internal

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	access_selector_handler "github.com/common-fate/sdk/service/control/config/accessselector"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TargetType struct {
	Type types.String `tfsdk:"type"`
	Name types.String `tfsdk:"name"`
}

type AccessSelector struct {
	ID           types.String `tfsdk:"id"`
	Role         types.String `tfsdk:"role"`
	Name         types.String `tfsdk:"name"`
	SelectorType types.String `tfsdk:"selector_type"`
	WorkFlowId   types.String `tfsdk:"workflow_id"`
	Targets      []TargetType `tfsdk:"targets"`
}

// AccessRuleResource is the data source implementation.
type AccessSelectorResource struct {
	client configv1alpha1connect.AccessSelectorServiceClient
}

var (
	_ resource.Resource                = &AccessSelectorResource{}
	_ resource.ResourceWithConfigure   = &AccessSelectorResource{}
	_ resource.ResourceWithImportState = &AccessSelectorResource{}
)

// Metadata returns the data source type name.
func (r *AccessSelectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_selector"
}

// Configure adds the provider configured client to the data source.
func (r *AccessSelectorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := access_selector_handler.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *AccessSelectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Used to create a Just-In-Time (JIT) workflows for Google Cloud Platform (GCP). Common Fate's policy engine will use the created workflows and use the best fit for each individual access request made.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The target role that will be applied to the resource.",
				Required:            true,
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "The unique name of the selector. Call this something memorable and relevant to the policy being created. For example: `prod-access`",
				Optional:            true,
			},

			"selector_type": schema.StringAttribute{
				MarkdownDescription: "The type of selector that you want the configured access to target.",
				Required:            true,
			},
			"workflow_id": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The ID of the workflow created, that workflow will be used when creating access requests for the targeted resources.",
				Required:            true,
				//todo: possibly add a validator here
			},
			"targets": schema.ListNestedAttribute{
				Required: true,

				MarkdownDescription: "A list of key value pairs of targeted resources. eg. {Type: 'GCPProject', Name: 'demo-project-1'}",

				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,

							MarkdownDescription: "The type of the targeted resource. Example: `GCPProject`",
						},
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The name of the targeted resource. Example: `demo-project-1`",
						},
					},
				},
			},
		},
		MarkdownDescription: `Used to create a Just-In-Time (JIT) workflows for Google Cloud Platform (GCP). Common Fate's policy engine will use the created workflows and use the best fit for each individual access request made.`,
	}
}

func (r *AccessSelectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AccessSelector

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	targets := []*configv1alpha1.Targets{}

	for _, target := range data.Targets {
		targets = append(targets, &configv1alpha1.Targets{
			Type: target.Type.ValueString(),
			Name: target.Name.ValueString(),
		})
	}

	res, err := r.client.CreateAccessSelector(ctx, connect.NewRequest(&configv1alpha1.CreateAccessSelectorRequest{
		Name:         data.Name.ValueString(),
		WorkflowId:   data.WorkFlowId.ValueString(),
		Role:         data.Role.ValueString(),
		SelectorType: data.SelectorType.ValueString(),
		Targets:      targets,
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

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Selector.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccessSelectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AccessSelector

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.ReadAccessSelector(ctx, connect.NewRequest(&configv1alpha1.GetAccessSelectorRequest{
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

func (r *AccessSelectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AccessSelector

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	targets := []*configv1alpha1.Targets{}

	for _, target := range data.Targets {
		targets = append(targets, &configv1alpha1.Targets{
			Type: target.Type.ValueString(),
			Name: target.Name.ValueString(),
		})
	}

	res, err := r.client.UpdateAccessSelector(ctx, connect.NewRequest(&configv1alpha1.UpdateAccessSelectorRequest{

		Selector: &configv1alpha1.AccessSelector{
			Id:           data.ID.ValueString(),
			Name:         data.Name.ValueString(),
			WorkflowId:   data.WorkFlowId.ValueString(),
			Role:         data.Role.ValueString(),
			SelectorType: data.SelectorType.ValueString(),
			Targets:      targets,
		},
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

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Selector.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessSelectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *AccessSelector

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
	_, err := r.client.DeleteAccessSelector(ctx, connect.NewRequest(&configv1alpha1.DeleteAccessSelectorRequest{
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

func (r *AccessSelectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

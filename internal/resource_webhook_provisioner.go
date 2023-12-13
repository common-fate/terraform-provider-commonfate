package internal

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	"github.com/common-fate/sdk/service/control/configsvc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookProvisionerModel struct {
	ID           types.String      `tfsdk:"id"`
	URL          types.String      `tfsdk:"url"`
	Capabilities []CapabilityModel `tfsdk:"capabilities"`
}

type CapabilityModel struct {
	TargetType  types.String `tfsdk:"target_type"`
	RoleType    types.String `tfsdk:"role_type"`
	BelongingTo EID          `tfsdk:"belonging_to"`
}

// AccessRuleResource is the data source implementation.
type WebhookProvisionerResource struct {
	client *configsvc.Client
}

var (
	_ resource.Resource                = &WebhookProvisionerResource{}
	_ resource.ResourceWithConfigure   = &WebhookProvisionerResource{}
	_ resource.ResourceWithImportState = &WebhookProvisionerResource{}
)

// Metadata returns the data source type name.
func (r *WebhookProvisionerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_provisioner"
}

// Configure adds the provider configured client to the data source.
func (r *WebhookProvisionerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *WebhookProvisionerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers a provisioner with a webhook URL to provision access.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The resource ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The webhook URL.",
				Optional:            true,
			},
			"capabilities": schema.ListNestedAttribute{
				MarkdownDescription: "The resources and integrations that this provisioner supports.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"target_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The type of target such as `GCP::Project` or `AWS::Account`",
						},
						"role_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The type of target such as `GCP::Project` or `AWS::Account`",
						},
						"belonging_to": schema.SingleNestedAttribute{
							Attributes: EIDAttrs,
							Required:   true,
						},
					},
				},
			},
		},
		MarkdownDescription: `Registers a provisioner with a webhook URL to provision access.`,
	}
}

func (r *WebhookProvisionerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *WebhookProvisionerModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	var capabilities []*configv1alpha1.Capability

	for _, c := range data.Capabilities {
		capabilities = append(capabilities, &configv1alpha1.Capability{
			TargetType:  c.TargetType.ValueString(),
			RoleType:    c.RoleType.ValueString(),
			BelongingTo: c.BelongingTo.ToAPI(),
		})
	}

	res, err := r.client.WebhookProvisioner().CreateWebhookProvisioner(ctx, connect.NewRequest(&configv1alpha1.CreateWebhookProvisionerRequest{
		Url:          data.URL.ValueString(),
		Capabilities: capabilities,
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
	data.ID = types.StringValue(res.Msg.WebhookProvisioner.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *WebhookProvisionerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state WebhookProvisionerModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	_, err := r.client.WebhookProvisioner().GetWebhookProvisioner(ctx, connect.NewRequest(&configv1alpha1.GetWebhookProvisionerRequest{
		Id: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read Webhook Provisioner",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebhookProvisionerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data WebhookProvisionerModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	var capabilities []*configv1alpha1.Capability

	for _, c := range data.Capabilities {
		capabilities = append(capabilities, &configv1alpha1.Capability{
			TargetType:  c.TargetType.ValueString(),
			RoleType:    c.RoleType.ValueString(),
			BelongingTo: c.BelongingTo.ToAPI(),
		})
	}

	res, err := r.client.WebhookProvisioner().UpdateWebhookProvisioner(ctx, connect.NewRequest(&configv1alpha1.UpdateWebhookProvisionerRequest{
		WebhookProvisioner: &configv1alpha1.WebhookProvisioner{
			Id:           data.ID.ValueString(),
			Url:          data.URL.ValueString(),
			Capabilities: capabilities,
		},
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

	data.ID = types.StringValue(res.Msg.WebhookProvisioner.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookProvisionerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *WebhookProvisionerModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.WebhookProvisioner().DeleteWebhookProvisioner(ctx, connect.NewRequest(&configv1alpha1.DeleteWebhookProvisionerRequest{
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

func (r *WebhookProvisionerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

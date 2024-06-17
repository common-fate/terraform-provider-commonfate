package webhook

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	integrationv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1/integrationv1alpha1connect"
	"github.com/common-fate/sdk/service/control/integration"
	"github.com/common-fate/terraform-provider-commonfate/pkg/diags"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookIntegrationModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	URL                     types.String `tfsdk:"url"`
	SendAuditLogEvents      types.Bool   `tfsdk:"send_audit_log_events"`
	SendAuthorizationEvents types.Bool   `tfsdk:"send_authorization_events"`
	Headers                 []Header     `tfsdk:"headers"`
	FilterForActions        types.Set    `tfsdk:"filter_for_actions"`
}

type Header struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type WebhookIntegrationResource struct {
	client integrationv1alpha1connect.IntegrationServiceClient
}

var (
	_ resource.Resource                = &WebhookIntegrationResource{}
	_ resource.ResourceWithConfigure   = &WebhookIntegrationResource{}
	_ resource.ResourceWithImportState = &WebhookIntegrationResource{}
)

// Metadata returns the data source type name.
func (r *WebhookIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_integration"
}

// Configure adds the provider configured client to the data source.
func (r *WebhookIntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := integration.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
func (r *WebhookIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers a webhook integration`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A name for the integration",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL to dispatch webhook events to",
				Required:            true,
			},
			"send_audit_log_events": schema.BoolAttribute{
				MarkdownDescription: "Set to true to dispatch Audit Log events to the webhook.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"send_authorization_events": schema.BoolAttribute{
				MarkdownDescription: "Set to true to dispatch authorization events to the webhook",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"headers": schema.ListNestedAttribute{
				MarkdownDescription: "HTTP headers to use when sending the webhook event",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "The HTTP header key",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The HTTP header value",
							Required:            true,
						},
					},
				},
			},
			"filter_for_actions": schema.SetAttribute{
				MarkdownDescription: "Filter for event actions to send to the webhook",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
		MarkdownDescription: `Registers a webhook integration`,
	}
}

func (r *WebhookIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *WebhookIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}
	var filterForActions []string
	diag := data.FilterForActions.ElementsAs(ctx, &filterForActions, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	integ := integrationv1alpha1.Webhook{
		Url:                     data.URL.ValueString(),
		SendAuditLogEvents:      data.SendAuditLogEvents.ValueBool(),
		SendAuthorizationEvents: data.SendAuthorizationEvents.ValueBool(),
		FilterForActions:        filterForActions,
	}

	for _, h := range data.Headers {
		integ.Headers = append(integ.Headers, &integrationv1alpha1.Header{
			Key:   h.Key.ValueString(),
			Value: h.Value.ValueString(),
		})
	}

	res, err := r.client.CreateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.CreateIntegrationRequest{
		Name: data.Name.ValueString(),
		Config: &integrationv1alpha1.Config{
			Config: &integrationv1alpha1.Config_Webhook{
				Webhook: &integ,
			},
		},
	}))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: Webhook Integration",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	diags.ToTerraform(res.Msg.Integration.Diagnostics, &resp.Diagnostics)

	data.Id = types.StringValue(res.Msg.Integration.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *WebhookIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state WebhookIntegrationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	res, err := r.client.GetIntegration(ctx, connect.NewRequest(&integrationv1alpha1.GetIntegrationRequest{
		Id: state.Id.ValueString(),
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read Webhook Integration",
			err.Error(),
		)
		return
	}

	integ := res.Msg.Integration.Config.GetWebhook()
	if integ == nil {
		resp.Diagnostics.AddError(
			"Returned integration did not contain any Webhook configuration",
			"",
		)
		return
	}

	var headers []Header

	for _, h := range integ.Headers {
		headers = append(headers, Header{
			Key:   types.StringValue(h.Key),
			Value: types.StringValue(h.Value),
		})
	}

	filterForActions, diag := types.SetValueFrom(ctx, types.StringType, integ.FilterForActions)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	state = WebhookIntegrationModel{
		Id:                      types.StringValue(state.Id.ValueString()),
		Name:                    types.StringValue(res.Msg.Integration.Name),
		URL:                     types.StringValue(integ.Url),
		SendAuditLogEvents:      types.BoolValue(integ.SendAuditLogEvents),
		SendAuthorizationEvents: types.BoolValue(integ.SendAuthorizationEvents),
		Headers:                 headers,

		FilterForActions: filterForActions,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebhookIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data WebhookIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to update Webhook Integration",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}
	var filterForActions []string
	diag := data.FilterForActions.ElementsAs(ctx, &filterForActions, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	integ := integrationv1alpha1.Webhook{
		Url:                     data.URL.ValueString(),
		SendAuditLogEvents:      data.SendAuditLogEvents.ValueBool(),
		SendAuthorizationEvents: data.SendAuthorizationEvents.ValueBool(),
		FilterForActions:        filterForActions,
	}

	for _, h := range data.Headers {
		integ.Headers = append(integ.Headers, &integrationv1alpha1.Header{
			Key:   h.Key.ValueString(),
			Value: h.Value.ValueString(),
		})
	}

	res, err := r.client.UpdateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.UpdateIntegrationRequest{
		Integration: &integrationv1alpha1.Integration{
			Id:   data.Id.ValueString(),
			Name: data.Name.ValueString(),
			Config: &integrationv1alpha1.Config{
				Config: &integrationv1alpha1.Config_Webhook{
					Webhook: &integ,
				},
			},
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Webhook Integration",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	diags.ToTerraform(res.Msg.Integration.Diagnostics, &resp.Diagnostics)

	data.Id = types.StringValue(res.Msg.Integration.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *WebhookIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Webhook Integration",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteIntegration(ctx, connect.NewRequest(&integrationv1alpha1.DeleteIntegrationRequest{
		Id: data.Id.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Webhook Integration",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *WebhookIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

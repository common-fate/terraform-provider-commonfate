package datastax

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	integrationv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1"
	"github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1/integrationv1alpha1connect"
	"github.com/common-fate/sdk/service/control/integration"
	"github.com/common-fate/terraform-provider-commonfate/internal/utilities/diags"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DataStaxIntegrationModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	APIKeySecretPath types.String `tfsdk:"api_key_secret_path"`
}

type DataStaxIntegrationResource struct {
	client integrationv1alpha1connect.IntegrationServiceClient
}

var (
	_ resource.Resource                = &DataStaxIntegrationResource{}
	_ resource.ResourceWithConfigure   = &DataStaxIntegrationResource{}
	_ resource.ResourceWithImportState = &DataStaxIntegrationResource{}
)

// Metadata returns the data source type name.
func (r *DataStaxIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datastax_integration"
}

// Configure adds the provider configured client to the data source.
func (r *DataStaxIntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
// schema is based off the governance api
func (r *DataStaxIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers a DataStax integration`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the integration: use a short label which is descriptive of the organization you're connecting to",
				Required:            true,
			},
			"api_key_secret_path": schema.StringAttribute{
				MarkdownDescription: "Path to secret for the DataStax API Key",
				Required:            true,
			},
		},
		MarkdownDescription: `Registers a DataStax integration`,
	}
}

func (r *DataStaxIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *DataStaxIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	res, err := r.client.CreateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.CreateIntegrationRequest{
		Name: data.Name.ValueString(),
		Config: &integrationv1alpha1.Config{
			Config: &integrationv1alpha1.Config_Datastax{
				Datastax: &integrationv1alpha1.DataStax{
					ApiKeySecretPath: data.APIKeySecretPath.ValueString(),
				},
			},
		},
	}))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: DataStax Integration",
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
func (r *DataStaxIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state DataStaxIntegrationModel

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
			"Failed to read DataStax Integration",
			err.Error(),
		)
		return
	}

	integ := res.Msg.Integration.Config.GetDatastax()
	if integ == nil {
		resp.Diagnostics.AddError(
			"Returned integration did not contain any DataStax configuration",
			"",
		)
		return
	}

	state = DataStaxIntegrationModel{
		Id:               types.StringValue(state.Id.ValueString()),
		Name:             types.StringValue(res.Msg.Integration.Name),
		APIKeySecretPath: types.StringValue(integ.ApiKeySecretPath),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DataStaxIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data DataStaxIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to update DataStax Integration",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}

	res, err := r.client.UpdateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.UpdateIntegrationRequest{
		Integration: &integrationv1alpha1.Integration{
			Id:   data.Id.ValueString(),
			Name: data.Name.ValueString(),
			Config: &integrationv1alpha1.Config{
				Config: &integrationv1alpha1.Config_Datastax{
					Datastax: &integrationv1alpha1.DataStax{
						ApiKeySecretPath: data.APIKeySecretPath.ValueString(),
					},
				},
			},
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update DataStax Integration",
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

func (r *DataStaxIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *DataStaxIntegrationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete DataStax Integration",
			"An unexpected error occurred while parsing the response.",
		)

		return
	}

	_, err := r.client.DeleteIntegration(ctx, connect.NewRequest(&integrationv1alpha1.DeleteIntegrationRequest{
		Id: data.Id.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete DataStax Integration",
			"An unexpected error occurred while parsing the response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *DataStaxIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

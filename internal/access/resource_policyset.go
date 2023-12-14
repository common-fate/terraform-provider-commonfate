package access

import (
	"context"
	"fmt"

	config_client "github.com/common-fate/sdk/config"
	"github.com/common-fate/sdk/service/authz/policyset"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicyModel struct {
	ID   types.String `tfsdk:"id"`
	Text types.String `tfsdk:"text"`
}

type PolicySetResource struct {
	client *policyset.Client
}

var (
	_ resource.Resource              = &PolicySetResource{}
	_ resource.ResourceWithConfigure = &PolicySetResource{}
)

// Metadata returns the data source type name.
func (r *PolicySetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policyset"
}

// Configure adds the provider configured client to the data source.
func (r *PolicySetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := policyset.NewFromConfig(cfg)

	r.client = &client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *PolicySetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Creates a Cedar PolicySet used to authorize access decisions in Common Fate.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate policy ID",
				Required:            true,
			},

			"text": schema.StringAttribute{
				MarkdownDescription: "The Cedar policy to define permissions as policies in your Common Fate instance.",
				Required:            true,
			},
		},
		MarkdownDescription: `Creates a Cedar PolicySet used to authorize access decisions in Common Fate.`,
	}
}

func (r *PolicySetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *PolicyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.Create(ctx, policyset.CreateInput{
		PolicySet: policyset.Input{
			ID:   data.ID.ValueString(),
			Text: data.Text.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: Access Policy",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *PolicySetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state PolicyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	got, err := r.client.Get(ctx, policyset.GetInput{
		ID: state.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read PolicySet",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(got.PolicySet.ID)
	state.Text = types.StringValue(got.PolicySet.Text)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PolicySetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data PolicyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	got, err := r.client.Update(ctx, policyset.UpdateInput{
		PolicySet: policyset.Input{
			ID:   data.ID.ValueString(),
			Text: data.Text.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"error updating policy",
			"JSON Error: "+err.Error(),
		)

		return
	}

	data.ID = types.StringValue(got.PolicySet.Id)
	data.Text = types.StringValue(got.PolicySet.Text)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PolicySetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *PolicyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.Delete(ctx, policyset.DeleteInput{
		ID: data.ID.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting policy",
			err.Error(),
		)

		return
	}
}

package internal

import (
	"context"
	"fmt"

	config_client "github.com/common-fate/sdk/config"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	idp_handler "github.com/common-fate/sdk/service/control/config/idp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type identitySourceModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Type      types.String `tfsdk:"type"`
}

// AccessRuleResource is the data source implementation.
type IdentitySourceResource struct {
	client configv1alpha1connect.IDPServiceClient
}

var (
	_ resource.Resource                = &IdentitySourceResource{}
	_ resource.ResourceWithConfigure   = &IdentitySourceResource{}
	_ resource.ResourceWithImportState = &IdentitySourceResource{}
)

// Metadata returns the data source type name.
func (r *IdentitySourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_source"
}

// Configure adds the provider configured client to the data source.
func (r *IdentitySourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := idp_handler.NewFromConfig(cfg)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *IdentitySourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `An identity source is where you want Common Fate to pull users from. This resource will be using SCIM to keep track of users in your IDP.
`,
		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace to pull from in your idp",
				Required:            true,
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Access Rule. The name is what users will see when they look at what they can request access to.  Make this something that has meaning in your context, such as Dev Admin or Prod Admin.",
			},
		},
		MarkdownDescription: `An identity source is where you want Common Fate to pull users from. This resource will be using SCIM to keep track of users in your IDP.`,
	}
}

func (r *IdentitySourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *identitySourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	//TODO:
	//send out request to the api and save the created resource to the state...

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *IdentitySourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state identitySourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client

	// TODO: Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	// if accessRule.HTTPResponse.StatusCode != http.StatusOK {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Refresh Resource",
	// 		"An unexpected error occurred while attempting to refresh resource state: "+string(accessRule.Body))
	// 	return
	// }

	// TODO: Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	// if accessRule.HTTPResponse.StatusCode == http.StatusNotFound {
	// 	resp.State.RemoveResource(ctx)

	// 	return
	// }

	//TODO: if state is updated the update the terraform state for the object.
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IdentitySourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data identitySourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	//TODO: build update request and send it off to the api client

	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Update Resource",
	// 		"An unexpected error occurred while communicating with Common Fate API. "+
	// 			"Please report this issue to the provider developers.\n\n"+
	// 			"JSON Error: "+err.Error(),
	// 	)

	// 	return

	// }
	// if res.StatusCode() >= 300 {
	// 	resp.Diagnostics.AddError(
	// 		"Failed to Create Resource",
	// 		fmt.Sprintf("JSON Error: %s Status Code: %d", string(res.Body), res.StatusCode()),
	// 	)
	// 	return
	// }

	// if res.JSON200 == nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Update Resource",
	// 		"An unexpected error occurred while communicating with Common Fate API. "+
	// 			"res.JSON200 was nil\n\n"+
	// 			"JSON Error: "+string(res.Body),
	// 	)

	// 	return
	// }

	// data.ID = types.StringValue(res.JSON200.ID)
	// data.Status = types.StringValue(string(res.JSON200.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentitySourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *identitySourceModel

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

	// if res.JSON200 == nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to delete Resource",
	// 		"An unexpected error occurred while parsing the resource creation response. "+
	// 			"Please report this issue to the provider developers.\n\n"+
	// 			"JSON Error: "+res.Status(),
	// 	)

	// 	return
	// }
}

func (r *IdentitySourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

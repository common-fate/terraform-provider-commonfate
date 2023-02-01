package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	governance "github.com/common-fate/common-fate/governance/pkg/types"

	cf_types "github.com/common-fate/common-fate/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type accessRuleModel struct {
	Name        types.String   `tfsdk:"name"`
	Approval    *ApprovalModel `tfsdk:"approval"`
	Description types.String   `tfsdk:"description"`
	Groups      []types.String `tfsdk:"groups"`
	ID          types.String   `tfsdk:"id"`
	Status      types.String   `tfsdk:"status"`
	// Version     types.String   `tfsdk:"version"`
	Target         []TargetProviderModel `tfsdk:"target"`
	Duration       types.String          `tfsdk:"duration"`
	TargetProvider types.String          `tfsdk:"target_provider_id"`
}

type TimeConstraintsModel struct {
	MaxDurationSeconds types.Int64 `tfsdk:"maxDurationSeconds"`
}

type ApprovalModel struct {
	Groups *[]types.String `tfsdk:"groups"`
	Users  *[]types.String `tfsdk:"users"`
}

type TargetProviderModel struct {
	Field types.String `tfsdk:"field"`
	Value []string     `tfsdk:"value"`
}

// AccessRuleResource is the data source implementation.
type AccessRuleResource struct {
	client *governance.ClientWithResponses
}

var (
	_ resource.Resource                = &AccessRuleResource{}
	_ resource.ResourceWithConfigure   = &AccessRuleResource{}
	_ resource.ResourceWithImportState = &AccessRuleResource{}
)

// Metadata returns the data source type name.
func (r *AccessRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_rule"
}

// Configure adds the provider configured client to the data source.
func (r *AccessRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*governance.ClientWithResponses)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *AccessRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Access rules control who can request access to what, and the requirements surrounding their requests.

To create an access rule, you must be an administrator.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Access Aule ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Access Rule. The name is what users will see when they look at what they can request access to.  Make this something that has meaning in your context, such as Dev Admin or Prod Admin.",
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Description of the Access Rule. Make this something that has meaning in your context so users understand what this gives access to.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the Access Rule",
			},
			"groups": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "configures who can request this access rule. Access is governed by identity provider groups. For example, you have a group for your “web app developers” and you are creating a rule that grants temporary access to “production web app account”.",
			},

			"target_provider_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Id of the provider. Eg. `aws-sso-v2. Make sure the provider has been configured before attempting to create an access rule for it.",
			},
			"duration": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The duration section allows you to configure constraints around how long your users may request access for. (unit is seconds)",
			},
			"approval": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "configure whether an approval is required when a user requests this rule, this is optional. Can specify individual users or whole groups to request approval from.",

				Attributes: map[string]schema.Attribute{
					"groups": schema.ListAttribute{
						MarkdownDescription: "Groups to be given access to request the rule being created.",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"users": schema.ListAttribute{
						MarkdownDescription: "Users to be given access to request the rule being created.",

						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"target": schema.ListNestedAttribute{
				Required: true,

				MarkdownDescription: "Configuration options for initialising the provider's setup. In the webapp this is the `provider` section when creating an access rule.",

				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Name of the targeted field. Example: `accountId`",
						},
						"value": schema.ListAttribute{
							Required:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Value of the targeted field. Example: `123456789123`",
						},
					},
				},
			},
		},
		MarkdownDescription: `Access rules control who can request access to what, and the requirements surrounding their requests.

To create an access rule, you must be an administrator in Common Fate. See [Creating an admin user](https://docs.commonfate.io/common-fate/deploying-common-fate/deploying/#creating-an-admin-user)
`,
	}
}

func (r *AccessRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *accessRuleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {

		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return

	}

	dur, err := strconv.Atoi(data.Duration.ValueString())

	if err != nil {

		resp.Diagnostics.AddError(
			"failed to configure duration",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return

	}

	createRequest := governance.GovCreateAccessRuleJSONRequestBody{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		TimeConstraints: cf_types.TimeConstraints{MaxDurationSeconds: dur},
	}

	for _, g := range data.Groups {
		createRequest.Groups = append(createRequest.Groups, g.ValueString())
	}

	//if approval not empty

	if data.Approval != nil {
		if len(*data.Approval.Groups) > 0 {
			for _, g := range *data.Approval.Groups {
				createRequest.Approval.Groups = append(createRequest.Approval.Groups, g.ValueString())
			}
		}

		if len(*data.Approval.Users) > 0 {
			for _, u := range *data.Approval.Users {
				createRequest.Approval.Users = append(createRequest.Approval.Users, u.ValueString())
			}
		}
	} else {
		createRequest.Approval = cf_types.ApproverConfig{Groups: []string{}, Users: []string{}}

	}

	args := make(map[string]cf_types.CreateAccessRuleTargetDetailArguments)
	for _, v := range data.Target {

		args[v.Field.ValueString()] = cf_types.CreateAccessRuleTargetDetailArguments{Values: v.Value}
	}

	createRequest.Target = cf_types.CreateAccessRuleTarget{ProviderId: data.TargetProvider.ValueString(), With: cf_types.CreateAccessRuleTarget_With{AdditionalProperties: args}}

	//create the new access model with the client
	res, err := r.client.GovCreateAccessRuleWithResponse(ctx, createRequest)

	if err != nil {

		resp.Diagnostics.AddError(
			"Failed to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	if res.JSON201 == nil {

		resp.Diagnostics.AddError(
			"Failed to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+res.Status(),
		)

		return

	}

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.JSON201.ID)
	data.Status = types.StringValue(string(res.JSON201.Status))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AccessRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state accessRuleModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read access rule

	accessRule, err := r.client.GovGetAccessRuleWithResponse(ctx, state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read access rule",
			err.Error(),
		)
		return
	}

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if accessRule.HTTPResponse.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			"An unexpected error occurred while attempting to refresh resource state: "+string(accessRule.Body))
		return
	}

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if accessRule.HTTPResponse.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)

		return
	}

	var res cf_types.AccessRuleDetail

	err = json.Unmarshal(accessRule.Body, &res)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			"An unexpected error occurred while attempting to refresh resource state. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AccessRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data accessRuleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {

		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return

	}

	dur, err := strconv.Atoi(data.Duration.ValueString())

	if err != nil {

		resp.Diagnostics.AddError(
			"failed to convert time to int",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	updateRequest := governance.GovUpdateAccessRuleJSONRequestBody{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		TimeConstraints: cf_types.TimeConstraints{MaxDurationSeconds: dur},
	}

	for _, g := range data.Groups {
		updateRequest.Groups = append(updateRequest.Groups, g.ValueString())
	}

	if data.Approval != nil {
		if len(*data.Approval.Groups) > 0 {
			for _, g := range *data.Approval.Groups {
				updateRequest.Approval.Groups = append(updateRequest.Approval.Groups, g.ValueString())
			}
		}

		if len(*data.Approval.Users) > 0 {
			for _, u := range *data.Approval.Users {
				updateRequest.Approval.Users = append(updateRequest.Approval.Users, u.ValueString())
			}
		}
	} else {
		updateRequest.Approval = cf_types.ApproverConfig{Groups: []string{}, Users: []string{}}

	}
	args := make(map[string]cf_types.CreateAccessRuleTargetDetailArguments)
	for _, v := range data.Target {

		args[v.Field.ValueString()] = cf_types.CreateAccessRuleTargetDetailArguments{Values: v.Value}
	}

	updateRequest.Target = cf_types.CreateAccessRuleTarget{ProviderId: data.TargetProvider.ValueString(), With: cf_types.CreateAccessRuleTarget_With{AdditionalProperties: args}}
	//update the new access model with the client

	fmt.Println(data.ID.ValueString())
	res, err := r.client.GovUpdateAccessRuleWithResponse(ctx, data.ID.ValueString(), updateRequest)

	if err != nil {

		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+

				"JSON Error: "+res.Status()+" id: "+data.ID.ValueString(),
		)

		return

	}

	if res.JSON200 == nil {

		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+

				"JSON Error: "+res.Status()+" id: "+data.ID.ValueString(),
		)

		return

	}

	data.ID = types.StringValue(res.JSON200.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *AccessRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *accessRuleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {

		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return

	}

	//create the new access model with the client
	res, err := r.client.GovArchiveAccessRuleWithResponse(ctx, data.ID.ValueString())

	if err != nil {

		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	if res.JSON200 == nil {

		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+res.Status(),
		)

		return

	}

}

func (r *AccessRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package commonfate

import (
	"context"
	"fmt"

	"github.com/common-fate/common-fate/governance"

	cf_types "github.com/common-fate/common-fate/pkg/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type accessRuleModel struct {
	Name            types.String         `tfsdk:"name"`
	Approval        ApprovalModel        `tfsdk:"approval"`
	Description     types.String         `tfsdk:"description"`
	Groups          []types.String       `tfsdk:"groups"`
	ID              types.String         `tfsdk:"id"`
	Status          types.String         `tfsdk:"status"`
	Version         types.String         `tfsdk:"version"`
	Target          TargetModel          `tfsdk:"target"`
	TimeConstraints TimeConstraintsModel `tfsdk:"timeConstraints"`
}

type TimeConstraintsModel struct {
	MaxDurationSeconds types.Int64 `tfsdk:"maxDurationSeconds"`
}

type ApprovalModel struct {
	Groups []types.String `tfsdk:"groups"`
	Users  []types.String `tfsdk:"users"`
}

type TargetModel struct {
	Provider TargetProviderModel `tfsdk:"provider"`
}

type TargetProviderModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

// AccessRuleResource is the data source implementation.
type AccessRuleResource struct {
	client *governance.ClientWithResponses
}

// Metadata returns the data source type name.
func (r AccessRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_rule"
}

// Configure adds the provider configured client to the data source.
func (r AccessRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r AccessRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			// "id": schema.StringAttribute{
			// 	MarkdownDescription: "Access Aule ID",
			// 	Required:            true,
			// },
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the Access Rule",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the Access Rule",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the Access Rule",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Version of the access rule",
			},
			"target_provider_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "id of the provider",
			},
			"duration": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "duration of the rule",
			},
			"approval": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"groups": schema.ListAttribute{
							ElementType: types.StringType,
						},
						"users": schema.ListAttribute{
							ElementType: types.StringType,
						},
					},
				},
			},
			"target": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "id of the provider",
						},
						"value": schema.ListAttribute{
							ElementType: types.StringType,
						},
					},
				},
			},
		},
		MarkdownDescription: "Manages a thing.",
	}
}

func (r AccessRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

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

	target := cf_types.CreateAccessRuleTarget{
		ProviderId: data.Target.Provider.ID.ValueString(),
	}

	//create the new access model with the client
	res, err := r.client.GovCreateAccessRuleWithResponse(ctx, governance.GovCreateAccessRuleJSONRequestBody{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Target:      target,
	})

	if err != nil {

		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return

	}

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.JSON201.ID)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r AccessRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	// var state accessRuleModel

	// accessRules, err := r.client.GovListAccessRulesWithResponse(ctx, &governance.GovListAccessRulesParams{})

	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Read HashiCups Coffees",
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// var res governance.ListAccessRulesDetailResponse

	// err = json.Unmarshal(accessRules.Body, &res)
	// if err != nil {
	// 	return
	// }

	// for _, accessRule := range res.AccessRules {
	// 	accessRuleState := accessRuleModel{
	// 		ID:          types.StringValue(accessRule.ID),
	// 		Description: types.StringValue(accessRule.Description),
	// 		Target:      TargetModel{Provider: TargetProviderModel{ID: types.StringValue(accessRule.Target.Provider.Id), Type: types.StringValue(accessRule.Target.Provider.Type)}},
	// 		Status:      types.StringValue(string(accessRule.Status)),
	// 		Version:     types.StringValue(accessRule.ID),
	// 	}

	// 	for _, group := range accessRule.Groups {
	// 		accessRuleState.Groups = append(accessRuleState.Groups, types.StringValue(group))
	// 	}

	// 	for _, apGroup := range accessRule.Approval.Groups {
	// 		accessRuleState.Approval.Groups = append(accessRuleState.Approval.Groups, types.StringValue(apGroup))
	// 	}
	// 	for _, apUser := range accessRule.Approval.Users {
	// 		accessRuleState.Approval.Users = append(accessRuleState.Approval.Users, types.StringValue(apUser))
	// 	}

	// 	state.AccessRules = append(state.AccessRules, accessRuleState)
	// }

	// // Set state
	// diags := resp.State.Set(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
}

func (r AccessRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
}

func (r AccessRuleResource) Delete(ctx context.Context, res resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
}

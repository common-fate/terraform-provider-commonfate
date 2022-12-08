package commonfate

// type accessRuleResourceModel struct {
// 	AccessRules []accessRuleModel `tfsdk:"accessRules"`
// }
// type accessRuleModel struct {
// 	Approval        ApprovalModel        `tfsdk:"approval"`
// 	Description     types.String         `tfsdk:"description"`
// 	Groups          []types.String       `tfsdk:"groups"`
// 	ID              types.String         `tfsdk:"id"`
// 	Status          types.String         `tfsdk:"status"`
// 	Version         types.String         `tfsdk:"version"`
// 	Target          TargetModel          `tfsdk:"target"`
// 	TimeConstraints TimeConstraintsModel `tfsdk:"timeConstraints"`
// }

// type TimeConstraintsModel struct {
// 	MaxDurationSeconds types.Int64 `tfsdk:"maxDurationSeconds"`
// }

// type ApprovalModel struct {
// 	Groups []types.String `tfsdk:"groups"`
// 	Users  []types.String `tfsdk:"users"`
// }

// type TargetModel struct {
// 	Provider TargetProviderModel `tfsdk:"provider"`
// }

// type TargetProviderModel struct {
// 	ID   types.String `tfsdk:"id"`
// 	Type types.String `tfsdk:"type"`
// }

// // AccessRuleResource is the data source implementation.
// type AccessRuleResource struct {
// 	client *governance.ClientWithResponses
// }

// func (d *AccessRuleResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {

// }

// // Configure adds the provider configured client to the data source.
// func (d *AccessRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
// 	if req.ProviderData == nil {
// 		return
// 	}

// 	d.client = req.ProviderData.(*governance.ClientWithResponses)
// }

// // Metadata returns the data source type name.
// func (d *AccessRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
// 	resp.TypeName = req.ProviderTypeName + "_accessrules"
// }

// // GetSchema defines the schema for the data source.
// // schema is based off the governance api
// func (d *AccessRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

// 	resp.Schema = schema.Schema{}

// 	// resp.Schema = schema.Schema{
// 	// 	Attributes: map[string]schema.Attribute{
// 	// 		"accessRules": {
// 	// 			Computed: true,
// 	// 			Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{

// 	// 				"approval": {
// 	// 					Computed: true,
// 	// 					Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
// 	// 						"groups": {
// 	// 							Type:     types.StringType,
// 	// 							Computed: true,
// 	// 						},
// 	// 						"users": {
// 	// 							Type:     types.StringType,
// 	// 							Computed: true,
// 	// 						},
// 	// 					}),
// 	// 				},
// 	// 				"description": {
// 	// 					Type:     types.StringType,
// 	// 					Computed: true,
// 	// 				},
// 	// 				"groups": {
// 	// 					Type:     types.StringType,
// 	// 					Computed: true,
// 	// 				},
// 	// 				"id": {
// 	// 					Type:     types.StringType,
// 	// 					Computed: true,
// 	// 				},
// 	// 				"status": {
// 	// 					Type:     types.StringType,
// 	// 					Computed: true,
// 	// 				},
// 	// 				"version": {
// 	// 					Type:     types.StringType,
// 	// 					Computed: true,
// 	// 				},
// 	// 				"target": {
// 	// 					Computed: true,
// 	// 					Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
// 	// 						"provider": {
// 	// 							Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
// 	// 								"id": {
// 	// 									Type:     types.StringType,
// 	// 									Computed: true,
// 	// 								},
// 	// 								"type": {
// 	// 									Type:     types.StringType,
// 	// 									Computed: true,
// 	// 								},
// 	// 							}),
// 	// 						},
// 	// 					}),
// 	// 				},
// 	// 				"timeConstraints": {
// 	// 					Computed: true,
// 	// 					Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
// 	// 						"maxDurationSeconds": {
// 	// 							Type:     types.Int64Type,
// 	// 							Computed: true,
// 	// 						},
// 	// 					}),
// 	// 				},
// 	// 			}),
// 	// 		},
// 	// 	},
// 	// }
// }

// func (r AccessRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
// }

// // Read refreshes the Terraform state with the latest data.
// func (r AccessRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

// 	var state accessRuleResourceModel

// 	accessRules, err := r.client.GovListAccessRulesWithResponse(ctx, &governance.GovListAccessRulesParams{})

// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Unable to Read HashiCups Coffees",
// 			err.Error(),
// 		)
// 		return
// 	}

// 	var res governance.ListAccessRulesDetailResponse

// 	err = json.Unmarshal(accessRules.Body, &res)
// 	if err != nil {
// 		return
// 	}

// 	for _, accessRule := range res.AccessRules {
// 		accessRuleState := accessRuleModel{
// 			ID:          types.StringValue(accessRule.ID),
// 			Description: types.StringValue(accessRule.Description),
// 			Target:      TargetModel{Provider: TargetProviderModel{ID: types.StringValue(accessRule.Target.Provider.Id), Type: types.StringValue(accessRule.Target.Provider.Type)}},
// 			Status:      types.StringValue(string(accessRule.Status)),
// 			Version:     types.StringValue(accessRule.ID),
// 		}

// 		for _, group := range accessRule.Groups {
// 			accessRuleState.Groups = append(accessRuleState.Groups, types.StringValue(group))
// 		}

// 		for _, apGroup := range accessRule.Approval.Groups {
// 			accessRuleState.Approval.Groups = append(accessRuleState.Approval.Groups, types.StringValue(apGroup))
// 		}
// 		for _, apUser := range accessRule.Approval.Users {
// 			accessRuleState.Approval.Users = append(accessRuleState.Approval.Users, types.StringValue(apUser))
// 		}

// 		state.AccessRules = append(state.AccessRules, accessRuleState)
// 	}

// 	// Set state
// 	diags := resp.State.Set(ctx, &state)
// 	resp.Diagnostics.Append(diags...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// }

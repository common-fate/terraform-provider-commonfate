package access

import (
	"context"

	"github.com/common-fate/terraform-provider-commonfate/pkg/eid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &PolicyDataSource{}

type PolicyDataSource struct{}

type PolicyDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Policies     []Policy     `tfsdk:"policies"`
	PolicyAsText types.String `tfsdk:"policy_as_text"`
}

func (d *PolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (d *PolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{

			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate policy ID",
				Required:            true,
			},

			"policy_as_text": schema.StringAttribute{
				MarkdownDescription: "The converted policy into text for to be used with the policyset resource",
				Computed:            true,
			},

			"policies": schema.SetNestedAttribute{
				MarkdownDescription: "a list of policies to be used in Common Fate",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"effect": schema.StringAttribute{
							MarkdownDescription: "The effect specifies the intent of the policy, to either permit` or forbid any request that matches the scope and conditions specified in the policy",
							Required:            true,
						},
						"advice": schema.StringAttribute{
							MarkdownDescription: "Decorators are annotations added to Cedar policies to provide additional instructions or messages to end users",
							Optional:            true,
						},
						"principal": schema.ObjectAttribute{
							MarkdownDescription: "The principal component specifies the entity seeking access.",
							Optional:            true,
							AttributeTypes:      eid.EIDAttrsForDataSource,
						},
						"principal_is": schema.ObjectAttribute{
							MarkdownDescription: "The principal component specifies the entity seeking access.",
							Optional:            true,
							AttributeTypes:      eid.EIDAttrsForDataSource,
						},
						"principal_in": schema.ListAttribute{
							MarkdownDescription: "a list of principal component's specifying the entities seeking access",
							Optional:            true,
							ElementType: basetypes.ObjectType{
								AttrTypes: eid.EIDAttrsForDataSource,
							},
						},

						"action": schema.ObjectAttribute{
							MarkdownDescription: "Actions define the operations that can be performed within Common Fate.",
							Optional:            true,
							AttributeTypes:      eid.EIDAttrsForDataSource,
						},
						"action_is": schema.ObjectAttribute{
							MarkdownDescription: "Actions define the operations that can be performed within Common Fate.",
							Optional:            true,
							AttributeTypes:      eid.EIDAttrsForDataSource,
						},
						"action_in": schema.ListAttribute{
							MarkdownDescription: "actions_in define a set of operations that can be performed within Common Fate",
							Optional:            true,
							ElementType: basetypes.ObjectType{
								AttrTypes: eid.EIDAttrsForDataSource,
							},
						},
						"resource": schema.ObjectAttribute{
							MarkdownDescription: "The resource component specifies the target or subject of the action. It identifies the entity upon which actions are taken.",
							Optional:            true,
							AttributeTypes:      eid.EIDAttrsForDataSource,
						},
						"resource_is": schema.ObjectAttribute{
							MarkdownDescription: "The resource component specifies the target or subject of the action. It identifies the entity upon which actions are taken.",
							Optional:            true,
							AttributeTypes:      eid.EIDAttrsForDataSource,
						},
						"resource_in": schema.ListAttribute{
							MarkdownDescription: "resource_in component specifies a set of the target or subject of the action. It identifies the entity upon which actions are taken.",
							Optional:            true,
							ElementType: basetypes.ObjectType{
								AttrTypes: eid.EIDAttrsForDataSource,
							},
						},

						"when": schema.SingleNestedAttribute{
							MarkdownDescription: "The when and unless components define additional conditions under which the action is allowed.",
							Optional:            true,

							Attributes: map[string]schema.Attribute{
								"text": schema.StringAttribute{
									MarkdownDescription: "when can be used with the text attribute to define the when clause in plain-text.",
									Required:            true,
								},
								"structured_embedded_expression": schema.SingleNestedAttribute{
									MarkdownDescription: "when can be used with `structured_embedded_expression` to define a more structured when clause.",

									Attributes: map[string]schema.Attribute{
										"resource":   schema.StringAttribute{Required: true},
										"expression": schema.StringAttribute{Required: true},
										"value":      schema.StringAttribute{Required: true},
									},
									Optional: true,
								},
							},
						},
						"unless": schema.SingleNestedAttribute{
							MarkdownDescription: "Specifies the duration for each extension. Defaults to the value of access_duration_seconds if not provided.",
							Optional:            true,

							Attributes: map[string]schema.Attribute{
								"text": schema.StringAttribute{
									MarkdownDescription: "unless can be used with the text attribute to define the when clause in plain-text.",
									Required:            true,
								},
								"structured_embedded_expression": schema.SingleNestedAttribute{
									MarkdownDescription: "unless can be used with `structured_embedded_expression` to define a more structured when clause.",

									Attributes: map[string]schema.Attribute{
										"resource":   schema.StringAttribute{Required: true},
										"expression": schema.StringAttribute{Required: true},
										"value":      schema.StringAttribute{Required: true},
									},
									Optional: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *PolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PolicyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var policyText string
	for _, policy := range data.Policies {

		if (policy.Action != nil && policy.ActionIn != nil) || (policy.ActionIn != nil && policy.Action != nil || policy.ActionIs != nil) || (policy.ActionIs != nil && policy.Action != nil || policy.ActionIn != nil) {
			resp.Diagnostics.AddError(
				"Unable to Create DataSource: Access Policy",
				"Cannot have mulitple values for action condition",
			)

			return
		}

		if (policy.Principal != nil && policy.PrincipalIn != nil) || (policy.PrincipalIn != nil && policy.Principal != nil || policy.PrincipalIs != nil) || (policy.PrincipalIs != nil && policy.Principal != nil || policy.PrincipalIn != nil) {
			resp.Diagnostics.AddError(
				"Unable to Create DataSource: Access Policy",
				"Cannot have mulitple values for Principal condition",
			)

			return
		}

		if (policy.Resource != nil && policy.ResourceIn != nil) || (policy.ResourceIn != nil && policy.Resource != nil || policy.ResourceIs != nil) || (policy.ResourceIs != nil && policy.Resource != nil || policy.ResourceIn != nil) {
			resp.Diagnostics.AddError(
				"Unable to Create DataSource: Access Policy",
				"Cannot have mulitple values for Resource condition",
			)

			return
		}

		currentPolicy, err := PolicyToString(policy)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create DataSource: Access Policy",
				"An unexpected error occurred while parsing the policy. "+
					"Please report this issue to the provider developers.\n\n"+
					"JSON Error: "+err.Error(),
			)

			return
		}
		policyText = policyText + currentPolicy + "\n\n"
	}
	data.PolicyAsText = types.StringValue(policyText)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

package access

import (
	"context"

	"github.com/common-fate/terraform-provider-commonfate/pkg/eid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PolicyDataSource{}

type PolicyDataSource struct{}

type PolicyDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Policies     *[]Policy    `tfsdk:"policies"`
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
				Optional:            true,
			},

			"policies": schema.SetNestedAttribute{
				MarkdownDescription: "Configuration for extending access",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"effect": schema.StringAttribute{
							MarkdownDescription: "The effect on the cedar policy that you want to make. Either 'permit' or 'forbid'",
							Required:            true,
						},
						"principal": schema.SingleNestedAttribute{
							MarkdownDescription: "Specifies the duration for each extension. Defaults to the value of access_duration_seconds if not provided.",
							Optional:            true,

							Attributes: map[string]schema.Attribute{
								"expression": schema.StringAttribute{
									MarkdownDescription: "The Cedar policy to define permissions as policies in your Common Fate instance.",
									Required:            true,
								},
								"resource": schema.SingleNestedAttribute{
									Attributes: eid.EIDAttrsForDataSource,
									Required:   true,
								},
							},
						},
						"action": schema.SingleNestedAttribute{
							MarkdownDescription: "Specifies the duration for each extension. Defaults to the value of access_duration_seconds if not provided.",
							Optional:            true,

							Attributes: map[string]schema.Attribute{
								"expression": schema.StringAttribute{
									MarkdownDescription: "The Cedar policy to define permissions as policies in your Common Fate instance.",
									Required:            true,
								},
								"resource": schema.SingleNestedAttribute{
									Attributes: eid.EIDAttrsForDataSource,
									Required:   true,
								},
							},
						},
						"resource": schema.SingleNestedAttribute{
							MarkdownDescription: "Specifies the duration for each extension. Defaults to the value of access_duration_seconds if not provided.",
							Optional:            true,

							Attributes: map[string]schema.Attribute{
								"expression": schema.StringAttribute{
									MarkdownDescription: "The Cedar policy to define permissions as policies in your Common Fate instance.",
									Required:            true,
								},
								"resource": schema.SingleNestedAttribute{
									Attributes: eid.EIDAttrsForDataSource,
									Required:   true,
								},
							},
						},
						"when": schema.SingleNestedAttribute{
							MarkdownDescription: "Specifies the duration for each extension. Defaults to the value of access_duration_seconds if not provided.",
							Optional:            true,

							Attributes: map[string]schema.Attribute{
								"text": schema.StringAttribute{
									MarkdownDescription: "The Cedar policy to define permissions as policies in your Common Fate instance.",
									Required:            true,
								},
								"structured_embedded_expression": schema.SingleNestedAttribute{
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
									MarkdownDescription: "The Cedar policy to define permissions as policies in your Common Fate instance.",
									Required:            true,
								},
								"structured_embedded_expression": schema.SingleNestedAttribute{
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
	for _, policy := range *data.Policies {
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

package aws

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	"github.com/common-fate/sdk/service/control/configsvc"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSEKSAccessEntryModel struct {
	ID                      types.String                 `tfsdk:"id"`
	Name                    types.String                 `tfsdk:"name"`
	KubernetesGroups        []types.String               `tfsdk:"kubernetes_groups"`
	Tags                    types.Map                    `tfsdk:"tags"`
	ClusterAccessPolicies   []ClusterAccessPolicyModel   `tfsdk:"cluster_access_policies"`
	NamespaceAccessPolicies []NamespaceAccessPolicyModel `tfsdk:"namespace_access_policies"`
}

type ClusterAccessPolicyModel struct {
	PolicyArn types.String `tfsdk:"policy_arn"`
}

type NamespaceAccessPolicyModel struct {
	Namespaces []types.String `tfsdk:"namespaces"`
	PolicyArn  types.String   `tfsdk:"policy_arn"`
}

type AWSEKSAccessEntryResource struct {
	client configv1alpha1connect.EKSAccessEntryTemplateServiceClient
}

func NewAWSEKSAccessEntryResource() resource.Resource {
	return &AWSEKSAccessEntryResource{}
}

var (
	_ resource.Resource                = &AWSEKSAccessEntryResource{}
	_ resource.ResourceWithConfigure   = &AWSEKSAccessEntryResource{}
	_ resource.ResourceWithImportState = &AWSEKSAccessEntryResource{}
)

// Metadata returns the data source type name.
func (r *AWSEKSAccessEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_eks_access_entry_template"
}

// Configure adds the provider configured client to the data source.
func (r *AWSEKSAccessEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = configsvc.NewFromConfig(cfg).AWSEKSAccessEntryTemplate()
}

// GetSchema defines the schema for the data source.
func (r *AWSEKSAccessEntryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers an AWS EKS Access Entry Template with Common Fate`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The EKS Access Entry Template ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the EKS Access Entry Template",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kubernetes_groups": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The Kubernetes groups associated with the template",
				Optional:            true,
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "The tags associated with the template",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"cluster_access_policies": schema.ListNestedAttribute{
				MarkdownDescription: "The cluster access policies associated with the template",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"policy_arn": schema.StringAttribute{
							MarkdownDescription: "The ARN of the cluster access policy",
							Required:            true,
						},
					},
				},
			},
			"namespace_access_policies": schema.ListNestedAttribute{
				MarkdownDescription: "The namespace access policies associated with the template",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"namespaces": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The namespaces associated with the policy",
							Required:            true,
						},
						"policy_arn": schema.StringAttribute{
							MarkdownDescription: "The ARN of the namespace access policy",
							Required:            true,
						},
					},
				},
			},
		},
		MarkdownDescription: `Registers an AWS EKS Access Entry Template with Common Fate`,
	}
}
func (r *AWSEKSAccessEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *AWSEKSAccessEntryModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	// Construct the tags
	var tags []*configv1alpha1.EKSAccessEntryTag
	for key, value := range data.Tags.Elements() {
		tagValue := value.(types.String).ValueString()
		tags = append(tags, &configv1alpha1.EKSAccessEntryTag{Key: key, Value: tagValue})
	}

	// Construct the cluster access policies
	var clusterAccessPolicies []*configv1alpha1.EKSClusterAccessPolicy
	for _, cap := range data.ClusterAccessPolicies {
		clusterAccessPolicies = append(clusterAccessPolicies, &configv1alpha1.EKSClusterAccessPolicy{
			PolicyArn: cap.PolicyArn.ValueString(),
		})
	}

	// Construct the namespace access policies
	var namespaceAccessPolicies []*configv1alpha1.EKSNamespaceAccessPolicy
	for _, nap := range data.NamespaceAccessPolicies {
		var namespaces []string
		for _, ns := range nap.Namespaces {
			namespaces = append(namespaces, ns.ValueString())
		}
		namespaceAccessPolicies = append(namespaceAccessPolicies, &configv1alpha1.EKSNamespaceAccessPolicy{
			Namespaces: namespaces,
			PolicyArn:  nap.PolicyArn.ValueString(),
		})
	}

	kubernetesGroups := make([]string, len(data.KubernetesGroups))
	for i, group := range data.KubernetesGroups {
		kubernetesGroups[i] = group.ValueString()
	}

	res, err := r.client.CreateEKSAccessEntryTemplate(ctx, connect.NewRequest(&configv1alpha1.CreateEKSAccessEntryTemplateRequest{
		Name:                    data.Name.ValueString(),
		KubernetesGroups:        kubernetesGroups,
		Tags:                    tags,
		ClusterAccessPolicies:   clusterAccessPolicies,
		NamespaceAccessPolicies: namespaceAccessPolicies,
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: AWS EKS Access Entry Template",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.AccessEntryTemplate.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AWSEKSAccessEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state AWSEKSAccessEntryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// Read the state from the client
	res, err := r.client.GetEKSAccessEntryTemplate(ctx, connect.NewRequest(&configv1alpha1.GetEKSAccessEntryTemplateRequest{
		Id: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read AWS EKS Access Entry Template",
			err.Error(),
		)
		return
	}

	tags, diags := convertTagsToMap(res.Msg.AccessEntryTemplate.Tags)
	resp.Diagnostics.Append(diags...)

	// Refresh state
	state = AWSEKSAccessEntryModel{
		ID:               types.StringValue(res.Msg.AccessEntryTemplate.Id),
		Name:             types.StringValue(res.Msg.AccessEntryTemplate.Name),
		KubernetesGroups: convertStringSliceToTypesStringSlice(res.Msg.AccessEntryTemplate.KubernetesGroups),
		Tags:             tags,
		ClusterAccessPolicies: convertClusterAccessPolicies(
			res.Msg.AccessEntryTemplate.ClusterAccessPolicies),
		NamespaceAccessPolicies: convertNamespaceAccessPolicies(
			res.Msg.AccessEntryTemplate.NamespaceAccessPolicies),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func convertStringSliceToTypesStringSlice(input []string) []types.String {
	output := make([]types.String, len(input))
	for i, v := range input {
		output[i] = types.StringValue(v)
	}
	return output
}

func convertTagsToMap(tags []*configv1alpha1.EKSAccessEntryTag) (types.Map, diag.Diagnostics) {
	tagMap := make(map[string]attr.Value)
	for _, tag := range tags {
		tagMap[tag.Key] = types.StringValue(tag.Value)
	}
	return types.MapValue(types.StringType, tagMap)
}

func convertClusterAccessPolicies(policies []*configv1alpha1.EKSClusterAccessPolicy) []ClusterAccessPolicyModel {
	output := make([]ClusterAccessPolicyModel, len(policies))
	for i, policy := range policies {
		output[i] = ClusterAccessPolicyModel{
			PolicyArn: types.StringValue(policy.PolicyArn),
		}
	}
	return output
}

func convertNamespaceAccessPolicies(policies []*configv1alpha1.EKSNamespaceAccessPolicy) []NamespaceAccessPolicyModel {
	output := make([]NamespaceAccessPolicyModel, len(policies))
	for i, policy := range policies {
		namespaces := make([]types.String, len(policy.Namespaces))
		for j, ns := range policy.Namespaces {
			namespaces[j] = types.StringValue(ns)
		}
		output[i] = NamespaceAccessPolicyModel{
			Namespaces: namespaces,
			PolicyArn:  types.StringValue(policy.PolicyArn),
		}
	}
	return output
}
func (r *AWSEKSAccessEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data AWSEKSAccessEntryModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}

	// Construct the tags
	var tags []*configv1alpha1.EKSAccessEntryTag
	for key, value := range data.Tags.Elements() {
		tagValue := value.(types.String).ValueString()
		tags = append(tags, &configv1alpha1.EKSAccessEntryTag{Key: key, Value: tagValue})
	}

	// Construct the cluster access policies
	var clusterAccessPolicies []*configv1alpha1.EKSClusterAccessPolicy
	for _, cap := range data.ClusterAccessPolicies {
		clusterAccessPolicies = append(clusterAccessPolicies, &configv1alpha1.EKSClusterAccessPolicy{
			PolicyArn: cap.PolicyArn.ValueString(),
		})
	}

	// Construct the namespace access policies
	var namespaceAccessPolicies []*configv1alpha1.EKSNamespaceAccessPolicy
	for _, nap := range data.NamespaceAccessPolicies {
		var namespaces []string
		for _, ns := range nap.Namespaces {
			namespaces = append(namespaces, ns.ValueString())
		}
		namespaceAccessPolicies = append(namespaceAccessPolicies, &configv1alpha1.EKSNamespaceAccessPolicy{
			Namespaces: namespaces,
			PolicyArn:  nap.PolicyArn.ValueString(),
		})
	}

	kubernetesGroups := make([]string, len(data.KubernetesGroups))
	for i, group := range data.KubernetesGroups {
		kubernetesGroups[i] = group.ValueString()
	}

	res, err := r.client.UpdateEKSAccessEntryTemplate(ctx, connect.NewRequest(&configv1alpha1.UpdateEKSAccessEntryTemplateRequest{
		AccessEntryTemplate: &configv1alpha1.EKSAccessEntryTemplate{
			Id:                      data.ID.ValueString(),
			Name:                    data.Name.ValueString(),
			KubernetesGroups:        kubernetesGroups,
			Tags:                    tags,
			ClusterAccessPolicies:   clusterAccessPolicies,
			NamespaceAccessPolicies: namespaceAccessPolicies,
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource: AWS EKS Access Entry Template",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// Update the state with the new data
	data.ID = types.StringValue(res.Msg.AccessEntryTemplate.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *AWSEKSAccessEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}
	var data *AWSEKSAccessEntryModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteEKSAccessEntryTemplate(ctx, connect.NewRequest(&configv1alpha1.DeleteEKSAccessEntryTemplateRequest{
		Id: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Resource: AWS EKS Access Entry Template",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r *AWSEKSAccessEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

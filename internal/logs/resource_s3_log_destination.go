package logs

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type S3LogDestinationModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	BucketName             types.String `tfsdk:"bucket_name"`
	RoleARN                types.String `tfsdk:"role_arn"`
	KeyTemplate            types.String `tfsdk:"key_template"`
	Compression            types.String `tfsdk:"compression"`
	FilterForActions       types.Set    `tfsdk:"filter_for_actions"`
	BatchDurationInMinutes types.Int64  `tfsdk:"batch_duration_in_minutes"`
	MaximumBatchSize       types.Int64  `tfsdk:"maximum_batch_size"`
}

type S3LogDestinationResource struct {
	client integrationv1alpha1connect.IntegrationServiceClient
}

func NewS3LogDestinationResource() resource.Resource {
	return &S3LogDestinationResource{}
}

var (
	_ resource.Resource                = &S3LogDestinationResource{}
	_ resource.ResourceWithConfigure   = &S3LogDestinationResource{}
	_ resource.ResourceWithImportState = &S3LogDestinationResource{}
)

// Metadata returns the data source type name.
func (r *S3LogDestinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_log_destination"
}

// Configure adds the provider configured client to the data source.
func (r *S3LogDestinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *S3LogDestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Description: `Registers an Amazon S3 log destination`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal Common Fate ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A name for the integration.",
				Required:            true,
			},
			"bucket_name": schema.StringAttribute{
				MarkdownDescription: "The S3 bucket name to send logs to.",
				Required:            true,
			},
			"role_arn": schema.StringAttribute{
				MarkdownDescription: "The role to assume when writing logs to the S3 bucket.",
				Required:            true,
			},
			"key_template": schema.StringAttribute{
				MarkdownDescription: "The template to use when writing events to the bucket. If not provided, will default to '{{`{{ .Year }}`}}/{{`{{ .Month }}`}}/{{`{{ .Day }}`}}/{{`{{ .Hour }}`}}_{{`{{ .Minute }}`}}_{{`{{ .Second }}`}}_{{`{{ .ID }}`}}'.",
				Optional:            true,
			},
			"compression": schema.StringAttribute{
				MarkdownDescription: "An optional compression algorithm to use when writing events. If provided, must be one of ['gzip'].",
				Optional:            true,
			},
			"filter_for_actions": schema.SetAttribute{
				MarkdownDescription: `Optionally filters the logs to be sent to the bucket. For example ['grant.cancelled', 'grant.revoked'] will only write audit log events to the S3 bucket for Grants being cancelled or revoked.

Available actions that you can filter on include:

- grant.requested
- grant.approved
- grant.activated
- grant.provisioned
- grant.provisioning_attempted
- grant.extended
- grant.deprovisioned
- grant.cancelled
- grant.revoked
- grant.provisioning_error
- grant.deprovisioning_error
`,
				ElementType: types.StringType,
				Optional:    true,
			},
			"batch_duration_in_minutes": schema.Int64Attribute{
				MarkdownDescription: "Specifies the frequency to write batches of events to the S3 bucket in minutes. Must be 5 or greater.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(5),
			},
			"maximum_batch_size": schema.Int64Attribute{
				MarkdownDescription: "Specifies the maximum batch size to use when writing files. Defaults to 5000 if unspecified. Increasing this may impact the memory usage of your Common Fate deployment.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(5000),
			},
		},
		MarkdownDescription: `Registers an Amazon S3 log destination.

Events are written in batches to the bucket in [JSONL format](https://jsonlines.org/) at a frequency specified by 'batch_duration_in_minutes'.
`,
	}
}

func (r *S3LogDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *S3LogDestinationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	var filters []string

	resp.Diagnostics.Append(data.FilterForActions.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integ := integrationv1alpha1.S3LogDestination{
		BucketName:           data.BucketName.ValueString(),
		RoleArn:              data.RoleARN.ValueString(),
		KeyTemplate:          data.KeyTemplate.ValueString(),
		Compression:          data.Compression.ValueString(),
		FilterForActions:     filters,
		BatchDurationMinutes: uint32(data.BatchDurationInMinutes.ValueInt64()),
		MaximumBatchSize:     uint32(data.MaximumBatchSize.ValueInt64()),
	}

	res, err := r.client.CreateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.CreateIntegrationRequest{
		Name: data.Name.ValueString(),
		Config: &integrationv1alpha1.Config{
			Config: &integrationv1alpha1.Config_S3LogDestination{
				S3LogDestination: &integ,
			},
		},
	}))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: S3 Log Destination",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	diags.ToTerraform(res.Msg.Integration.Diagnostics, &resp.Diagnostics)

	data.ID = types.StringValue(res.Msg.Integration.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *S3LogDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state S3LogDestinationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//read the state from the client
	res, err := r.client.GetIntegration(ctx, connect.NewRequest(&integrationv1alpha1.GetIntegrationRequest{
		Id: state.ID.ValueString(),
	}))

	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read S3 Log Destination",
			err.Error(),
		)
		return
	}

	integ := res.Msg.Integration.Config.GetS3LogDestination()
	if integ == nil {
		resp.Diagnostics.AddError(
			"Returned integration did not contain any S3 Log Destination configuration",
			"",
		)
		return
	}

	state = S3LogDestinationModel{
		ID:                     types.StringValue(state.ID.ValueString()),
		Name:                   types.StringValue(res.Msg.Integration.Name),
		BucketName:             types.StringValue(integ.BucketName),
		RoleARN:                types.StringValue(integ.RoleArn),
		KeyTemplate:            types.StringValue(integ.KeyTemplate),
		Compression:            types.StringValue(integ.Compression),
		BatchDurationInMinutes: types.Int64Value(int64(integ.BatchDurationMinutes)),
		MaximumBatchSize:       types.Int64Value(int64(integ.MaximumBatchSize)),
	}

	filters, diags := types.SetValueFrom(ctx, types.StringType, integ.FilterForActions)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state.FilterForActions = filters

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *S3LogDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data S3LogDestinationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to update S3 Log Destination",
			"An unexpected error occurred while parsing the resource update response.",
		)

		return
	}

	var filters []string

	resp.Diagnostics.Append(data.FilterForActions.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	integ := integrationv1alpha1.S3LogDestination{
		BucketName:           data.BucketName.ValueString(),
		RoleArn:              data.RoleARN.ValueString(),
		KeyTemplate:          data.KeyTemplate.ValueString(),
		Compression:          data.Compression.ValueString(),
		FilterForActions:     filters,
		BatchDurationMinutes: uint32(data.BatchDurationInMinutes.ValueInt64()),
		MaximumBatchSize:     uint32(data.MaximumBatchSize.ValueInt64()),
	}

	res, err := r.client.UpdateIntegration(ctx, connect.NewRequest(&integrationv1alpha1.UpdateIntegrationRequest{
		Integration: &integrationv1alpha1.Integration{
			Id:   data.ID.ValueString(),
			Name: data.Name.ValueString(),
			Config: &integrationv1alpha1.Config{
				Config: &integrationv1alpha1.Config_S3LogDestination{
					S3LogDestination: &integ,
				},
			},
		},
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update S3 Log Destination",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	diags.ToTerraform(res.Msg.Integration.Diagnostics, &resp.Diagnostics)

	data.ID = types.StringValue(res.Msg.Integration.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *S3LogDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *S3LogDestinationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete S3 Log Destination",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteIntegration(ctx, connect.NewRequest(&integrationv1alpha1.DeleteIntegrationRequest{
		Id: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete S3 Log Destination",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *S3LogDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

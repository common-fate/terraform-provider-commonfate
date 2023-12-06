package internal

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	config_client "github.com/common-fate/sdk/config"
	configv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1"
	configv1alpha1connect "github.com/common-fate/sdk/gen/commonfate/control/config/v1alpha1/configv1alpha1connect"
	pagerduty_handler "github.com/common-fate/sdk/service/control/config/pagerduty"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &PagerdutyScheduleDataSource{}
var _ datasource.DataSourceWithConfigure = &PagerdutyScheduleDataSource{}

type PagerdutyScheduleDataSource struct {
	client configv1alpha1connect.PagerDutyServiceClient
}

type PagerdutyScheduleDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *PagerdutyScheduleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pagerduty_schedules"
}

func (d *PagerdutyScheduleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{

		Attributes: map[string]schema.Attribute{

			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The computed internal ID for this data source.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the PagerDuty Schedule. This can be found in your PagerDuty admin console.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *PagerdutyScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config_client.Context)

	client := pagerduty_handler.NewFromConfig(cfg)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *PagerdutyScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PagerdutyScheduleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}

	//read the state from the client
	schedule, err := d.client.ReadPagerDutySchedules(ctx, connect.NewRequest(&configv1alpha1.ReadPagerDutySchedulesRequest{
		Id: data.Name.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to pull PagerDuty Schedules",
			err.Error(),
		)
		return
	}
	scheduleModel := PagerdutyScheduleDataSourceModel{
		ID:   types.StringValue(schedule.Msg.Id),
		Name: types.StringValue(schedule.Msg.Name),
	}

	data = scheduleModel

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

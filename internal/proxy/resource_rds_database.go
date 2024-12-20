package proxy

import (
	"context"
	"fmt"

	integrationv1alpha1 "github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1"

	config_client "github.com/common-fate/sdk/config"
	"github.com/common-fate/sdk/gen/commonfate/control/integration/v1alpha1/integrationv1alpha1connect"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RDSDatabaseModel struct {
	ID           types.String `tfsdk:"id"`
	InstanceID   types.String `tfsdk:"instance_id"`
	Name         types.String `tfsdk:"name"`
	Engine       types.String `tfsdk:"engine"`
	Endpoint     types.String `tfsdk:"endpoint"`
	Region       types.String `tfsdk:"region"`
	AWSAccountID types.String `tfsdk:"aws_account_id"`
	Database     types.String `tfsdk:"database"`
	ProxyId      types.String `tfsdk:"proxy_id"`

	Users []DatabaseUser `tfsdk:"users"`
}

type Database struct {
}

type DatabaseUser struct {
	Name                      types.String `tfsdk:"name"`
	UserName                  types.String `tfsdk:"username"`
	PasswordSecretsManagerARN types.String `tfsdk:"password_secrets_manager_arn"`
	Endpoint                  types.String `tfsdk:"endpoint"`
	DefaultLocalPort          types.Int64  `tfsdk:"default_local_port"`
}

// AccessRuleResource is the data source implementation.
type RDSDatabaseResource struct {
	client integrationv1alpha1connect.ProxyServiceClient
}

var (
	_ resource.Resource                = &RDSDatabaseResource{}
	_ resource.ResourceWithConfigure   = &RDSDatabaseResource{}
	_ resource.ResourceWithImportState = &RDSDatabaseResource{}
)

// Metadata returns the data source type name.
func (r *RDSDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy_rds_database"
}

// Configure adds the provider configured client to the data source.
func (r *RDSDatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client := integrationv1alpha1connect.NewProxyServiceClient(cfg.HTTPClient, cfg.APIURL)

	r.client = client
}

// GetSchema defines the schema for the data source.
// schema is based off the governance api
func (r *RDSDatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Access Workflows are used to describe how long access should be applied. Created Workflows can be referenced in other resources created.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The internal resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"instance_id": schema.StringAttribute{
				MarkdownDescription: "The name of the parent instance id that the database will be connected to",
				Required:            true,
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "A unique name for the resource",
				Required:            true,
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "A SQL engine of the RDS database",
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region the database is in",
				Required:            true,
			},
			"aws_account_id": schema.StringAttribute{
				MarkdownDescription: "The AWS account id the database is in",
				Required:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The endpoint of the database",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "The name of the database",
				Required:            true,
			},
			"proxy_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the proxy in the same account as the database.",
				Required:            true,
			},

			"users": schema.ListNestedAttribute{
				MarkdownDescription: "A list of users that exist in the database",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name for the user",
							Required:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "The user name for the user",
							Required:            true,
						},
						"password_secrets_manager_arn": schema.StringAttribute{
							MarkdownDescription: "The secrets manager arn for the RDS database password",
							Required:            true,
						},
						"endpoint": schema.StringAttribute{
							MarkdownDescription: "Override default endpoint behaviour by specifying a endpoint on a per user basis.",
							Optional:            true,
						},
						"default_local_port": schema.Int64Attribute{
							MarkdownDescription: "The default local port to use for the user when running `granted rds proxy`",
							Optional:            true,
							Default:             int64default.StaticInt64(0),
							Computed:            true,
						},
					},
				},
			},
		},
		MarkdownDescription: `Registers a RDS database with a Common Fate Proxy.`,
	}
}

func (r *RDSDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *RDSDatabaseModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	resource := &integrationv1alpha1.AWSRDSDatabase{
		Name:       data.Name.ValueString(),
		Engine:     data.Engine.ValueString(),
		InstanceId: data.InstanceID.ValueString(),
		Region:     data.Region.ValueString(),
		Database:   data.Database.ValueString(),
		Endpoint:   data.Endpoint.ValueString(),
		Account:    data.AWSAccountID.ValueString(),
	}

	for _, user := range data.Users {
		newUser := &integrationv1alpha1.AWSRDSDatabaseUser{
			Name:                      user.Name.ValueString(),
			Username:                  user.UserName.ValueString(),
			PasswordSecretsManagerArn: user.PasswordSecretsManagerARN.ValueString(),
			DefaultLocalPort:          uint32(user.DefaultLocalPort.ValueInt64()),
		}

		if user.Endpoint.ValueString() != "" {
			newUser.Endpoint = user.Endpoint.ValueStringPointer()
		}
		resource.Users = append(resource.Users, newUser)

	}

	createReq := integrationv1alpha1.CreateProxyRdsResourceRequest{
		ProxyId: data.ProxyId.ValueString(),

		RdsDatabase: resource,
	}

	res, err := r.client.CreateProxyRdsResource(ctx, connect.NewRequest(&createReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource: RDS Database",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *RDSDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var state RDSDatabaseModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// read the state from the client
	res, err := r.client.GetProxyRdsResource(ctx, connect.NewRequest(&integrationv1alpha1.GetProxyRdsResourceRequest{
		Id: state.ID.ValueString(),
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read RDS resource",
			err.Error(),
		)
		return
	}

	// refresh state
	state = RDSDatabaseModel{
		ID:           types.StringValue(res.Msg.Id),
		InstanceID:   types.StringValue(res.Msg.RdsDatabase.InstanceId),
		Name:         types.StringValue(res.Msg.RdsDatabase.Name),
		Engine:       types.StringValue(res.Msg.RdsDatabase.Engine),
		Endpoint:     types.StringValue(res.Msg.RdsDatabase.Endpoint),
		Region:       types.StringValue(res.Msg.RdsDatabase.Region),
		AWSAccountID: types.StringValue(res.Msg.RdsDatabase.Account),
		Database:     types.StringValue(res.Msg.RdsDatabase.Database),
		ProxyId:      types.StringValue(res.Msg.ProxyId),
	}
	for _, user := range res.Msg.RdsDatabase.Users {
		state.Users = append(state.Users, DatabaseUser{
			Name:                      types.StringValue(user.Name),
			UserName:                  types.StringValue(user.Username),
			PasswordSecretsManagerARN: types.StringValue(user.PasswordSecretsManagerArn),
			DefaultLocalPort:          types.Int64Value(int64(user.DefaultLocalPort)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RDSDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)

		return
	}
	var data *RDSDatabaseModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	resource := &integrationv1alpha1.AWSRDSDatabase{

		Name:       data.Name.ValueString(),
		Engine:     data.Engine.ValueString(),
		InstanceId: data.InstanceID.ValueString(),
		Region:     data.Region.ValueString(),
		Account:    data.AWSAccountID.ValueString(),
		Database:   data.Database.ValueString(),
		Endpoint:   data.Endpoint.ValueString(),
	}

	for _, user := range data.Users {

		newUser := &integrationv1alpha1.AWSRDSDatabaseUser{
			Name:                      user.Name.ValueString(),
			Username:                  user.UserName.ValueString(),
			PasswordSecretsManagerArn: user.PasswordSecretsManagerARN.ValueString(),
			DefaultLocalPort:          uint32(user.DefaultLocalPort.ValueInt64()),
		}

		if user.Endpoint.ValueString() != "" {
			newUser.Endpoint = user.Endpoint.ValueStringPointer()
		}
		resource.Users = append(resource.Users, newUser)

	}

	updateReq := integrationv1alpha1.UpdateProxyRdsResourceRequest{
		Id:          data.ID.ValueString(),
		ProxyId:     data.ProxyId.ValueString(),
		RdsDatabase: resource,
	}

	res, err := r.client.UpdateProxyRdsResource(ctx, connect.NewRequest(&updateReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource: RDS Database",
			"An unexpected error occurred while communicating with Common Fate API. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	// // Convert from the API data model to the Terraform data model
	// // and set any unknown attribute values.
	data.ID = types.StringValue(res.Msg.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RDSDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
	}
	var data *RDSDatabaseModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response.",
		)

		return
	}

	_, err := r.client.DeleteProxyRdsResource(ctx, connect.NewRequest(&integrationv1alpha1.DeleteProxyRdsResourceRequest{
		Id: data.ID.ValueString(),
	}))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Resource",
			"An unexpected error occurred while parsing the resource creation response. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}
}

func (r *RDSDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

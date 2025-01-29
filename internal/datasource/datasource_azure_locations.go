package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource                     = &azureLocationsDataSource{}
	_ datasource.DataSourceWithConfigValidators = &azureLocationsDataSource{}
)

// New is a helper function to simplify the provider implementation.
func NewAzureLocations() datasource.DataSource {
	return &azureLocationsDataSource{}
}

// data source implementation.
type azureLocationsDataSource struct {
}

type azureLocationsDataSourceModel struct {
	SubscriptionID    types.String `tfsdk:"subscription_id"`
	SubscriptionName  types.String `tfsdk:"subscription_display_name"`
	LocationOverrides types.Map    `tfsdk:"localtion_overrides"`
	LocationMaps      types.Map    `tfsdk:"location_maps"`
}

func (d *azureLocationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_locations"
}

func (d *azureLocationsDataSource) Schema(ctx context.Context, ds datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data resource fetches types from the Azure CAF project for use in the configuration parameter of `namestring`.",
		Attributes: map[string]schema.Attribute{
			"subscription_id": schema.StringAttribute{
				Description: "Subscription ID to pull locations from (cannot be used with `subscription_display_name`).",
				Required:    false,
				Optional:    true,
			},
			"subscription_display_name": schema.StringAttribute{
				Description: "Subscription Display Name to pull locations from (cannot be used with `subscription_id`).",
				Required:    false,
				Optional:    true,
			},
			"localtion_overrides": schema.MapAttribute{
				Description: "Variable maps to override specifc parts of the final location maps.",
				Required:    false,
				Optional:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
			"location_maps": schema.ObjectAttribute{
				Description:    "Maps to support location name substitutions.",
				Computed:       true,
				AttributeTypes: configAttributes(),
			},
		},
	}
}

func (d *azureLocationsDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("subscription_id"),
			path.MatchRoot("subscription_display_name"),
		),
	}
}

func (d *azureLocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config azureLocationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read configuration data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

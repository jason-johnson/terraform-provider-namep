package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource                     = &azureCafTypesDataSource{}
	_ datasource.DataSourceWithConfigure        = &azureCafTypesDataSource{}
	_ datasource.DataSourceWithConfigValidators = &azureCafTypesDataSource{}
)

// New is a helper function to simplify the provider implementation.
func NewAzureCafTypes() datasource.DataSource {
	return &azureCafTypesDataSource{}
}

// data source implementation.
type azureCafTypesDataSource struct {
}

func (d *azureCafTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_caf_types"
}

func (d *azureCafTypesDataSource) Schema(ctx context.Context, ds datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data resource fetches types from the Azure CAF project for use in the configuration parameter of `namestring`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"newest": schema.BoolAttribute{
				Description: "If true, this data source will always check if it has the latest version of the data.",
				Required:    false,
				Optional:    true,
			},
			"version": schema.StringAttribute{
				Description: "The version of the Azure CAF types to fetch.  Cannot be set with `newest`.",
				Required:    false,
				Optional:    true,
			},
			"types": schema.MapAttribute{
				Description: "The type info map loaded from the Azure CAF project.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":               types.StringType,
						"slug":               types.StringType,
						"min_length":         types.Int32Type,
						"max_length":         types.Int32Type,
						"lowercase":          types.BoolType,
						"validatation_regex": types.StringType,
						"default_selector":   types.StringType,
					},
				},
			},
		},
	}
}

func (d *azureCafTypesDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("newest"),
			path.MatchRoot("version"),
		),
	}
}

func (d *azureCafTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

}

func (d *azureCafTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

}

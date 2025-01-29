package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &configurationDataSource{}
)

// New is a helper function to simplify the provider implementation.
func NewConfiguration() datasource.DataSource {
	return &configurationDataSource{}
}

// data source implementation.
type configurationDataSource struct {
}

type configurationDataSourceModel struct {
	Formats       types.Map    `tfsdk:"formats"`
	Variables     types.Map    `tfsdk:"variables"`
	VariableMaps  types.Map    `tfsdk:"variable_maps"`
	Types         types.Map    `tfsdk:"types"`
	Configuration types.Object `tfsdk:"configuration"`
}

type configurationModel struct {
	Formats      types.Map `tfsdk:"formats"`
	Variables    types.Map `tfsdk:"variables"`
	VariableMaps types.Map `tfsdk:"variable_maps"`
	Types        types.Map `tfsdk:"types"`
}

func (d *configurationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration"
}

func (d *configurationDataSource) Schema(ctx context.Context, ds datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data resource fetches types from the Azure CAF project for use in the configuration parameter of `namestring`.",
		Attributes: map[string]schema.Attribute{
			"formats": schema.MapAttribute{
				Description: "Formats map to include in final configuration.",
				Required:    false,
				Optional:    true,
				ElementType: types.StringType,
			},
			"variables": schema.MapAttribute{
				Description: "Variables map to include in final configuration.",
				Required:    false,
				Optional:    true,
				ElementType: types.StringType,
			},
			"variable_maps": schema.MapAttribute{
				Description: "Variable maps map to include in final configuration.",
				Required:    false,
				Optional:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
			"types": schema.MapAttribute{
				Description: `A map of types to include in the final configuration.`,
				Required:    false,
				Optional:    true,
				ElementType: typesAttributes(),
			},
			"configuration": schema.ObjectAttribute{
				Description:    "The configuration produced from the inputs.",
				Computed:       true,
				AttributeTypes: configAttributes(),
			},
		},
	}
}

func (d *configurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config configurationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	formats, diag := tomap(ctx, config.Formats)
	if diag.HasError() {
		resp.Diagnostics.Append(diag.Errors()...)
	}
	config.Formats = formats

	variables, diag := tomap(ctx, config.Variables)
	if diag.HasError() {
		resp.Diagnostics.Append(diag.Errors()...)
	}
	config.Variables = variables

	if config.Types.IsNull() {
		configTypes := make(map[string](map[string]string))
		ct, diag := types.MapValueFrom(ctx, typesAttributes(), configTypes)

		if diag.HasError() {
			resp.Diagnostics.Append(diag.Errors()...)
		} else {
			config.Types = ct
		}
	}

	if config.VariableMaps.IsNull() {
		variableMaps := make(map[string](map[string]string))
		vm, diag := types.MapValueFrom(ctx, types.MapType{ElemType: types.StringType}, variableMaps)

		if diag.HasError() {
			resp.Diagnostics.Append(diag.Errors()...)
		} else {
			config.VariableMaps = vm
		}
	}

	configuration := configurationModel{
		Formats:      config.Formats,
		Variables:    config.Variables,
		VariableMaps: config.VariableMaps,
		Types:        config.Types,
	}
	c, diag := types.ObjectValueFrom(ctx, configAttributes(), configuration)

	if diag.HasError() {
		resp.Diagnostics.Append(diag.Errors()...)
	}
	config.Configuration = c

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read configuration data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func configAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"formats": types.MapType{
			ElemType: types.StringType,
		},
		"variables": types.MapType{
			ElemType: types.StringType,
		},
		"variable_maps": types.MapType{
			ElemType: types.MapType{
				ElemType: types.StringType,
			},
		},
		"types": types.MapType{
			ElemType: typesAttributes(),
		},
	}
}

func tomap(ctx context.Context, m types.Map) (types.Map, diag.Diagnostics) {
	if m.IsNull() {
		formats := make(map[string]string)
		return types.MapValueFrom(ctx, types.StringType, formats)
	}
	return m, nil
}

package datasource

import (
	"context"
	"fmt"

	"terraform-provider-namep/internal/shared"
	"terraform-provider-namep/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

type azureCafTypesDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Newest  types.Bool   `tfsdk:"newest"`
	Version types.String `tfsdk:"version"`
	Types   types.Map    `tfsdk:"types"`
}

type cafTypeFields struct {
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	MinLength         int    `json:"min_length"`
	MaxLength         int    `json:"max_length"`
	Lowercase         bool   `json:"lowercase"`
	ValidatationRegex string `json:"validation_regex"`
	Dashes            bool   `json:"dashes"`
	Scope             string `json:"scope"`
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
						"name":             types.StringType,
						"slug":             types.StringType,
						"min_length":       types.Int32Type,
						"max_length":       types.Int32Type,
						"lowercase":        types.BoolType,
						"validation_regex": types.StringType,
						"default_selector": types.StringType,
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
	var config azureCafTypesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	var defs []cafTypeFields

	err := utils.GetJSON("https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/resourceDefinition.json", &defs)

	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch Azure CAF types", err.Error())
		return
	}

	var oodDefs []cafTypeFields

	err = utils.GetJSON("https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/resourceDefinition_out_of_docs.json", &oodDefs)

	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch Azure CAF types", err.Error())
		return
	}

	results := make(map[string]shared.TypeFields, len(defs)+len(oodDefs))

	for _, def := range oodDefs {

		results[def.Name] = toSharedTypeFields(def)
	}

	for _, def := range defs {
		results[def.Name] = toSharedTypeFields(def)
	}

	config.ID = types.StringValue("foo")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read azure name data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func toSharedTypeFields(def cafTypeFields) shared.TypeFields {
	dashes := "nodashes"
	if def.Dashes {
		dashes = "dashes"
	}
	defaultSelector := fmt.Sprintf("azure_%s_%s", dashes, def.Scope)

	return shared.TypeFields{
		Name:            def.Name,
		Slug:            def.Slug,
		MinLength:       def.MinLength,
		MaxLength:       def.MaxLength,
		Lowercase:       def.Lowercase,
		ValidationRegex: def.ValidatationRegex,
		DefaultSelector: defaultSelector,
	}
}

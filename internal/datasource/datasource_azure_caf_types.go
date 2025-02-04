package datasource

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"terraform-provider-namep/internal/cloud/azure"
	"terraform-provider-namep/internal/shared"
	"terraform-provider-namep/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource                     = &azureCafTypesDataSource{}
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
	Version types.String `tfsdk:"version"`
	Static  types.Bool   `tfsdk:"static"`
	Source  types.String `tfsdk:"source"`
	Types   types.Map    `tfsdk:"types"`
}

func (d *azureCafTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_caf_types"
}

func (d *azureCafTypesDataSource) Schema(ctx context.Context, ds datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data resource fetches types from the Azure CAF project for use in the configuration parameter of `namestring`.",
		Attributes: map[string]schema.Attribute{
			"static": schema.BoolAttribute{
				Description: "Static flag to determine if the data source should be static.",
				Required:    false,
				Optional:    true,
			},
			"version": schema.StringAttribute{
				Description: `The version of the Azure CAF types to fetch.  The newest version will be used if not specified.
							  Possible to specify a branch name, tag name or commit hash (hash must be unique but does not have to be complete).`,
				Required: false,
				Optional: true,
			},
			"source": schema.StringAttribute{
				Description: "The source URL the Azure CAF types were loaded from.",
				Computed:    true,
			},
			"types": schema.MapAttribute{
				Description: "The type info map loaded from the Azure CAF project.",
				Computed:    true,
				ElementType: typesAttributes(),
			},
		},
	}
}

func (d *azureCafTypesDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("version"),
			path.MatchRoot("static"),
		),
	}
}

func (d *azureCafTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config azureCafTypesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	var source string
	var typeInfoMap map[string]shared.TypeFields

	if config.Static.ValueBool() {
		source = "static"
		typeInfoMap = make(map[string]shared.TypeFields, len(azure.ResourceDefinitions))

		for _, def := range azure.ResourceDefinitions {
			typeInfoMap[def.ResourceTypeName] = toSharedTypeFields(def, false)
		}
	} else {
		source, typeInfoMap = getTypeInfoMap(config.Version, &resp.Diagnostics)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	typesAttrs := typesAttributes()
	result, diag := types.MapValueFrom(ctx, typesAttrs, typeInfoMap)

	if diag.HasError() {
		resp.Diagnostics.Append(diag.Errors()...)
		return
	}

	config.Source = types.StringValue(source)
	config.Types = result

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read azure caf type data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func getTypeInfoMap(version types.String, diags *diag.Diagnostics) (string, map[string]shared.TypeFields) {
	cafUrl, oodCafUrl, err := getResourceFileStrings(version)

	if err != nil {
		diags.AddError("Failed to determine the version to fetch", err.Error())
		return "", nil
	}

	var defs []azure.ResourceStructure

	err = utils.GetJSON(cafUrl, &defs)

	if err != nil {
		tflog.Error(context.Background(), fmt.Sprintf("Failed to fetch Azure CAF types (url: %s): %v", cafUrl, err))
		diags.AddError("Failed to fetch Azure CAF types", err.Error())
		return "", nil
	}

	var oodDefs []azure.ResourceStructure

	err = utils.GetJSON(oodCafUrl, &oodDefs)

	if err != nil {
		tflog.Error(context.Background(), fmt.Sprintf("Failed to fetch Azure CAF types (url: %s): %v", oodCafUrl, err))
		diags.AddError("Failed to fetch 'out of doc' Azure CAF types", err.Error())
		return "", nil
	}

	typeInfoMap := make(map[string]shared.TypeFields, len(defs)+len(oodDefs))

	for _, def := range oodDefs {
		typeInfoMap[def.ResourceTypeName] = toSharedTypeFields(def, true)
	}

	for _, def := range defs {
		typeInfoMap[def.ResourceTypeName] = toSharedTypeFields(def, true)
	}

	return cafUrl, typeInfoMap
}

func getResourceFileStrings(versionString types.String) (string, string, error) {
	version := versionString.ValueString()

	if versionString.IsNull() {
		version = "master"
	}

	if versionString.IsUnknown() {
		return "", "", fmt.Errorf("unknown version received, please specify the version directly") // should be impossible
	}

	re := regexp.MustCompile(`/^v\d+\.\d+\.\d+(-preview)?$/gm`)

	if re.MatchString(version) {
		version = fmt.Sprintf("refs/tags/%s", version)
	}

	caf := fmt.Sprintf("https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/%s/resourceDefinition.json", version)
	oodcaf := fmt.Sprintf("https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/%s/resourceDefinition_out_of_docs.json", version)

	return caf, oodcaf, nil
}

func toSharedTypeFields(def azure.ResourceStructure, unquoteRegex bool) shared.TypeFields {
	dashes := "nodashes"
	if def.Dashes {
		dashes = "dashes"
	}
	defaultSelector := fmt.Sprintf("azure_%s_%s", dashes, def.Scope)
	validationRegex := def.ValidationRegExp

	if unquoteRegex {
		var err error
		validationRegex, err = strconv.Unquote(def.ValidationRegExp)

		if err != nil {
			tflog.Error(context.Background(), fmt.Sprintf("Failed to unquote validation regex: %v", err))
			validationRegex = def.ValidationRegExp
		}
	}

	return shared.TypeFields{
		Name:            def.ResourceTypeName,
		Slug:            def.CafPrefix,
		MinLength:       def.MinLength,
		MaxLength:       def.MaxLength,
		Lowercase:       def.LowerCase,
		ValidationRegex: validationRegex,
		DefaultSelector: defaultSelector,
	}
}

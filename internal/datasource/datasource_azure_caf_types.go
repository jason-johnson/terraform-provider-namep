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
	_ datasource.DataSourceWithConfigure        = &azureCafTypesDataSource{}
	_ datasource.DataSourceWithConfigValidators = &azureCafTypesDataSource{}
)

// New is a helper function to simplify the provider implementation.
func NewAzureCafTypes() datasource.DataSource {
	return &azureCafTypesDataSource{}
}

// data source implementation.
type azureCafTypesDataSource struct {
	static bool
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
		Description: `This data resource creates a map of type names to type information.  The types are fetched from the [Azure CAF project](https://github.com/aztfmod/terraform-provider-azurecaf), unless the ` + "`static`" + ` field is true.
If the ` + "`static`" + ` field is true then the types retrieved when this provider was built will be used. Note that the static values can get out of date since they cannot be changed without a new version of the provider.  Also note that if ` + "`static`" + ` is
set to true in the provider, it will be used regardless of the value in the data source.  There will, however, be no conflict between the provider ` + "`static`" + ` field and the ` + "`version`" + ` field in this datasource (it will be ignored).

The purpose of this data source is for creating the types to to be passed to the ` + "`types`" + ` parameter in the [namep_configuration](configuration.md) data source.  Alternatively, it could be assigned to a ` + "`locals`" + ` variable to 
add other types for the ` + "`types`" + ` parameter.

## Version Compatibility

**Important**: When using specific Azure CAF versions with this data source, be aware that Azure CAF version ` + "`v1.2.29`" + ` or earlier will not include all available Azure resource types. To have complete Azure resource type coverage, you must either:

- Avoid specifying the ` + "`version`" + ` parameter to get the latest Azure CAF types, or  
- Specify Azure CAF version ` + "`v1.2.30`" + ` or later or
- Use version 2.1.* of this provider

## Default Selector

The ` + "`defaultSelector`" + ` for this resource is made up of 3 components: the word "azure", the word "dashes" or "nodashes" (depending on if dashes are allowed in the name of the resource type), and the ` + "`scope`" + ` of the resource.
The main ` + "`scope`" + ` to be concerned about is the "global" scope, which means the name must be unique across all of Azure.  The other scopes are "subscription", "resourceGroup", and "resource".  When using the ` + "`defaultSelector`" + ` to set
formats for the resources, it is recommended to use at least the first 2 components (e.g. "azure_dashes") since some names cannot have dashes and should have a different format than those which can.
`,
		Attributes: map[string]schema.Attribute{
			"static": schema.BoolAttribute{
				Description: "Static flag to determine if the data source should use data retrieved when this data source was built.  If false, the data source will be downloaded from the Azure CAF project.",
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

func (d *azureCafTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(shared.NamepConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *NamepConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.static = config.Static
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

	if d.static || config.Static.ValueBool() {
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
	cafUrl, err := getResourceFileStrings(version)

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

	typeInfoMap := make(map[string]shared.TypeFields, len(defs))

	for _, def := range defs {
		typeInfoMap[def.ResourceTypeName] = toSharedTypeFields(def, true)
	}

	return cafUrl, typeInfoMap
}

func getResourceFileStrings(versionString types.String) (string, error) {
	version := versionString.ValueString()

	if versionString.IsNull() {
		version = "master"
	}

	if versionString.IsUnknown() {
		return "", fmt.Errorf("unknown version received, please specify the version directly") // should be impossible
	}

	re := regexp.MustCompile(`/^v\d+\.\d+\.\d+(-preview)?$/gm`)

	if re.MatchString(version) {
		version = fmt.Sprintf("refs/tags/%s", version)
	}

	caf := fmt.Sprintf("https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/%s/resourceDefinition.json", version)

	return caf, nil
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

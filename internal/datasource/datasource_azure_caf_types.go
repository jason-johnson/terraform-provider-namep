package datasource

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"terraform-provider-namep/internal/shared"
	"terraform-provider-namep/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &azureCafTypesDataSource{}
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

type cafTypeFields struct {
	Name              string `json:"name"`
	Slug              string `json:"slug,omitempty"`
	MinLength         int    `json:"min_length"`
	MaxLength         int    `json:"max_length"`
	Lowercase         bool   `json:"lowercase,omitempty"`
	Regex             string `json:"regex,omitempty"`
	ValidatationRegex string `json:"validation_regex,omitempty"`
	Dashes            bool   `json:"dashes"`
	Scope             string `json:"scope,omitempty"`
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

func (d *azureCafTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config azureCafTypesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	cafUrl, oodCafUrl, err := getResourceFileStrings(config.Version)

	if err != nil {
		resp.Diagnostics.AddError("Failed to determine the version to fetch", err.Error())
		return
	}

	var defs []cafTypeFields

	err = utils.GetJSON(cafUrl, &defs)

	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch Azure CAF types", err.Error())
		return
	}

	var oodDefs []cafTypeFields

	err = utils.GetJSON(oodCafUrl, &oodDefs)

	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch 'out of doc' Azure CAF types", err.Error())
		return
	}

	typeInfoMap := make(map[string]shared.TypeFields, len(defs)+len(oodDefs))

	for _, def := range oodDefs {
		typeInfoMap[def.Name] = toSharedTypeFields(def)
	}

	for _, def := range defs {
		typeInfoMap[def.Name] = toSharedTypeFields(def)
	}

	typesAttrs := typesAttributes()
	result, diag := types.MapValueFrom(ctx, typesAttrs, typeInfoMap)

	if diag.HasError() {
		resp.Diagnostics.Append(diag.Errors()...)
		return
	}

	config.Source = types.StringValue(cafUrl)
	config.Types = result

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read azure caf type data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
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

func toSharedTypeFields(def cafTypeFields) shared.TypeFields {
	dashes := "nodashes"
	if def.Dashes {
		dashes = "dashes"
	}
	defaultSelector := fmt.Sprintf("azure_%s_%s", dashes, def.Scope)
	validationRegex, err := strconv.Unquote(def.ValidatationRegex)

	if err != nil {
		tflog.Error(context.Background(), fmt.Sprintf("Failed to unquote validation regex: %v", err))
		validationRegex = def.ValidatationRegex
	}

	return shared.TypeFields{
		Name:            def.Name,
		Slug:            def.Slug,
		MinLength:       def.MinLength,
		MaxLength:       def.MaxLength,
		Lowercase:       def.Lowercase,
		ValidationRegex: validationRegex,
		DefaultSelector: defaultSelector,
	}
}

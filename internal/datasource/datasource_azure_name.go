package datasource

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"terraform-provider-namep/internal/cloud/azure"
	"terraform-provider-namep/internal/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &azureNameDataSource{}
	_ datasource.DataSourceWithConfigure = &azureNameDataSource{}
)

// New is a helper function to simplify the provider implementation.
func New() datasource.DataSource {
	return &azureNameDataSource{}
}

// data source implementation.
type azureNameDataSource struct {
	config shared.NamepConfig
}

type azureNameDataSourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	ResourceType              types.String `tfsdk:"type"`
	Location                  types.String `tfsdk:"location"`
	ExtraTokens               types.Map    `tfsdk:"extra_tokens"`
	Result                    types.String `tfsdk:"result"`
	extra_variables_overrides map[string]string
}

func (d *azureNameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_name"
}

// Schema defines the schema for the data source.
func (d *azureNameDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`namep_azure_name` defines a name for an azure resource.\nThe format will be used based on the the resource type selected and the appropriate format string.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			nameProp: schema.StringAttribute{
				Description: "Name to put in the `#{NAME}` location of the formats.",
				Required:    true,
			},
			typeProp: schema.StringAttribute{
				Optional:    true,
				Description: "Type of resource to create a name for (resource name used by terraform, required for `#{SLUG}`).",
				Validators: []validator.String{
					stringInResourceMap(),
				},
			},
			locationProp: schema.StringAttribute{
				Description: "Value to use for the `#{LOC}` portion of the format.  Also used to compute `#{SHORT_LOC}` and `#{ALT_SHORT_LOC}`.",
				Optional:    true,
			},
			extraTokensProp: schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Extra variables for use in format (see Supported Variables) for this data source (may override provider extra_tokens).",
				Optional:    true,
			},
			resultProp: schema.StringAttribute{
				Description: "The name created from the format.",
				Computed:    true,
			},
		}}
}

func (d *azureNameDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.config = config
}

// Read refreshes the Terraform state with the latest data.
func (d *azureNameDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config azureNameDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	extra_variables_overrides := make(map[string]string, len(config.ExtraTokens.Elements()))
	resp.Diagnostics.Append(config.ExtraTokens.ElementsAs(ctx, &extra_variables_overrides, false)...)
	config.extra_variables_overrides = extra_variables_overrides

	if resp.Diagnostics.HasError() {
		return
	}

	name := calculateName(config.Name.ValueString(), d.config, config, &resp.Diagnostics)

	config.ID = types.StringValue(name)
	config.Result = types.StringValue(name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read azure name data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func calculateName(name string, providerConfig shared.NamepConfig, config azureNameDataSourceModel, diags *diag.Diagnostics) string {
	extra_variables := make(map[string]string)

	for name, value := range providerConfig.ExtraVariables {
		extra_variables[strings.ToUpper(name)] = value
	}

	for name, value := range config.extra_variables_overrides {
		extra_variables[strings.ToUpper(name)] = value
	}

	var location string

	if config.Location.IsNull() {
		location = providerConfig.DefaultLocation
	} else {
		location = config.Location.ValueString()
	}

	re := regexp.MustCompile(`#\{-?\w+-?}`)

	var name_type string
	if config.Name.IsNull() || config.Name.ValueString() == "" {
		name_type = "general"
	} else {
		name_type = config.ResourceType.ValueString()
	}

	definition, exists := azure.ResourceDefinitions[name_type]

	var format string

	if !exists {
		format, exists = providerConfig.AzureResourceFormats[name_type]

		if !exists {
			diags.AddError("resource type", fmt.Sprintf("resource type %q unknown to provider", name_type))
			return "FAILED"
		}
	} else {
		format = getFormatString(providerConfig, definition)
	}

	locationDefinition, locsOk := azure.LocationDefinitions[location]

	result := re.ReplaceAllStringFunc(format, func(token string) (r string) {
		tl := len(token)
		if tl < 1 {
			diags.AddError("format (bizarre variable)", fmt.Sprintf("bizarre variable received %q", token))
			return token
		}

		token, prefixDash, postfixDash := preprocessToken(token[2 : tl-1])
		tokenProcessed := true
		var tokenResult string

		switch token {
		case "LOC":
			tokenResult = location // TODO: location could be "", check that
		case "SHORT_LOC":
			if !locsOk {
				diags.AddError("format (SHORT_LOC)", fmt.Sprintf("SHORT_LOC used but no short map for location %q", location))
				tokenProcessed = false
				tokenResult = location
			} else {
				tokenResult = locationDefinition.ShortName
			}
		case "ALT_SHORT_LOC":
			if !locsOk {
				diags.AddError("format (ALT_SHORT_LOC)", fmt.Sprintf("ALT_SHORT_LOC used but no short map for location %q", location))
				tokenProcessed = false
				tokenResult = location
			} else {
				tokenResult = locationDefinition.AltShortName
			}

		case "NAME":
			tokenResult = name
		case "SLUG":
			if definition.CafPrefix == "" {
				if name_type == "general" {
					diags.AddError("format (SLUG: resource_type missing)", fmt.Sprintf("resource type must be defined to use SLUG (format: %s)", format))
					tokenProcessed = false
				} else {
					diags.AddError("format (SLUG: no slug defined)", fmt.Sprintf("no slug defined for resource type '%s'", name_type))
					tokenProcessed = false
				}
			}
			tokenResult = definition.CafPrefix
		default:
			tokenResult, exists = extra_variables[token]

			if !exists {
				idx, hasIndex := getTokenSliceIndex(token)

				if hasIndex {
					if idx >= providerConfig.SliceTokensAvailable {
						diags.AddError("format (TOKEN_ invalid index)", fmt.Sprintf("invalid slice index used ('%s') in format", token))
						tokenProcessed = false
						tokenResult = fmt.Sprintf("${%s}", token)
					} else {
						tokenResult = strings.ToLower(providerConfig.SliceTokens[idx])
					}
				} else {
					diags.AddError("format (unknown variable)", fmt.Sprintf("unknown variable '%s' in format", token))
					tokenProcessed = false
					tokenResult = fmt.Sprintf("${%s}", token)
				}
			}
		}

		if tokenProcessed && len(tokenResult) > 0 {
			if prefixDash {
				return string('-') + tokenResult
			} else if postfixDash {
				return tokenResult + string('-')
			}
		}
		return tokenResult
	})

	validateResult(result, definition, diags)

	return result
}

func preprocessToken(token string) (result string, pre bool, post bool) {
	pre = false
	post = false
	result = token
	l := len(token)

	if token[0] == '-' {
		pre = true
		result = token[1:]
	} else if token[l-1] == '-' {
		post = true
		result = token[0 : l-2]
	}

	return result, pre, post
}

func getTokenSliceIndex(token string) (int, bool) {
	re := regexp.MustCompile(`TOKEN_(\d+)`)

	results := re.FindStringSubmatch(token)

	if len(results) != 2 {
		return 0, false
	}

	result, err := strconv.Atoi(results[1])

	if err != nil { // should be impossible
		return 0, false
	}

	return result - 1, true
}

func getFormatString(config shared.NamepConfig, def azure.ResourceStructure) string {
	format, exists := config.AzureResourceFormats[def.ResourceTypeName]

	if !exists {
		if def.Dashes {
			format = config.DefaultResourceNameFormat
		} else {
			format = config.DefaultNodashNameFormat
		}
	}

	return format
}

func validateResult(result string, definition azure.ResourceStructure, diags *diag.Diagnostics) {
	errorSeen := false

	if definition.LowerCase && strings.ToLower(result) != result {
		diags.AddError("validate", fmt.Sprintf("resulting name must be lowercase: %s", result))
		errorSeen = true
	}

	var validName = regexp.MustCompile(definition.ValidationRegExp)

	if !validName.MatchString(result) {

		if len(result) > definition.MaxLength {
			diags.AddError("validate", fmt.Sprintf("resulting name is too long (%d > %d): %s", len(result), definition.MaxLength, result))
			errorSeen = true
		}

		// NOTE: Regex will generally catch everything but not tell us what's wrong so we only show it if
		// NOTE: nothing else was a problem.  This could hide an error with the string until the other issues are fixed
		if !errorSeen {
			diags.AddError("validate", fmt.Sprintf("resulting name is invalid (validation regex: %s): %s", definition.ValidationRegExp, result))
		}
	}
}

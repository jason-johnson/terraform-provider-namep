package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAzureName() *schema.Resource {
	return &schema.Resource{
		Description: "`namep_azure_name` defines a name for an azure resource.\n\n" +
			"The format will be used based on the the resource type selected and the appropriate format string.",

		ReadContext: dataSourceNameRead,

		Schema: map[string]*schema.Schema{
			nameProp: {
				Type:        schema.TypeString,
				Description: "Name to put in the #{NAME} location of the formats.",
				ForceNew:    true,
				Required:    true,
			},
			typeProp: {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Type of resource to create a name for (resource name used by terraform, required for #{SLUG}).",
				ValidateDiagFunc: stringInResourceMapKeys(ResourceDefinitions),
				ForceNew:         true,
			},
			locationProp: {
				Type:        schema.TypeString,
				Description: "Value to use for the #{LOC} portion of the format.  Also used to compute #{SHORT_LOC} and #{ALT_SHORT_LOC}.",
				Optional:    true,
				ForceNew:    true,
				Default:     "",
			},
			resultProp: {
				Type:        schema.TypeString,
				Description: "The name created from the format.",
				Computed:    true,
			},
		},
	}
}

func dataSourceNameRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get(nameProp).(string)

	result, diag := calculateName(name, d, m)

	if !diag.HasError() {
		d.Set(resultProp, result)
		d.SetId(name)
	}

	return diag
}

func calculateName(name string, d *schema.ResourceData, m interface{}) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	config, ok := m.(providerConfiguration)

	if !ok {
		return "", diag.Errorf("panic: provider configuration was of the wrong type")
	}
	location, diags := getValue(locationProp, d, config.default_location, diags)

	if diags.HasError() {
		return "", diags
	}

	re := regexp.MustCompile(`#\{-?\w+-?}`)

	name_type := d.Get(typeProp).(string)

	if name_type == "" {
		name_type = "general"
	}

	definition, exists := ResourceDefinitions[name_type]

	if !exists {
		diags = appendError(diags, fmt.Sprintf("resource type %q unknown to module", name_type))
		return "", diags
	}

	format, diags := getFormatString(d, config, definition, diags)
	locationDefinition, locsOk := LocationDefinitions[location]

	result := re.ReplaceAllStringFunc(format, func(token string) (r string) {
		tl := len(token)
		if tl < 1 {
			diags = appendError(diags, fmt.Sprintf("bizarre token received %q", token))
			return token
		}

		token, prefixDash, postfixDash := preprocessToken(token[2 : tl-1])
		tokenProcessed := true
		var tokenResult string

		switch token {
		case "LOC":
			tokenResult = location
		case "SHORT_LOC":
			if !locsOk {
				diags = appendError(diags, fmt.Sprintf("SHORT_LOC used but no short map for location %q", location))
				tokenProcessed = false
				tokenResult = location
			} else {
				tokenResult = locationDefinition.ShortName
			}
		case "ALT_SHORT_LOC":
			if !locsOk {
				diags = appendError(diags, fmt.Sprintf("ALT_SHORT_LOC used but no short map for location %q", location))
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
					diags = appendError(diags, fmt.Sprintf("resource type must be defined to use SLUG (format: %s)", format))
					tokenProcessed = false
				} else {
					diags = appendError(diags, fmt.Sprintf("no slug defined for resource type '%s'", name_type))
					tokenProcessed = false
				}
			}
			tokenResult = definition.CafPrefix
		default:
			tokenResult, exists = config.extra_tokens[token]

			if !exists {
				idx, hasIndex := getTokenSliceIndex(token)

				if hasIndex {
					if idx >= config.slice_tokens_available {
						diags = appendError(diags, fmt.Sprintf("invalid slice index used ('%s') in format", token))
						tokenProcessed = false
						tokenResult = fmt.Sprintf("${%s}", token)
					} else {
						tokenResult = strings.ToLower(config.slice_tokens[idx])
					}
				} else {
					diags = appendError(diags, fmt.Sprintf("unknown token '%s' in format", token))
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

	diags = validateResult(result, definition, diags)

	return result, diags
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

func getValue(field string, d *schema.ResourceData, defaultResult string, diags diag.Diagnostics) (string, diag.Diagnostics) {
	result := d.Get(field).(string)

	if result == "" {
		if defaultResult == "" {
			diags = appendError(diags, fmt.Sprintf("%s must be supplied as default or in the resource", field))
			return "", diags
		}
		return defaultResult, diags
	}
	return result, diags
}

func getFormatString(d *schema.ResourceData, config providerConfiguration, def ResourceStructure, diags diag.Diagnostics) (string, diag.Diagnostics) {
	format, exists := config.resource_formats[def.ResourceTypeName]

	if !exists {
		if def.Dashes {
			format = config.default_resource_name_format
		} else {
			format = config.default_nodash_name_format
		}
	}

	return format, diags
}

func validateResult(result string, definition ResourceStructure, diags diag.Diagnostics) diag.Diagnostics {
	errorSeen := false

	if definition.LowerCase && strings.ToLower(result) != result {
		diags = appendError(diags, fmt.Sprintf("resulting name must be lowercase: %s", result))
		errorSeen = true
	}

	var validName = regexp.MustCompile(definition.ValidationRegExp)

	if !validName.MatchString(result) {

		if len(result) > definition.MaxLength {
			diags = appendError(diags, fmt.Sprintf("resulting name is too long (%d > %d): %s", len(result), definition.MaxLength, result))
			errorSeen = true
		}

		// NOTE: Regex will generally catch everything but not tell us what's wrong so we only show it if
		// NOTE: nothing else was a problem.  This could hide an error with the string until the other issues are fixed
		if !errorSeen {
			diags = appendError(diags, fmt.Sprintf("resulting name is invalid (validation regex: %s): %s", definition.ValidationRegExp, result))
		}
	}

	return diags
}

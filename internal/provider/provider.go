package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				sliceStringProp: {
					Type: schema.TypeString,
					Description: "A String containing strings seperated by space (see Example Usage) which can be used in the format via " +
						"the `TOKEN_*` variables (see Supported Variables).\n\nThe point of this attribute is so users who have a " +
						"terraform string from some other resource (e.g. `subscription_name`) don't have to pre-process it but can " +
						"simply apply it here and use parts of it in their formats.",
					Optional: true,
				},
				extraTokensProp: {
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "Extra variables for use in format (see Supported Variables).  These can be overriden at the data source level.",
					Optional:    true,
				},
				defaultLocationProp: {
					Type:        schema.TypeString,
					Description: "Location to use if none are specified in the data_source.",
					Optional:    true,
				},
				defaultResourceNameFormatProp: {
					Type:        schema.TypeString,
					Description: "Default format to use for resources which can have dashes.",
					Optional:    true,
					Default:     "#{SLUG}-#{SHORT_LOC}-#{NAME}",
				},
				defaultNodashNameFormatProp: {
					Type:        schema.TypeString,
					Description: "Default format to use for resources which can not have dashes.",
					Optional:    true,
					Default:     "#{SLUG}#{SHORT_LOC}#{NAME}",
				},
				resourceFormatsProp: {
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "Which format to use for specific resource types (see Example Usage).\n\nThe purpose of this attribute " +
						"is to allow overrides to the format only for specific resources.",
					Optional: true,
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"namep_azure_name": dataSourceAzureName(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type providerConfiguration struct {
	default_location             string
	default_resource_name_format string
	default_nodash_name_format   string
	resource_formats             map[string]string
	extra_tokens                 map[string]string
	slice_tokens_available       int
	slice_tokens                 []string
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (result interface{}, diags diag.Diagnostics) {
		slice_string := d.Get(sliceStringProp).(string)

		slice_tokens := strings.Fields(slice_string)

		extra_tokens := make(map[string]string)
		extra_tokens_values, et_exists := d.GetOk(extraTokensProp)

		if et_exists {
			for name, value := range extra_tokens_values.(map[string]interface{}) {
				extra_tokens[strings.ToUpper(name)] = strings.ToLower(value.(string))
			}
		}

		resource_formats := make(map[string]string)
		resource_formats_schema, rf_exists := d.GetOk(resourceFormatsProp)

		if rf_exists {
			for name, value := range resource_formats_schema.(map[string]interface{}) {
				resource_formats[name] = value.(string)
			}
		}

		result = providerConfiguration{
			slice_tokens:                 slice_tokens,
			extra_tokens:                 extra_tokens,
			default_location:             d.Get(defaultLocationProp).(string),
			default_resource_name_format: d.Get(defaultResourceNameFormatProp).(string),
			default_nodash_name_format:   d.Get(defaultNodashNameFormatProp).(string),
			resource_formats:             resource_formats,
			slice_tokens_available:       len(slice_tokens),
		}

		return result, diags
	}
}

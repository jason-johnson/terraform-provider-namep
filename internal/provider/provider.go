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
					Type:        schema.TypeString,
					Description: "String containing tokens to be inserted in the format via #{TOKEN_1}, #{TOKEN_2}, etc.",
					Optional:    true,
					Default:     "",
				},
				extraTokensProp: {
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "Extra tokens for use in format (token names will be upper cased)",
					Optional:    true,
				},
				defaultLocationProp: {
					Type:        schema.TypeString,
					Description: "Location to use if none specified",
					Optional:    true,
					Default:     "",
				},
				defaultResourceNameFormatProp: {
					Type:        schema.TypeString,
					Description: "Default format to use for type unspecified resources",
					Optional:    true,
					Default:     "#{SLUG}-#{SHORT_LOC}-#{NAME}",
				},
				defaultNodashNameFormatProp: {
					Type:        schema.TypeString,
					Description: "Default format to use for entries with no dash",
					Optional:    true,
					Default:     "#{SLUG}#{SHORT_LOC}#{NAME}",
				},
				resourceFormatsProp: {
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "Format to use for specific resource types",
					Optional:    true,
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"scaffolding_data_source": dataSourceScaffolding(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"scaffolding_resource": resourceScaffolding(),
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

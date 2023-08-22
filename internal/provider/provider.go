package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

var (
	_ provider.Provider = &namepProvider{}
)

/* func init() {
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
} */

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &namepProvider{
			version: version,
		}
	}
}

type namepProvider struct {
	version string
}

type namepProviderModel struct {
	slice_string                 types.String `tfsdk:"slice_string"`
	default_location             types.String `tfsdk:"default_location"`
	default_resource_name_format types.String `tfsdk:"default_resource_name_format"`
	default_nodash_name_format   types.String `tfsdk:"default_nodash_name_format"`
	azure_resource_formats       types.Map    `tfsdk:"azure_resource_formats"`
	custom_resource_formats      types.Map    `tfsdk:"custom_resource_formats"`
	extra_tokens                 types.Map    `tfsdk:"extra_tokens"`
}

type NamepConfig struct {
	slice_tokens                 []string
	slice_tokens_available       int
	extra_variables              map[string]string
	default_location             string
	default_resource_name_format string
	default_nodash_name_format   string
	resource_formats             map[string]string
}

func (p *namepProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "namep"
	resp.Version = p.version
}

func (p *namepProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			sliceStringProp: schema.StringAttribute{
				Description: "A String containing strings seperated by space (see Example Usage) which can be used in the format via " +
					"the `TOKEN_*` variables (see Supported Variables).\n\nThe point of this attribute is so users who have a " +
					"terraform string from some other resource (e.g. `subscription_name`) don't have to pre-process it but can " +
					"simply apply it here and use parts of it in their formats.",
				Optional: true,
			},
			extraTokensProp: schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Extra variables for use in format (see Supported Variables).  These can be overriden at the data source level.",
				Optional:    true,
			},
			defaultLocationProp: schema.StringAttribute{
				Description: "Location to use if none are specified in the data_source.",
				Optional:    true,
			},
			defaultResourceNameFormatProp: schema.StringAttribute{
				Description: "Default format to use for resources which can have dashes.",
				Optional:    true,
				//Default:     "#{SLUG}-#{SHORT_LOC}-#{NAME}",
			},
			defaultNodashNameFormatProp: schema.StringAttribute{
				Description: "Default format to use for resources which can not have dashes.",
				Optional:    true,
				//Default:     "#{SLUG}#{SHORT_LOC}#{NAME}",
			},
			azureResourceFormatsProp: schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Which format to use for specific azure resource types (see Example Usage).\n\nThe purpose of this attribute " +
					"is to allow overrides to the format only for specific resources.",
				Optional: true,
			},
			customResourceFormatsProp: schema.MapAttribute{
				ElementType: types.StringType,
				Description: "Which format to use for specific custom resource types (see Example Usage).\n\nThe purpose of this attribute " +
					"is to allow creation of special custom formats for things that may not be recources.",
				Optional: true,
			},
		},
	}
}

func (p *namepProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config namepProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var npConfig NamepConfig

	if config.slice_string.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(sliceStringProp),
			"Unknown slice_string",
			"The provider cannot create names as there is an unknown configuration value for the slice_string. "+
				"Either target apply the source of the value first or set the value statically in the configuration.",
		)
	}

	npConfig.slice_tokens = strings.Fields(config.slice_string.ValueString())

	if config.extra_tokens.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(extraTokensProp),
			"Unknown extra_tokens",
			"The provider cannot create names as there is an unknown configuration value for the extra_tokens. "+
				"Either target apply the source of the value first or set the value statically in the configuration.",
		)
	}

	extra_variables := make(map[string]string, len(config.extra_tokens.Elements()))

	for key, value := range config.extra_tokens.Elements() {
		if value.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root(extraTokensProp).AtMapKey(key),
				fmt.Sprintf("Unknown: %s.%s)", extraTokensProp, key),
				fmt.Sprintf("The provider cannot create names as there is an unknown configuration value for %s.%s. "+
					"Either target apply the source of the value first or set the value statically in the configuration.",
					extraTokensProp, key),
			)
		}

		extra_variables[strings.ToUpper(key)] = value.String()
	}

	npConfig.extra_variables = extra_variables

	azure_resource_formats := make(map[string]types.String, len(config.azure_resource_formats.Elements()))
	resp.Diagnostics.Append(config.azure_resource_formats.ElementsAs(ctx, &azure_resource_formats, false)...)
}

func (p *namepProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
	}
}

func (p *namepProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
	}
}

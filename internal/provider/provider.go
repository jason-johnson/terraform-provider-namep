package provider

import (
	"context"
	"fmt"
	"strings"
	namep "terraform-provider-namep/internal/datasource"
	"terraform-provider-namep/internal/shared"
	"terraform-provider-namep/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &namepProvider{}
)

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
	SliceString                     types.String `tfsdk:"slice_string"`
	DefaultLocation                 types.String `tfsdk:"default_location"`
	DefaultResourceNameFormat       types.String `tfsdk:"default_resource_name_format"`
	DefaultNodashNameFormat         types.String `tfsdk:"default_nodash_name_format"`
	DefaultGlobalResourceNameFormat types.String `tfsdk:"default_global_resource_name_format"`
	DefaultGlobalNodashNameFormat   types.String `tfsdk:"default_global_nodash_name_format"`
	AzureResourceFormats            types.Map    `tfsdk:"azure_resource_formats"`
	CustomResourceFormats           types.Map    `tfsdk:"custom_resource_formats"`
	ExtraTokens                     types.Map    `tfsdk:"extra_tokens"`
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
				Description: "Default format to use for resources which can have dashes. Defaults to `#{SLUG}-#{SHORT_LOC}-#{NAME}`.",
				Optional:    true,
			},
			defaultNodashNameFormatProp: schema.StringAttribute{
				Description: "Default format to use for resources which can not have dashes. Defaults to `#{SLUG}#{SHORT_LOC}#{NAME}`.",
				Optional:    true,
			},
			defaultGlobalResourceNameFormatProp: schema.StringAttribute{
				Description: fmt.Sprintf("Default format to use for resources which can have dashes in global scope. Defaults to `%s`.", defaultResourceNameFormatProp),
				Optional:    true,
			},
			defaultGlobalNodashNameFormatProp: schema.StringAttribute{
				Description: fmt.Sprintf("Default format to use for resources which can not have dashes and are in global scope. Defaults to `%s`.", defaultNodashNameFormatProp),
				Optional:    true,
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

	var npConfig shared.NamepConfig

	npConfig.DefaultLocation = config.DefaultLocation.ValueString()
	npConfig.DefaultResourceNameFormat = utils.ValueStringOrDefault(config.DefaultResourceNameFormat, "#{SLUG}-#{SHORT_LOC}-#{NAME}")
	npConfig.DefaultNodashNameFormat = utils.ValueStringOrDefault(config.DefaultNodashNameFormat, "#{SLUG}#{SHORT_LOC}#{NAME}")
	npConfig.DefaultGlobalResourceNameFormat = utils.ValueStringOrDefault(config.DefaultGlobalResourceNameFormat, npConfig.DefaultResourceNameFormat)
	npConfig.DefaultGlobalNodashNameFormat = utils.ValueStringOrDefault(config.DefaultGlobalNodashNameFormat, npConfig.DefaultNodashNameFormat)

	utils.CheckUnknown(sliceStringProp, config.SliceString, &resp.Diagnostics, path.Root(sliceStringProp))

	npConfig.SliceTokens = strings.Fields(config.SliceString.ValueString())
	npConfig.SliceTokensAvailable = len(npConfig.SliceTokens)

	utils.CheckUnknown(extraTokensProp, config.ExtraTokens, &resp.Diagnostics, path.Root(extraTokensProp))

	extra_variables := make(map[string]string, len(config.ExtraTokens.Elements()))
	resp.Diagnostics.Append(config.ExtraTokens.ElementsAs(ctx, &extra_variables, false)...)

	npConfig.ExtraVariables = extra_variables

	azure_resource_formats := make(map[string]string, len(config.AzureResourceFormats.Elements()))
	resp.Diagnostics.Append(config.AzureResourceFormats.ElementsAs(ctx, &azure_resource_formats, false)...)

	npConfig.AzureResourceFormats = azure_resource_formats

	custom_resource_formats := make(map[string]string, len(config.CustomResourceFormats.Elements()))
	resp.Diagnostics.Append(config.CustomResourceFormats.ElementsAs(ctx, &custom_resource_formats, false)...)

	npConfig.CustomResourceFormats = custom_resource_formats

	resp.DataSourceData = npConfig
	resp.ResourceData = npConfig
}

func (p *namepProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		namep.NewAzure,
		namep.NewCustom,
	}
}

func (p *namepProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

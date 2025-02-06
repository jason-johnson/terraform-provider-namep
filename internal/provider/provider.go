package provider

import (
	"context"
	namep "terraform-provider-namep/internal/datasource"
	namepf "terraform-provider-namep/internal/functions"
	"terraform-provider-namep/internal/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
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
	Static types.Bool `tfsdk:"static"`
}

func (p *namepProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "namep"
	resp.Version = p.version
}

func (p *namepProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"static": schema.BoolAttribute{
				Description: "Static flag to determine if all applicable data sources should use static setting, defaults to false.",
				Optional:    true,
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

	npConfig.Static = config.Static.ValueBool()

	resp.DataSourceData = npConfig
	resp.ResourceData = npConfig
}

func (p *namepProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		namep.NewAzureCafTypes,
		namep.NewConfiguration,
		namep.NewAzureLocations,
	}
}

func (p *namepProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *namepProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		namepf.NewNameStringFunction,
	}
}

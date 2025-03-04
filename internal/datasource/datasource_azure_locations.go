package datasource

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-namep/internal/cloud/azure"
	"terraform-provider-namep/internal/shared"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
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
	_ datasource.DataSource                     = &azureLocationsDataSource{}
	_ datasource.DataSourceWithConfigure        = &azureLocationsDataSource{}
	_ datasource.DataSourceWithConfigValidators = &azureLocationsDataSource{}
)

// New is a helper function to simplify the provider implementation.
func NewAzureLocations() datasource.DataSource {
	return &azureLocationsDataSource{}
}

// data source implementation.
type azureLocationsDataSource struct {
	static bool
}

type azureLocationsDataSourceModel struct {
	SubscriptionID   types.String `tfsdk:"subscription_id"`
	SubscriptionName types.String `tfsdk:"subscription_display_name"`
	Static           types.Bool   `tfsdk:"static"`
	LocationMaps     types.Map    `tfsdk:"location_maps"`
}

func (d *azureLocationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_locations"
}

func (d *azureLocationsDataSource) Schema(ctx context.Context, ds datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `This data resource creates a map of maps of variables for locations: [locs](#locs) and [locs_from_display_name](#locs_from_display_name).  The locations will be fetched from the specified (or active if none specified) Azure
subscription unless ` + "`static`" + ` is set to true.
If ` + "`static`" + ` is set to true, the locations that were build with the namep provider will be used.  Note that the static values can get out of date since they cannot be changed without a new version of the provider.  Also note that if ` + "`static`" + ` is
set to true in the provider, it will be used regardless of the value in the data source.  There will, however, be no conflict between the provider ` + "`static`" + ` field and the subscription fields in this datasource.

The main use of this provider is to create these location maps to be passed to the ` + "`variable_maps`" + ` parameter in the [namep_configuration](configuration.md) data source.  Alternatively, it could be assigned to a ` + "`locals`" + ` variable to 
add other maps for the ` + "`variable_maps`" + ` parameter.

## locs

This is a map from the Azure location name (e.g. "eastus") to a short name (e.g. "eus").  The short name is created by changing directions (e.g. "east") to a single letter and countries to their top level domain code (generally
the same as the ISO 3166-1 alpha-2 code).

## locs_from_display_name

This is a map from the lowercase display name of the location (e.g. "east us") to the Azure location name (e.g. "eastus").  This is useful for users that want to use the display name in their configuration but need the Azure location name.
Note this cannot be used to go from display name to short name since the ` + "`namestring`" + ` function does not support double map lookups.

## Common use

These variables are generally for use in formats to put a short form of the location in the computed name.  For example, a variable might be defined called ` + "`LOC`" + ` which will have the azure name of the location of the resource.  The format would then
have ` + "`{LOCS[LOC]}`" + ` present to convert this azure location name to its short form to reduce the size of the name.
		`,
		Attributes: map[string]schema.Attribute{
			"subscription_id": schema.StringAttribute{
				Description: "Subscription ID to pull locations from (cannot be used with `subscription_display_name` or `static`).",
				Required:    false,
				Optional:    true,
			},
			"subscription_display_name": schema.StringAttribute{
				Description: "Subscription Display Name to pull locations from (cannot be used with `subscription_id` or `static`).",
				Required:    false,
				Optional:    true,
			},
			"static": schema.BoolAttribute{
				Description: "Static flag to determine if the data source should be static (cannot be used with `subscription_display_name` or `subscription_id`).",
				Required:    false,
				Optional:    true,
			},
			"location_maps": schema.MapAttribute{
				Description: "Maps of maps for location substitutions, as described above.",
				Computed:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
		},
	}
}

func (d *azureLocationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *azureLocationsDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("subscription_id"),
			path.MatchRoot("subscription_display_name"),
			path.MatchRoot("static"),
		),
	}
}

func (d *azureLocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config azureLocationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	var subscriptionId string
	var locations map[string]map[string]string

	if d.static || config.Static.ValueBool() {
		subscriptionId, locations = createStaticLocationMaps()
	} else {
		subscriptionId, locations = createLocationMaps(ctx, config.SubscriptionID, config.SubscriptionName, &resp.Diagnostics)
	}

	locationMaps, diag := types.MapValueFrom(ctx, types.MapType{ElemType: types.StringType}, locations)

	if diag.HasError() {
		resp.Diagnostics.Append(diag.Errors()...)
	}

	config.SubscriptionID = types.StringValue(subscriptionId)
	config.LocationMaps = locationMaps

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read azure locations data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func createStaticLocationMaps() (string, map[string]map[string]string) {
	locations := make(map[string]map[string]string)
	locations["locs"] = make(map[string]string)
	locations["locs_from_display_name"] = make(map[string]string)

	for k, v := range azure.LocationDefinitions {
		shortName := computeShortName(k)
		locations["locs"][k] = shortName
		displayName := strings.ToLower(v.RegionName)
		locations["locs_from_display_name"][displayName] = k
	}

	return "static", locations
}

func createLocationMaps(ctx context.Context, subscriptionID types.String, subscriptionName types.String, diags *diag.Diagnostics) (string, map[string]map[string]string) {
	locations := make(map[string]map[string]string)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		diags.AddError("Failed to obtain a credential", fmt.Sprintf("failed to obtain a credential: %v", err))
		return "", locations
	}

	subsrId, err := subscriptionId(ctx, cred, subscriptionID, subscriptionName)
	if err != nil {
		diags.AddError("failed to get subscription ID", fmt.Sprintf("failed to get subscription ID: %v", err))
		return subsrId, locations
	}

	locations["locs"] = make(map[string]string)
	locations["locs_from_display_name"] = make(map[string]string)

	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	pager := clientFactory.NewClient().NewListLocationsPager(subsrId, &armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to locations advance page: %v", err)
		}
		for _, v := range page.Value {
			shortName := computeShortName(*v.Name)
			locations["locs"][*v.Name] = shortName
			displayName := strings.ToLower(*v.DisplayName)
			locations["locs_from_display_name"][displayName] = *v.Name
		}
	}

	return subsrId, locations
}

func subscriptionId(ctx context.Context, cred azcore.TokenCredential, subscriptionId types.String, subscriptionName types.String) (string, error) {
	if !subscriptionId.IsNull() {
		return subscriptionId.ValueString(), nil
	}

	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to create client: %v", err))
		return "", err
	}
	pager := clientFactory.NewClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		for _, v := range page.Value {
			if subscriptionName.IsNull() {
				return *v.SubscriptionID, nil
			}
			if *v.DisplayName == subscriptionName.ValueString() {
				return *v.SubscriptionID, nil
			}
		}
	}

	return "", fmt.Errorf("subscription %s not found", subscriptionName.ValueString())
}

func computeShortName(location string) string {
	countryMap := map[string]string{
		"southafrica": "za",
		"europe":      "eu",
		"australia":   "au",
		"sweden":      "se",
		"switzerland": "ch",
		"india":       "in",
		"japan":       "jp",
		"korea":       "kr",
		"brazil":      "br",
		"canada":      "ca",
		"france":      "fr",
		"germany":     "de",
		"norway":      "no",
		"newzealand":  "nz",
		"italy":       "it",
		"poland":      "pl",
		"spain":       "es",
		"mexico":      "mx",
		"israel":      "il",
		"qatar":       "qa",
		"singapore":   "sg",
		"jio":         "j", // Jio is part of the India regions but we just shorten it here
	}

	dirMap := map[string]string{
		"east":    "e",
		"west":    "w",
		"north":   "n",
		"south":   "s",
		"central": "c",
	}

	newName := location
	for k, v := range countryMap {
		newName = strings.Replace(newName, k, v, -1)
	}
	for k, v := range dirMap {
		newName = strings.Replace(newName, k, v, -1)
	}

	return newName
}

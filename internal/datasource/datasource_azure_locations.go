package datasource

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource                     = &azureLocationsDataSource{}
	_ datasource.DataSourceWithConfigValidators = &azureLocationsDataSource{}
)

// New is a helper function to simplify the provider implementation.
func NewAzureLocations() datasource.DataSource {
	return &azureLocationsDataSource{}
}

// data source implementation.
type azureLocationsDataSource struct {
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
		Description: "This data resource fetches types from the Azure CAF project for use in the configuration parameter of `namestring`.",
		Attributes: map[string]schema.Attribute{
			"subscription_id": schema.StringAttribute{
				Description: "Subscription ID to pull locations from (cannot be used with `subscription_display_name`).",
				Required:    false,
				Optional:    true,
			},
			"subscription_display_name": schema.StringAttribute{
				Description: "Subscription Display Name to pull locations from (cannot be used with `subscription_id`).",
				Required:    false,
				Optional:    true,
			},
			"static": schema.BoolAttribute{
				Description: "Static flag to determine if the data source should be static.",
				Required:    false,
				Optional:    true,
			},
			"location_maps": schema.MapAttribute{
				Description: "Maps to support location name substitutions.",
				Computed:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
		},
	}
}

func (d *azureLocationsDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("subscription_id"),
			path.MatchRoot("subscription_display_name"),
		),
	}
}

func (d *azureLocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config azureLocationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to obtain a credential: %v", err))
		return
	}

	subscriptionId, err := subscriptionId(ctx, cred, config.SubscriptionID, config.SubscriptionName)
	if err != nil {
		resp.Diagnostics.AddError("failed to get subscription ID", fmt.Sprintf("failed to get subscription ID: %v", err))
		return
	}

	locations := make(map[string]map[string]string)
	locations["locs"] = make(map[string]string)
	locations["locs_from_display_name"] = make(map[string]string)

	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	pager := clientFactory.NewClient().NewListLocationsPager(subscriptionId, &armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil})
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

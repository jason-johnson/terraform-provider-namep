package datasource_test

import (
	"terraform-provider-namep/internal/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceAzureLocations_empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_azure_locations" "example" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_azure_locations.example",
						tfjsonpath.New("location_maps"),
						knownvalue.MapExact(map[string]knownvalue.Check{
							"locs":                   knownvalue.MapPartial(map[string]knownvalue.Check{}),
							"locs_from_display_name": knownvalue.MapPartial(map[string]knownvalue.Check{}),
						}),
					),
				},
			},
		},
	})
}

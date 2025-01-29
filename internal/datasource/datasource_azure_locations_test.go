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
				Config: `data "namep_configuration" "example" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_configuration.example",
						tfjsonpath.New("configuration"),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"formats":       knownvalue.MapExact(map[string]knownvalue.Check{}),
							"variables":     knownvalue.MapExact(map[string]knownvalue.Check{}),
							"variable_maps": knownvalue.MapExact(map[string]knownvalue.Check{}),
							"types":         knownvalue.MapExact(map[string]knownvalue.Check{}),
						}),
					),
				},
			},
		},
	})
}

package datasource_test

import (
	"regexp"
	"terraform-provider-namep/internal/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceAzureCafTypes_read(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_azure_caf_types" "example" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_azure_caf_types.example",
						tfjsonpath.New("types"),
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"azurerm_resource_group": knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("azurerm_resource_group"),
								"slug": knownvalue.StringExact("rg"),
							}),
						}),
					),
					statecheck.ExpectKnownValue(
						"data.namep_azure_caf_types.example",
						tfjsonpath.New("source"),
						knownvalue.StringRegexp(regexp.MustCompile(`^http`)),
					),
				},
			},
		},
	})
}

func TestAccDataSourceAzureCafTypes_static_provider(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `provider "namep" { static = true }
				data "namep_azure_caf_types" "example" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_azure_caf_types.example",
						tfjsonpath.New("types"),
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"azurerm_resource_group": knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("azurerm_resource_group"),
								"slug": knownvalue.StringExact("rg"),
							}),
						}),
					),
					statecheck.ExpectKnownValue(
						"data.namep_azure_caf_types.example",
						tfjsonpath.New("source"),
						knownvalue.StringExact("static"),
					),
				},
			},
		},
	})
}
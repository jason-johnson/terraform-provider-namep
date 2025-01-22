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

func TestAccDataSourceAzureCafTypes_conflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_azure_caf_types" "example" {
							newest = true
							version = "main"
						}`,
				ExpectError: regexp.MustCompile(`These attributes cannot be configured together\: \[newest,version]`),
			},
		},
	})
}

func TestAccDataSourceAzureCafTypes_read(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_azure_caf_types" "example" {
							newest = true
						}`,
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
				},
			},
		},
	})
}

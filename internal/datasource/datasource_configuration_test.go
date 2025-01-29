package datasource_test

import (
	"terraform-provider-namep/internal/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceConfiguration_empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_configuration" "example" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_configuration.example",
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

func TestAccDataSourceConfiguration_formats(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_configuration" "example" {
				  formats = {
				    "example": "example"
				  }
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_configuration.example",
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

func TestAccDataSourceConfiguration_variables(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_configuration" "example" {
				  variables = {
				    "example": "example"
				  }
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_configuration.example",
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

func TestAccDataSourceConfiguration_variable_maps(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_configuration" "example" {
				  variable_maps = {
				    "example_map": {
				    	"example_key": "example_value"
					}
				  }
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_configuration.example",
						tfjsonpath.New("configuration"),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"variable_maps": knownvalue.MapExact(map[string]knownvalue.Check{
								"example_map": knownvalue.MapExact(map[string]knownvalue.Check{
									"example_key": knownvalue.StringExact("example_value"),
								}),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestAccDataSourceConfiguration_types(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_azure_caf_types" "example" {}

				data "namep_configuration" "example" {
				  types = data.namep_azure_caf_types.example.types
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.namep_configuration.example",
						tfjsonpath.New("configuration"),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"types": knownvalue.MapPartial(map[string]knownvalue.Check{
							"azurerm_resource_group": knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("azurerm_resource_group"),
								"slug": knownvalue.StringExact("rg"),
							}),
						}),
						}),
					),
				},
			},
		},
	})
}
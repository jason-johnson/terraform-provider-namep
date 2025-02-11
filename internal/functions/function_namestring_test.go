package functions_test

import (
	"terraform-provider-namep/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestCustomNameFunction_DelayConfig(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `resource "terraform_data" "test" { input = "test-value" }
						 data "namep_azure_locations" "example" {}
						 data "namep_azure_caf_types" "example" {}
						 data "namep_configuration" "example" {
						   variable_maps = data.namep_azure_locations.example.location_maps
						   types = data.namep_azure_caf_types.example.types
						   formats = {
						   	 azurerm_resource_group = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}"
						     azure_dashes = "#{SLUG}-#{APP}-#{NAME}#{-SALT}"
						   }

						   variables = {
						     name = "main"
						     env = "dev"
						     app = "myapp"
						     salt = terraform_data.test.output
						     loc = "westeurope"
						   }
					     }	
											
				output "test_rg" {
					value = provider::namep::namestring("azurerm_resource_group", data.namep_configuration.example.configuration)
				}
				output "test_kv" {
					value = provider::namep::namestring("azurerm_key_vault", data.namep_configuration.example.configuration)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_rg", knownvalue.StringExact("rg-myapp-dev-weu-main")),
					statecheck.ExpectKnownOutputValue("test_kv", knownvalue.StringExact("kv-myapp-main-test-value")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownOutputValue("test_rg", knownvalue.StringExact("rg-myapp-dev-weu-main")),
						plancheck.ExpectUnknownOutputValue("test_kv"),
					},
				},
			},
		},
	})
}

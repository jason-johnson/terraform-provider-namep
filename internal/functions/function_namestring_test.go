package functions_test

import (
	"fmt"
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

						 locals {
						   config = {
						     variable_maps = data.namep_azure_locations.example.location_maps
							 formats = {
							   azurerm_resource_group = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}"
							   azure_dashes = "#{SLUG}-#{APP}-#{NAME}#{-SALT}"
							 }
							 types = data.namep_azure_caf_types.example.types
							 variables = {
							   name = "main"
							   env = "dev"
							   app = "myapp"
							   loc = "westeurope"
							   salt = terraform_data.test.output
							 }
						   }
						 }
											
				output "test_rg" {
					value = provider::namep::namestring("azurerm_resource_group", local.config)
				}
				output "test_kv" {
					value = provider::namep::namestring("azurerm_key_vault", local.config)
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

const default_config_fmt = `
resource "terraform_data" "test" {
  input = "test-value"
}

locals {
	config = {
	  variable_maps = {
	    locs = {
		  westeurope = "weu"											
		}
	  }
	  variables = {
	    name = "NOT SET"
	    app = "myapp"
	    env = "dev"
	    salt = "uxx1"
	    loc = "westeurope"
		testoutput = resource.terraform_data.test.output
	  }

	  %s

	  types = {
	    azurerm_resource_group = {
		  name = "azurerm_resource_group"
		  slug = "rg"
		  min_length = 1
		  max_length = 90
		  lowercase = true
		  validation_regex = "^[a-z0-9-]*$"
		  default_selector = "azure_dashes_global"
		}
		too_short = {
		  name = "too_short"
		  slug = "ts"
		  min_length = 100
		  max_length = 200
		  lowercase = true
		  validation_regex = "^[a-z0-9-]{100-200}$"
		  default_selector = "azure_dashes_global"
		}
		too_long = {
		  name = "too_long"
		  slug = "tl"
		  min_length = 1
		  max_length = 2
		  lowercase = true
		  validation_regex = "^[a-z0-9-]{1-2}$"
		  default_selector = "azure_dashes_global"
		}
	  }
	}
}
`

var config_with_rg_format_fmt = fmt.Sprintf(default_config_fmt, `formats = {
	    azurerm_resource_group = "#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}#{-SALT}"
}`)

var config_with_default_format_fmt = fmt.Sprintf(default_config_fmt, `formats = {
	azure_dashes_global = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}#{-SALT}"
}`)

var config_with_default_delayed_format_fmt = fmt.Sprintf(default_config_fmt, `formats = {
	azure_dashes_global = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{testoutput}#{-SALT}"
}`)

const config_with_azure_caf_types_fmt = `
data "namep_azure_caf_types" "example" {}

locals {
	config = {
	  variable_maps = {
	    locs = {
		  westeurope = "weu"											
		}
	  }
	  variables = {
	    name = "main"
	    app = "myapp"
	    env = "dev"
	    salt = "uxx1"
	    loc = "westeurope"
	  }

	  formats = {
	  	azure_dashes_subscription = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}#{-SALT}"
	  }

	  types = data.namep_azure_caf_types.example.types
	}
}
`

const config_resolution_types_fmt = `
data "namep_configuration" "example" {
	types = {
		specific_type = {
			name = "specific_type"
			slug = "st"
			min_length = 1
			max_length = 90
			lowercase = true
			validation_regex = "^.*$"
			default_selector = "generic_first_second"
		}
	}
	
	%s
}

output "test" {
	value = provider::namep::namestring("specific_type", data.namep_configuration.example.configuration)
}
`

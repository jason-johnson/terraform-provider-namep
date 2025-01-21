package functions_test

import (
	"regexp"
	"terraform-provider-namep/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestCustomNameFunction_MapArgs(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccFuncCustomName_map_args_custom_rg_fmt,
				Check:  resource.TestCheckOutput("test", "test-value"),
			},
		},
	})
}

// The example implementation does not enable AllowNullValue, however this
// acceptance test shows how to verify the behavior.
func TestCustomNameFunction_Null(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `output "test" {
							value = provider::namep::toname(null, null)
						}`,
				ExpectError: regexp.MustCompile(`Invalid value for "resource_type" parameter: argument must not be null\.`),
			},
		},
	})
}

// The example implementation does not enable AllowUnknownValues, however this
// acceptance test shows how to verify the behavior.
func TestCustomNameFunction_Unknown(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `resource "terraform_data" "test" {
							input = "test-value"
						}
						output "test" {
							value = provider::namep::toname(resource.terraform_data.test.output, {formats = {}, variable_maps = {}, variables = {}, types = {}})
						}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("test-value")),
				},
			},
		},
	})
}

const testAccFuncCustomName_map_args_custom_rg_fmt = `
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
	  }
	  formats = {
	    azurerm_resource_group = "#{APP}-#{ENV}-#{LOCS[LOC]}-#{NAME}#{-SALT}"
	  }

	  types = {
	    azurerm_resource_group = {
		  name = "azurerm_resource_group"
		  slug = "rg"
		  min_length = 1
		  max_length = 90
		  lowercase = true
		  validatation_regex = "^[a-z0-9-]*$"
		  default_selector = "azure_true_global"
		}
	  }
	}
}

output "test" {
    value = provider::namep::toname("azurerm_resource_group", local.config, { name = "mygroup" })
}
`

package functions_test

import (
	"fmt"
	"regexp"
	"terraform-provider-namep/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

// The example implementation does not enable AllowNullValue, however this
// acceptance test shows how to verify the behavior.
func TestCustomNameFunction_Null(t *testing.T) {
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

func TestCustomNameFunction_ResourceGroup(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", config_with_rg_format_fmt, `output "test" {
					value = provider::namep::toname("azurerm_resource_group", local.config, { name = "mygroup" })
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("myapp-dev-weu-mygroup-uxx1")),
				},
			},
		},
	})
}

func TestCustomNameFunction_GlobalFormat(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", config_with_default_format_fmt, `output "test" {
					value = provider::namep::toname("azurerm_resource_group", local.config, { name = "mygroup" })
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("rg-myapp-dev-weu-mygroup-uxx1")),
				},
			},
		},
	})
}

func TestCustomNameFunction_DelayedFormat(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", config_with_default_delayed_format_fmt, `output "test" {
					value = provider::namep::toname("azurerm_resource_group", local.config)
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("rg-myapp-dev-weu-test-value-uxx1")),
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
		  validatation_regex = "^[a-z0-9-]*$"
		  default_selector = "azure_true_global"
		}
	  }
	}
}
`

var config_with_rg_format_fmt = fmt.Sprintf(default_config_fmt, `formats = {
	    azurerm_resource_group = "#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}#{-SALT}"
}`)

var config_with_default_format_fmt = fmt.Sprintf(default_config_fmt, `formats = {
	azure_true_global = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}#{-SALT}"
}`)

var config_with_default_delayed_format_fmt = fmt.Sprintf(default_config_fmt, `formats = {
	azure_true_global = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{testoutput}#{-SALT}"
}`)

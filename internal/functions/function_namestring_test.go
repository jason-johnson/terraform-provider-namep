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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
							value = provider::namep::namestring(null, null)
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
					value = provider::namep::namestring("azurerm_resource_group", local.config, { name = "mygroup" })
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
					value = provider::namep::namestring("azurerm_resource_group", local.config, { NAME = "mygroup" })
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
					value = provider::namep::namestring("azurerm_resource_group", local.config)
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("rg-myapp-dev-weu-test-value-uxx1")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownOutputValue("test"),
					},
				},
			},
		},
	})
}

func TestCustomNameFunction_TooShort(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", config_with_default_format_fmt, `output "test" {
					value = provider::namep::namestring("too_short", local.config, { name = "main" })
				}`),
				ExpectError: regexp.MustCompile(`resulting name is too short`),
			},
		},
	})
}

func TestCustomNameFunction_TooLong(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", config_with_default_format_fmt, `output "test" {
					value = provider::namep::namestring("too_long", local.config, { name = "main" })
				}`),
				ExpectError: regexp.MustCompile(`resulting name is too long`),
			},
		},
	})
}

func TestCustomNameFunction_Bad_Case(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", config_with_default_format_fmt, `output "test" {
					value = provider::namep::namestring("azurerm_resource_group", local.config, { name = "MAIN" })
				}`),
				ExpectError: regexp.MustCompile(`resulting name must be lowercase`),
			},
		},
	})
}

func TestCustomNameFunction_AzureCaf(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", config_with_azure_caf_types_fmt, `output "test" {
					value = provider::namep::namestring("azurerm_resource_group", local.config)
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("rg-myapp-dev-weu-main-uxx1")),
				},
			},
		},
	})
}

func TestCustomNameFunction_Config(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `data "namep_azure_locations" "example" {}
						 data "namep_azure_caf_types" "example" {}
						 data "namep_configuration" "example" {
						   variable_maps = data.namep_azure_locations.example.location_maps
						   types = data.namep_azure_caf_types.example.types
						   formats = {
						     azure_dashes_subscription = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}#{-SALT}"
						   }

						   variables = {
						     name = "main"
						     env = "dev"
						     app = "myapp"
						     salt = "uxx1"
						     loc = "westeurope"
						   }
					     }	
											
				output "test" {
					value = provider::namep::namestring("azurerm_resource_group", data.namep_configuration.example.configuration)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("rg-myapp-dev-weu-main-uxx1")),
				},
			},
		},
	})
}

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
//						plancheck.ExpectKnownOutputValue("test_rg", knownvalue.StringExact("rg-myapp-dev-weu-main")),
						plancheck.ExpectUnknownOutputValue("test_kv"),
					},
				},
			},
		},
	})
}

func TestCustomNameFunction_Resolution_Specific(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config_resolution_types_fmt, `formats = {
					specific_type = "specific_format"
					generic_first_second = "third_form"
					generic_first = "second_form"
					generic = "first_form"
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("specific_format")),
				},
			},
		},
	})
}

func TestCustomNameFunction_Resolution_3rd(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config_resolution_types_fmt, `formats = {
					generic_first_second = "third_form"
					generic_first = "second_form"
					generic = "first_form"
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("third_form")),
				},
			},
		},
	})
}

func TestCustomNameFunction_Resolution_2nd(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config_resolution_types_fmt, `formats = {
					generic_first = "second_form"
					generic = "first_form"
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("second_form")),
				},
			},
		},
	})
}

func TestCustomNameFunction_Resolution_1st(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config_resolution_types_fmt, `formats = {
					generic = "first_form"
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.StringExact("first_form")),
				},
			},
		},
	})
}

func TestCustomNameFunction_Resolution_None(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(config_resolution_types_fmt, `formats = {}`),
				ExpectError: regexp.MustCompile(`No format found`),
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

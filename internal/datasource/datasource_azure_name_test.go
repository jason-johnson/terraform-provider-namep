package datasource_test

import (
	"terraform-provider-namep/internal/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceAzureName_default_dashed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAzureName_default_rg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.namep_azure_name.foo", "result", "rg-weu-mygroup")),
			},
		},
	})
}

func TestAccDataSourceAzureName_default_nodash(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAzureName_default_sa,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.namep_azure_name.foo", "result", "stweumyacct"),
				),
			},
		},
	})
}

func TestAccDataSourceAzureName_custom_rg_fmt(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAzureName_custom_rg_fmt,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.namep_azure_name.rg", "result", "myapp-dev-weu-uxx1-mygroup"),
					resource.TestCheckResourceAttr("data.namep_azure_name.wapp", "result", "app-weu-myapp"),
				),
			},
		},
	})
}

func TestAccDataSourceAzureName_custom_type_fmt(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAzureName_custom_type_fmt,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.namep_azure_name.rg", "result", "myapp-dev-weu-uxx1-mygroup"),
					resource.TestCheckResourceAttr("data.namep_custom_name.custom", "result", "thing-dev-weu-uxx1-mycustom"),
				),
			},
		},
	})
}

func TestAccDataSourceAzureName_override_extra_token(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAzureName_override_extra_token,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.namep_azure_name.rg", "result", "rg-myapp-dev-weu-mygroup"),
					resource.TestCheckResourceAttr("data.namep_azure_name.saa", "result", "unsetmyappdevweusa1"),
					resource.TestCheckResourceAttr("data.namep_azure_name.sab", "result", "staccmyappdevweusa2"),
				),
			},
		},
	})
}

// test configurations

const testAccDataSourceAzureName_default_rg = `
data "namep_azure_name" "foo" {
  name = "mygroup"
  location = "westeurope"
  type = "azurerm_resource_group"
}
`

const testAccDataSourceAzureName_default_sa = `
data "namep_azure_name" "foo" {
  name = "myacct"
  location = "westeurope"
  type = "azurerm_storage_account"
}
`

const testAccDataSourceAzureName_custom_rg_fmt = `
provider "namep" {
  slice_string     = "MYAPP DEV"
  default_location = "westeurope"
  extra_tokens = {
    branch = "uxx1"
  }
  azure_resource_formats = {
    azurerm_resource_group = "#{TOKEN_1}-#{TOKEN_2}-#{SHORT_LOC}#{-BRANCH}-#{NAME}"
  }
}

data "namep_azure_name" "rg" {
  name = "mygroup"
  location = "westeurope"
  type = "azurerm_resource_group"
}

data "namep_azure_name" "wapp" {
  name = "myapp"
  location = "westeurope"
  type = "azurerm_app_service"
}
`

const testAccDataSourceAzureName_custom_type_fmt = `
provider "namep" {
  slice_string     = "MYAPP DEV"
  default_location = "westeurope"
  extra_tokens = {
    branch = "uxx1"
  }
  default_resource_name_format = "#{TOKEN_1}-#{TOKEN_2}-#{SHORT_LOC}#{-BRANCH}-#{NAME}"
  default_nodash_name_format = "#{TOKEN_1}#{TOKEN_2}#{SHORT_LOC}#{BRANCH}#{NAME}"
  custom_resource_formats = {
    my_type = "thing-#{TOKEN_2}-#{SHORT_LOC}#{-BRANCH}-#{NAME}"
  }
}

data "namep_azure_name" "rg" {
  name = "mygroup"
  location = "westeurope"
  type = "azurerm_resource_group"
}

data "namep_custom_name" "custom" {
  name = "mycustom"
  location = "westeurope"
  type = "my_type"
}
`

const testAccDataSourceAzureName_override_extra_token = `
provider "namep" {
  slice_string     = "MYAPP DEV"
  default_location = "westeurope"
  extra_tokens = {
    myslug = "unset"
  }
	default_resource_name_format = "#{SLUG}-#{TOKEN_1}-#{TOKEN_2}-#{SHORT_LOC}-#{NAME}"
	default_nodash_name_format = "#{MYSLUG}#{TOKEN_1}#{TOKEN_2}#{SHORT_LOC}#{NAME}"
}

data "namep_azure_name" "rg" {
  name = "mygroup"
  location = "westeurope"
  type = "azurerm_resource_group"
}

data "namep_azure_name" "saa" {
  name = "sa1"
  location = "westeurope"
  type = "azurerm_storage_account"
}

data "namep_azure_name" "sab" {
	name = "sa2"
	location = "westeurope"
	type = "azurerm_storage_account"
	extra_tokens = {
		myslug = "stacc"
	  }
  }
`

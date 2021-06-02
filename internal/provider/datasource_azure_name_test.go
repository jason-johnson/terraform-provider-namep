package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAzureName(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAzureName,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.namep_azure_name.foo", "name", "mygroup"),
				),
			},
		},
	})
}

const testAccDataSourceAzureName = `
data "namep_azure_name" "foo" {
  name = "mygroup"
	location = "westeurope"
  type = "azurerm_resource_group"
}
`

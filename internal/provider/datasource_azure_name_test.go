package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAzureName(t *testing.T) {
	t.Skip("data source not yet implemented, remove this once you add your own code")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAzureName,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.namep_azure_name.foo", "name", regexp.MustCompile("^my")),
				),
			},
		},
	})
}

const testAccDataSourceAzureName = `
data "namep_azure_name" "foo" {
  name = "mygroup"
  type = "azurerm_resource_group"
}
`

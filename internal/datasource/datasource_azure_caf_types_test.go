package datasource_test

import (
	"regexp"
	"terraform-provider-namep/internal/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceAzureCafTypes_conflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "namep_azure_caf_types" "foo" {
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
				Config: `data "namep_azure_caf_types" "foo" {
							newest = true
						}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.namep_azure_name.rg", "result", "rg-myapp-dev-weu-mygroup"),
				),
			},
		},
	})
}

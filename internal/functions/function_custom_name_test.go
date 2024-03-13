package functions_test

import (
	"regexp"
	"terraform-provider-namep/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestCustomNameFunction_Valid(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
output "test" {
    value = provider::namep::custom_name("test-value")
}
`,
				Check: resource.TestCheckOutput("test", "test-value"),
			},
		},
	})
}

// The example implementation does not return any errors, however
// this acceptance test verifies how the function should behave if it did.
func TestCustomNameFunction_Invalid(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"namep": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
output "test" {
    value = provider::namep::custom_name("invalid")
}
`,
				ExpectError: regexp.MustCompile(`error summary`),
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
				Config: `
output "test" {
    value = provider::namep::custom_name(null)
}
`,
				ExpectError: regexp.MustCompile(`Invalid Function Call`),
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
				Config: `
resource "terraform_data" "test" {
    input = "test-value"
}

output "test" {
    value = provider::namep::custom_name(resource.terraform_data.test.output)
}
`,
				Check: resource.TestCheckOutput("test", "test-value"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownOutputValue("test"),
					},
				},
			},
		},
	})
}

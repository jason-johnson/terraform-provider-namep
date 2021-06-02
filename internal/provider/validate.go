package provider

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Annoyingly there is no way in Go to just ignore the map value type even though we don't use it
func stringInResourceMapKeys(m map[string]ResourceStructure) schema.SchemaValidateDiagFunc {
	mapKeys := make([]string, 0, len(m))
	for k := range m {
		mapKeys = append(mapKeys, k)
	}

	return stringInSlice(mapKeys)
}

func stringInSlice(valid []string) schema.SchemaValidateDiagFunc {
	f := validation.StringInSlice(valid, false)

	return func(v interface{}, path cty.Path) (diags diag.Diagnostics) {

		warnings, errors := f(v, fmt.Sprintf("%s", path))

		for _, warn := range warnings {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  warn,
			})
		}

		for _, error := range errors {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  error.Error(),
			})
		}

		return diags
	}
}

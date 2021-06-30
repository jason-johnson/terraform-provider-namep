package provider

import (
	"fmt"

	"github.com/agext/levenshtein"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Annoyingly there is no way in Go to just ignore the map value type even though we don't use it
func resourceStructureToKeysSlice(m map[string]ResourceStructure) []string {
	mapKeys := make([]string, 0, len(m))
	for k := range m {
		mapKeys = append(mapKeys, k)
	}

	return mapKeys
}

func stringIsValidResourceName(m map[string]ResourceStructure) schema.SchemaValidateDiagFunc {
	valid := resourceStructureToKeysSlice(m)

	return func(v interface{}, path cty.Path) (diags diag.Diagnostics) {

		f := validation.StringInSlice(valid, false)
		warnings, errors := f(v, fmt.Sprintf("%s", path))

		for _, warn := range warnings {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  warn,
			})
		}

		for _, error := range errors {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  error.Error(),
			})
		}

		return diags
	}
}

func NameSuggestion(given string, suggestions []string) string {
	for _, suggestion := range suggestions {
		dist := levenshtein.Distance(given, suggestion, nil)
		if dist < 3 { // threshold determined experimentally
			return suggestion
		}
	}
	return ""
}

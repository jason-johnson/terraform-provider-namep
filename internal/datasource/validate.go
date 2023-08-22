package datasource

import (
	"errors"
	"fmt"

	"github.com/agext/levenshtein"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	return func(i interface{}, path cty.Path) (diags diag.Diagnostics) {

		v, ok := i.(string)
		if !ok {
			err := path.NewError(errors.New("expected type of be string"))

			return diag.FromErr(err)
		}

		for _, str := range valid {
			if v == str {
				return diags
			}
		}

		suggestion := NameSuggestion(v, valid)
		warning := fmt.Sprintf("type %q not found in resources, some variables may not work", v)

		if suggestion != "" {
			warning = fmt.Sprintf("type %q not found in resources, did you mean %q?", v, suggestion)
		}

		return append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       warning,
			AttributePath: path,
		})
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

package datasource

import (
	"context"
	"fmt"
	"terraform-provider-namep/internal/cloud/azure"

	"github.com/agext/levenshtein"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/exp/maps"
)

type stringInAzureResourceMapValidator struct{}

func stringInAzureResourceMap() stringInAzureResourceMapValidator {
	return stringInAzureResourceMapValidator{}
}

func (v stringInAzureResourceMapValidator) Description(ctx context.Context) string {
	return "string must must be present in the defined Azure Resource definitions"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v stringInAzureResourceMapValidator) MarkdownDescription(ctx context.Context) string {
	return "string must must be present in the defined Azure Resource definitions"
}

func (v stringInAzureResourceMapValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	valid := maps.Keys(azure.ResourceDefinitions)
	s := req.ConfigValue.ValueString()

	for _, str := range valid {
		if s == str {
			return
		}
	}

	suggestion := NameSuggestion(s, valid)
	warning := fmt.Sprintf("type %q not found in resources, some variables may not work", v)

	if suggestion != "" {
		warning = fmt.Sprintf("type %q not found in resources, did you mean %q?", v, suggestion)
	}

	resp.Diagnostics.AddAttributeWarning(
		req.Path,
		"Unknown Azure Resource Type",
		warning,
	)
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

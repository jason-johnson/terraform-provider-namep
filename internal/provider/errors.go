package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/diag"

func appendError(diags diag.Diagnostics, message string) diag.Diagnostics {
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  message,
		Detail:   message,
	})
	return diags
}

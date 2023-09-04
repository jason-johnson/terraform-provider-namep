package datasource

import "github.com/hashicorp/terraform-plugin-framework/diag"

type resourceNameCollection interface {
	get(name string) (resourceNameInfo, bool)
}

type resourceNameInfo interface {
	name() string
	allowsDashes() bool
	slug() string
	validateResult(result string, diags *diag.Diagnostics)
}

package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func CheckUnknown(name string, value attr.Value, diags *diag.Diagnostics, path path.Path) {
	if value.IsUnknown() {
		diags.AddAttributeError(
			path,
			fmt.Sprintf("Unknown %s", name),
			fmt.Sprintf("The provider cannot create names as there is an unknown configuration value for the %s. "+
				"Either target apply the source of the value first or set the value statically in the configuration.",
				name),
		)
	}
}

func ValueStringOrDefault(value basetypes.StringValue, defaultValue string) string {
	if value.IsNull() {
		return defaultValue
	}

	return value.ValueString()
}

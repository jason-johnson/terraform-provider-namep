package utils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func CheckUnknowMapValues(ctx context.Context, name string, values types.Map, diags *diag.Diagnostics, path path.Path) {
	m := make(map[string]types.String, len(values.Elements()))
	diags.Append(tfsdk.ValueFrom(ctx, values, types.MapType{ElemType: types.StringType}, &m)...)

	for k, v := range m {
		if v.IsUnknown() {
			diags.AddAttributeError(
				path,
				fmt.Sprintf("Unknown value for %s[%s]", name, k),
				fmt.Sprintf("The provider cannot create names as there is an unknown configuration value for the %s[%s]. "+
					"Either target apply the source of the value first or set the value statically in the configuration.",
					name, k),
			)
		}
	}
}

func ValueStringOrDefault(value basetypes.StringValue, defaultValue string) string {
	if value.IsNull() {
		return defaultValue
	}

	return value.ValueString()
}

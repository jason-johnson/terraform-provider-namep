package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func typesAttributes() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":             types.StringType,
			"slug":             types.StringType,
			"min_length":       types.Int32Type,
			"max_length":       types.Int32Type,
			"lowercase":        types.BoolType,
			"validation_regex": types.StringType,
			"default_selector": types.StringType,
		},
	}
}

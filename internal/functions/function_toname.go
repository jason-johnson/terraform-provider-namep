package functions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &ToNameFunction{}

func NewToNameFunction() function.Function {
	return &ToNameFunction{}
}

type ToNameFunction struct{}

type typeFields struct {
	Name               string `tfsdk:"name"`
	Slug               string `tfsdk:"slug"`
	MinLength          int32  `tfsdk:"min_length"`
	MaxLength          int32  `tfsdk:"max_length"`
	Lowercase          bool   `tfsdk:"lowercase"`
	Validatation_regex string `tfsdk:"validatation_regex"`
	DefaultSelector    string `tfsdk:"default_selector"`
}

func (f *ToNameFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "toname"
}

func (f *ToNameFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Generate an name string based on the resource type and a configuration",
		Description: "This function creates a name for a terraform resource or field.\nThe resulting format will be used based on the the resource type selected and the configuration.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "resource_type",
				Description: "Type of resource to create a name for (required for selecting format, certain variables and perform validation)",
			},
			function.ObjectParameter{
				Name:        "configurations",
				Description: "A configuration object that contains the variables and formats to use for the name.",
				AttributeTypes: map[string]attr.Type{
					"variables":     types.MapType{ElemType: types.StringType},
					"formats":       types.MapType{ElemType: types.StringType},
					"variable_maps": types.MapType{ElemType: types.MapType{ElemType: types.StringType}},
					"types": types.MapType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":               types.StringType,
								"slug":               types.StringType,
								"min_length":         types.Int32Type,
								"max_length":         types.Int32Type,
								"lowercase":          types.BoolType,
								"validatation_regex": types.StringType,
								"default_selector":   types.StringType,
							},
						},
					},
				},
			},
		},
		VariadicParameter: function.MapParameter{
			Name:        "overrides",
			Description: "Variable overrides.  Each argument will be processed in order, overriding the `variables` map which was passed in the configuration parameter.",
			ElementType: types.StringType,
		},

		Return: function.StringReturn{},
	}
}

func (f *ToNameFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var resourceType string

	var configurationsArg struct {
		Variables     *map[string]string              `tfsdk:"variables"`
		Formats       *map[string]string              `tfsdk:"formats"`
		Variable_maps *map[string](map[string]string) `tfsdk:"variable_maps"`
		Types         *map[string]types.Object        `tfsdk:"types"`
	}
	var overridesArg []map[string]string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &resourceType, &configurationsArg, &overridesArg))

	typeInfo := typeFields{
		DefaultSelector:    "custom",
		Validatation_regex: ".*", // No possible validation for default custom names
	}
	for k, o := range *configurationsArg.Types {
		if k == resourceType {
			diag := o.As(ctx, &typeInfo, basetypes.ObjectAsOptions{})
			resp.Error = function.ConcatFuncErrors(function.FuncErrorFromDiags(ctx, diag))
		}
	}
	format, exists := (*configurationsArg.Formats)[resourceType]

	if !exists {
		format, exists = (*configurationsArg.Formats)[typeInfo.DefaultSelector]

		if !exists {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("No format found for resource type '%s' or default selector '%s'", resourceType, typeInfo.DefaultSelector)))
			return
		}
	}

	for _, overrideValue := range overridesArg {
		if overrideValue == nil {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("Got empty map for override"))
			continue
		}

		for k, v := range overrideValue {
			(*configurationsArg.Variables)[k] = v
		}
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, fmt.Sprintf("type: %s, configurationsArg: %v, overridesArg: %v, typeInfo: %v, format: %s", resourceType, configurationsArg, overridesArg, typeInfo, format)))
}

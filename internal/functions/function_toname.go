package functions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &ToNameFunction{}

func NewToNameFunction() function.Function {
	return &ToNameFunction{}
}

type ToNameFunction struct{}

func (f *ToNameFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "toname"
}

func (f *ToNameFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Generate an name string based on the resource type and a standard configuration",
		Description: "This function creates a name for a terraform resource or field.\nThe resulting format will be used based on the the resource type selected and the configuration.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "resource_type",
				Description: "Type of resource to create a name for (required for selecting format, certain variables and perform validation)",
			},
			function.DynamicParameter{
				Name:        "configuration",
				Description: "Function Configuration.  See `Configuration` for more details.",
			},
		},
		VariadicParameter: function.DynamicParameter{
			Name:        "overrides",
			Description: "Function Configuration Overrides.  These should have the same format as the main configuration (specifically the nesting) but only need specify values to be overriden.",
		},

		Return: function.StringReturn{},
	}
}

func (f *ToNameFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var configArg types.Dynamic
	var overridesArgs []types.Dynamic
	var resourceType string
	var name string
	var location string
	extraTokens := map[string]string{}

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &resourceType, &configArg, &overridesArgs))

	switch t := configArg.UnderlyingValue().(type) {
	case types.Object:
		obj := configArg.UnderlyingValue().(types.Object)
		for key, attr := range obj.Attributes() {
			switch key {
			case "name":
				name = attr.String()
			case "location":
				location = attr.String()
			case "extra_tokens":
				switch ett := attr.(type) {
				case types.Object:
					attrObj := attr.(types.Object)
					for key, attr := range attrObj.Attributes() {
						extraTokens[key] = attr.String()
					}
				default:
					resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("extra_tokens expected to be a map, got %T", ett)))
				}
			}
		}
	default:
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("object expected, got %T", t)))
	}

	for _, overrideValue := range overridesArgs {
		if overrideValue.IsNull() || overrideValue.IsUnknown() {
			continue
		}
		switch t := overrideValue.UnderlyingValue().(type) {
		case types.Object:
			obj := overrideValue.UnderlyingValue().(types.Object)
			for key, attr := range obj.Attributes() {
				switch key {
				case "name":
					name = attr.String()
				default:
					resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("unknown key %s", key)))
				}
			}
		default:
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("object expected, got %T", t)))
		}
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, fmt.Sprintf("name: %s, type: %s, location: %s, extra_tokens: %v", name, resourceType, location, extraTokens)))
}

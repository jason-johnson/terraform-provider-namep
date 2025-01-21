package functions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
		Summary:     "Generate an name string based on the resource type and a configuration",
		Description: "This function creates a name for a terraform resource or field.\nThe resulting format will be used based on the the resource type selected and the configuration.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "resource_type",
				Description: "Type of resource to create a name for (required for selecting format, certain variables and perform validation)",
			},
		},
		VariadicParameter: function.DynamicParameter{
			Name:        "configurations",
			Description: "The first of these arguments will be a configuration as described in `Configuration`.  Each additional argument is expected to be a map of strings which will override the `variables` portion of the configuration.",
		},

		Return: function.StringReturn{},
	}
}

func (f *ToNameFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var configsArg []types.Dynamic
	var resourceType string
	var variables map[string]string
	var formats map[string]string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &resourceType, &configsArg))

	for _, configValue := range configsArg {
		if configValue.IsNull() {
			continue
		}
		if configValue.IsUnknown() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("unknown configuration, should be impossible"))
			continue
		}

		switch t := configValue.UnderlyingValue().(type) {
		case types.Object:
			obj := configValue.UnderlyingValue().(types.Object)
			for key, attr := range obj.Attributes() {
				switch key {
				case "variables":
					variables = to_map(attr, resp.Error)
				case "formats":
					formats = to_map(attr, resp.Error)
				default:
					resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("unknown key %s", key)))
				}
			}
		default:
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("object expected, got %T", t)))
		}
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, fmt.Sprintf("type: %s, variables: %v, formats: %v", resourceType, variables, formats)))
}


func to_map(value attr.Value, diag *function.FuncError) map[string]string {
	switch t := value.(type) {
	case types.Object:
		obj := value.(types.Object)
		m := make(map[string]string)
		for key, attr := range obj.Attributes() {
			switch tt := attr.(type) {
			case types.String:
				m[key] = attr.String()
			default:
				diag = function.ConcatFuncErrors(diag, function.NewFuncError(fmt.Sprintf("expected string type for key %s, but got %T", key, tt)))
			}
		}
		return m
	default:
		function.ConcatFuncErrors(diag, function.NewFuncError(fmt.Sprintf("object expected, got %T", t)))
	}
	return nil
}
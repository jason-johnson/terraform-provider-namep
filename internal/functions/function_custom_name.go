package functions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &CustomNameFunction{}

func NewCustomNameFunction() function.Function {
	return &CustomNameFunction{}
}

type CustomNameFunction struct{}

func (f *CustomNameFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "custom_name"
}

func (f *CustomNameFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Generate an custom name",
		Description: "This function creates a name for an azure resource.\nThe format will be used based on the the resource type selected and the appropriate format string.",

		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "params",
				Description: "Function parameters.  Valid keys are 'name', 'type', 'location', and 'extra_tokens'.",
			},
		},

		Return: function.StringReturn{},
	}
}

func (f *CustomNameFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var dynamicArg types.Dynamic
	var resourceType string
	var name string
	var location string
	extraTokens := map[string]string{}
	resourceTypeSeen := false

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &dynamicArg))

	switch t := dynamicArg.UnderlyingValue().(type) {
	case types.String:
		resourceType = t.String()
		resourceTypeSeen = true
	case types.Object:
		obj := dynamicArg.UnderlyingValue().(types.Object)
		for key, attr := range obj.Attributes() {
			switch key {
			case "name":
				name = attr.String()
			case "type":
				resourceType = attr.String()
				resourceTypeSeen = true
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

	if !resourceTypeSeen {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("type is required"))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, fmt.Sprintf("name: %s, type: %s, location: %s, extra_tokens: %v", name, resourceType, location, extraTokens)))
}

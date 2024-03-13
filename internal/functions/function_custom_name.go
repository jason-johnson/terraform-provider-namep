package functions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &CustomNameFunction{}

func NewCustomNameFunction() function.Function {
	return &CustomNameFunction{}
}

type CustomNameFunction struct {
	Name                      types.String `tfsdk:"name"`
	ResourceType              types.String `tfsdk:"type"`
	Location                  types.String `tfsdk:"location"`
	ExtraTokens               types.Map    `tfsdk:"extra_tokens"`
	extra_variables_overrides map[string]string
}

func (f *CustomNameFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "custom_name"
}

func (f *CustomNameFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Generate an custom name",
		Description: "This function creates a name for an azure resource.\nThe format will be used based on the the resource type selected and the appropriate format string.",

		Parameters: []function.Parameter{
			function.ObjectParameter{
				Name:        "arguments",
				Description: "Extra variables for use in format (see Supported Variables) for this data source (may override provider extra_tokens).",
				AttributeTypes: map[string]attr.Type{
					nameProp:        types.StringType,
					typeProp:        types.StringType,
					locationProp:    types.StringType,
					extraTokensProp: types.MapType{ElemType: types.StringType},
				},
			},
		},

		Return: function.StringReturn{},
	}
}

func (f *CustomNameFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var arguments struct {
		Name         types.String `tfsdk:"name"`
		ResourceType types.String `tfsdk:"type"`
		Location     types.String `tfsdk:"location"`
		ExtraTokens  types.Map    `tfsdk:"extra_tokens"`
	}

	// Read Terraform argument data into the variable
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &arguments))

	// Set the result to the same data
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, "it worked"))
}

package functions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &AzureNameFunction{}

type AzureNameFunction struct{}

func (f *AzureNameFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "azure_name"
}

func (f *AzureNameFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Create an azure resource name",
		Description: "This function creates a name for an azure resource.\nThe format will be used based on the the resource type selected and the appropriate format string.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "input",
				Description: "Value to echo",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *AzureNameFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input string

	// Read Terraform argument data into the variable
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))

	// Set the result to the same data
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, input))
}

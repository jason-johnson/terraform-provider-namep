package functions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &NameStringFunction{}

func NewToNameFunction() function.Function {
	return &NameStringFunction{}
}

type NameStringFunction struct{}

type typeFields struct {
	Name              string `tfsdk:"name"`
	Slug              string `tfsdk:"slug"`
	MinLength         int    `tfsdk:"min_length"`
	MaxLength         int    `tfsdk:"max_length"`
	Lowercase         bool   `tfsdk:"lowercase"`
	ValidatationRegex string `tfsdk:"validatation_regex"`
	DefaultSelector   string `tfsdk:"default_selector"`
}

func (f *NameStringFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "namestring"
}

func (f *NameStringFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
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

func (f *NameStringFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var resourceType string

	var configurationsArg struct {
		Variables    *map[string]string              `tfsdk:"variables"`
		Formats      *map[string]string              `tfsdk:"formats"`
		VariableMaps *map[string](map[string]string) `tfsdk:"variable_maps"`
		Types        *map[string]types.Object        `tfsdk:"types"`
	}
	var overridesArg []map[string]string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &resourceType, &configurationsArg, &overridesArg))

	typeInfo := typeFields{
		DefaultSelector:   "custom",
		ValidatationRegex: ".*", // No possible validation for default custom names
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
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("No format found for resource type %q or default format %q", resourceType, typeInfo.DefaultSelector)))
			return
		}
	}

	for _, overrideValue := range overridesArg {
		if overrideValue == nil {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("Got null map for override"))
			continue
		}

		for k, v := range overrideValue {
			(*configurationsArg.Variables)[k] = v
		}
	}

	variables := keysToUpper(*configurationsArg.Variables)

	variableMaps := make(map[string](map[string]string), len(*configurationsArg.VariableMaps))
	for k, v := range *configurationsArg.VariableMaps {
		variableMaps[strings.ToUpper(k)] = keysToUpper(v)
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, setCalculatedName(ctx, typeInfo, format, variables, variableMaps, resp))
}

func setCalculatedName(ctx context.Context, typeInfo typeFields, format string, variables map[string]string, variableMaps map[string](map[string]string), resp *function.RunResponse) *function.FuncError {
	re := regexp.MustCompile(`#\{-?[\w[\]]+-?}`)

	result := re.ReplaceAllStringFunc(format, func(token string) (r string) {
		tl := len(token)
		if tl < 1 {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("bizarre variable received %q", token)))
			return token
		}

		token, prefixDash, postfixDash := preprocessToken(token[2 : tl-1])
		tokenProcessed := true
		var tokenResult string

		if token == "SLUG" {
			tokenResult = typeInfo.Slug
		} else {
			varName, varMapName := variableLocation(token)

			v, varExists := variables[strings.ToUpper(varName)]

			if !varExists {
				resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("No variable found for %q", varName)))
				return token
			}

			if varMapName != "" {
				vm, mapExists := variableMaps[varMapName]

				if !mapExists {
					resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("No variable map found for %q", varMapName)))
					return token
				}

				v, varExists = vm[strings.ToUpper(v)]

				if !varExists {
					resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("No variable found for %q in map %q", varName, varMapName)))
					return token
				}
			}

			tokenResult = v
		}

		if tokenProcessed && len(tokenResult) > 0 {
			if prefixDash {
				tokenResult = string('-') + tokenResult
			} else if postfixDash {
				tokenResult = tokenResult + string('-')
			}
		}

		return tokenResult
	})

	resp.Error = validateResult(result, typeInfo, resp)

	return function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, result))
}

func keysToUpper(m map[string]string) map[string]string {
	newMap := make(map[string]string, len(m))
	for k, v := range m {
		newMap[strings.ToUpper(k)] = v
	}
	return newMap
}

func preprocessToken(token string) (result string, pre bool, post bool) {
	pre = false
	post = false
	result = token
	l := len(token)

	if token[0] == '-' {
		pre = true
		result = token[1:]
	} else if token[l-1] == '-' {
		post = true
		result = token[0 : l-2]
	}

	return result, pre, post
}

func variableLocation(token string) (varName string, varMapName string) {
	re := regexp.MustCompile(`(\w+)\[(\w+)]`)

	matches := re.FindAllStringSubmatch(token, -1)

	if matches != nil {
		return matches[0][2], matches[0][1]
	}

	return token, ""
}

func validateResult(result string, typeInfo typeFields, resp *function.RunResponse) *function.FuncError {
	re := regexp.MustCompile(typeInfo.ValidatationRegex)

	if !re.MatchString(result) {
		if len(result) > typeInfo.MaxLength {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("resulting name is too long (%d > %d): %s", len(result), typeInfo.MaxLength, result)))
		} else if typeInfo.Lowercase && strings.ToLower(result) != result {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("resulting name must be lowercase: %s", result)))
		} else if len(result) < typeInfo.MinLength {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("resulting name is too short (%d < %d): %s", len(result), typeInfo.MinLength, result)))
		} else {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("Resulting name does not match the validation regex (validation regex: %s): %q", typeInfo.ValidatationRegex, result)))
		}
	}

	return resp.Error
}

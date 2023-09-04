package datasource

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"terraform-provider-namep/internal/cloud/azure"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &azureNameDataSource{}
	_ datasource.DataSourceWithConfigure = &azureNameDataSource{}
)

// New is a helper function to simplify the provider implementation.
func NewAzure() datasource.DataSource {
	return &azureNameDataSource{&customNameDataSource{}}
}

// data source implementation.
type azureNameDataSource struct {
	*customNameDataSource
}

func (d *azureNameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_name"
}

func (d *azureNameDataSource) Schema(ctx context.Context, ds datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	d.customNameDataSource.Schema(ctx, ds, resp)

	resp.Schema.Description = "This data resource defines a name for an azure resource.\nThe format will be used based on the the resource type selected and the appropriate format string."

	typeAttr := resp.Schema.Attributes["type"].(schema.StringAttribute)
	resp.Schema.Attributes[typeProp] = schema.StringAttribute{
		Optional:    typeAttr.Optional,
		Description: typeAttr.Description,
		Validators: []validator.String{
			stringInAzureResourceMap(),
		},
	}
}

func (d *azureNameDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.customNameDataSource.Configure(ctx, req, resp)

	ni := make(map[string]resourceNameInfo, len(azure.ResourceDefinitions))
	for name, resource := range azure.ResourceDefinitions {
		ni[name] = resourceStructure{resource}
	}

	d.resourceNameInfoMap = azureResourceNameCollection{ni}
	d.resourceFormats = d.config.AzureResourceFormats
}

type resourceStructure struct {
	azure.ResourceStructure
}

func (r resourceStructure) name() string {
	return r.ResourceTypeName
}

func (r resourceStructure) allowsDashes() bool {
	return r.Dashes
}

func (r resourceStructure) slug() string {
	return r.CafPrefix
}

func (r resourceStructure) validateResult(result string, diags *diag.Diagnostics) {
	errorSeen := false

	if r.LowerCase && strings.ToLower(result) != result {
		diags.AddError("validate", fmt.Sprintf("resulting name must be lowercase: %s", result))
		errorSeen = true
	}

	var validName = regexp.MustCompile(r.ValidationRegExp)

	if !validName.MatchString(result) {

		if len(result) > r.MaxLength {
			diags.AddError("validate", fmt.Sprintf("resulting name is too long (%d > %d): %s", len(result), r.MaxLength, result))
			errorSeen = true
		}

		// NOTE: Regex will generally catch everything but not tell us what's wrong so we only show it if
		// NOTE: nothing else was a problem.  This could hide an error with the string until the other issues are fixed
		if !errorSeen {
			diags.AddError("validate", fmt.Sprintf("resulting name is invalid (validation regex: %s): %s", r.ValidationRegExp, result))
		}
	}
}

type azureResourceNameCollection struct {
	collection map[string]resourceNameInfo
}

func (c azureResourceNameCollection) get(name string) (resourceNameInfo, bool) {
	result, success := c.collection[name]

	return result, success
}

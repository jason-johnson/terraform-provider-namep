---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "namep_azure_caf_types Data Source - terraform-provider-namep"
subcategory: ""
description: |-
  This data resource creates a map of type names to type information.  The types are fetched from the Azure CAF project https://github.com/aztfmod/terraform-provider-azurecaf, unless the static field is true.
  If the static field is true then the types retrieved when this provider was built will be used. Note that the static values can get out of date since they cannot be changed without a new version of the provider.  Also note that if static is
  set to true in the provider, it will be used regardless of the value in the data source.  There will, however, be no conflict between the provider static field and the version field in this datasource (it will be ignored).
  The purpose of this data source is for creating the types to to be passed to the types parameter in the namep_configuration configuration.md data source.  Alternatively, it could be assigned to a locals variable to
  add other types for the types parameter.
  Default Selector
  The defaultSelector for this resource is made up of 3 components: the word "azure", the word "dashes" or "nodashes" (depending on if dashes are allowed in the name of the resource type), and the scope of the resource.
  The main scope to be concerned about is the "global" scope, which means the name must be unique across all of Azure.  The other scopes are "subscription", "resourceGroup", and "resource".  When using the defaultSelector to set
  formats for the resources, it is recommended to use at least the first 2 components (e.g. "azure_dashes") since some names cannot have dashes and should have a different format than those which can.
---

# namep_azure_caf_types (Data Source)

This data resource creates a map of type names to type information.  The types are fetched from the [Azure CAF project](https://github.com/aztfmod/terraform-provider-azurecaf), unless the `static` field is true.
If the `static` field is true then the types retrieved when this provider was built will be used. Note that the static values can get out of date since they cannot be changed without a new version of the provider.  Also note that if `static` is
set to true in the provider, it will be used regardless of the value in the data source.  There will, however, be no conflict between the provider `static` field and the `version` field in this datasource (it will be ignored).

The purpose of this data source is for creating the types to to be passed to the `types` parameter in the [namep_configuration](configuration.md) data source.  Alternatively, it could be assigned to a `locals` variable to 
add other types for the `types` parameter. 

## Default Selector

The `defaultSelector` for this resource is made up of 3 components: the word "azure", the word "dashes" or "nodashes" (depending on if dashes are allowed in the name of the resource type), and the `scope` of the resource.
The main `scope` to be concerned about is the "global" scope, which means the name must be unique across all of Azure.  The other scopes are "subscription", "resourceGroup", and "resource".  When using the `defaultSelector` to set
formats for the resources, it is recommended to use at least the first 2 components (e.g. "azure_dashes") since some names cannot have dashes and should have a different format than those which can.

## Example Usage

```terraform
data "namep_azure_caf_types" "example" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `static` (Boolean) Static flag to determine if the data source should use data retrieved when this data source was built.  If false, the data source will be downloaded from the Azure CAF project.
- `version` (String) The version of the Azure CAF types to fetch.  The newest version will be used if not specified.
							  Possible to specify a branch name, tag name or commit hash (hash must be unique but does not have to be complete).

### Read-Only

- `source` (String) The source URL the Azure CAF types were loaded from.
- `types` (Map of Object) The type info map loaded from the Azure CAF project. (see [below for nested schema](#nestedatt--types))

<a id="nestedatt--types"></a>
### Nested Schema for `types`

Read-Only:

- `default_selector` (String)
- `lowercase` (Boolean)
- `max_length` (Number)
- `min_length` (Number)
- `name` (String)
- `slug` (String)
- `validation_regex` (String)

---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "namep Provider"
subcategory: ""
description: |-
  A provider for creating names for terraform resources.
---

# namep Provider

A provider for creating names for terraform resources via user configured formats.  The provider enforces name format,
length and various other checks so non-compliant configurations fail before any resources are created with the name describing
what the problem is instead of failing with less obvious messages.

## Example Usage

```terraform
provider "namep" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `static` (Boolean) Static flag to determine if all applicable data sources should use static setting, defaults to false.

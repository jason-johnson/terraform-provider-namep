provider "namep" {
  slice_string     = "MYAPP DEV"
  default_location = "westeurope"
  extra_tokens = {
    branch = var.branch_name
  }
  resource_formats = {
    azurerm_resource_group = "#{TOKEN_1}-#{TOKEN_2}-#{SHORT_LOC}#{-BRANCH}-#{NAME}"
  }
}

# NOTE: if branch name is an empty string neither it nor the dash will show up in the name
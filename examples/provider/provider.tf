provider "namep" {
  slice_string     = "MYAPP DEV"
  default_location = "westeurope"
  extra_tokens = {
    branch = "uxx1"
  }
  resource_formats = {
    azurerm_resource_group = "#{TOKEN_1}-#{TOKEN_2}-#{SHORT_LOC}#{-BRANCH}-#{NAME}"
  }
}
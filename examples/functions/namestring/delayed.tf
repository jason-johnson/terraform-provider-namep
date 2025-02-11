resource "random_string" "rnd" {
  length  = 4
  special = false
  upper   = false
}

data "namep_azure_locations" "example" {}

data "namep_azure_caf_types" "example" {}

data "namep_configuration" "example" {
  variable_maps = data.namep_azure_locations.example.location_maps
  types         = data.namep_azure_caf_types.example.types
  formats = {
    azure_dashes_subscription = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}#{-SALT}"
  }

  variables = {
    name = "main"
    env  = "dev"
    app  = "myapp"
    salt = "NOT SET"
    loc  = "westeurope"
  }
}

output "test" {
  value = provider::namep::namestring("azurerm_resource_group", data.namep_configuration.example.configuration, { salt = random_string.rnd.result })
}

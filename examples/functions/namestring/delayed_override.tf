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
    azure_dashes        = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}"
    azure_dashes_global = "#{SLUG}-#{APP}-#{env}-#{LOCS[LOC]}-#{NAME}-#{RND}"
  }

  variables = {
    name = "main"
    env  = "dev"
    app  = "myapp"
    rnd  = "NOT SET"
    loc  = "westeurope"
  }
}

output "test" {
  value = provider::namep::namestring("azurerm_resource_group", data.namep_configuration.example.configuration, { rnd = random_string.rnd.result })
}

#!/bin/bash

# This file needs to be run periodically to update azure resource definitions.  We do check in the files though
# since we have to rebuild if they change anyway

#curl -sLOJ https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/gen.go
#sed -i 's/azurecaf/internal\/provider/g' gen.go
#mv gen.go tools
az account list-locations --query "[?metadata.geographyGroup].{region:metadata.geographyGroup, name:displayName, azName:name}" > tools/azure/data/locationDefinitions.json
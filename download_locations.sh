#!/bin/bash

# This file needs to be run periodically to update azure resource definitions.  We do check in the files though
# since we have to rebuild if they change anyway

#curl -sLOJ https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/gen.go
#sed -i 's/azurecaf/internal\/provider/g' gen.go
#mv gen.go tools
az account list-locations --query "[?metadata.geographyGroup].{region:metadata.geographyGroup, name:displayName, azName:name}" | jq 'INDEX(.name)' > tmp_locations.json
jq 'INDEX(.name)' tools/azure/data/locationShortNames.json > tmp_location_short_names.json
jq -s '.[0] * .[1]|to_entries[]|.value|select(.azName)|select(.short_name_1)' tmp_locations.json tmp_location_short_names.json | jq -s > tools/azure/data/locationDefinitions.json
rm tmp_location_short_names.json tmp_locations.json
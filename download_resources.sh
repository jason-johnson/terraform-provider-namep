#!/bin/bash

# This file needs to be run periodically to update azure resource definitions.  We do check in the files though
# since we have to rebuild if they change anyway

#curl -sLOJ https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/gen.go
#sed -i 's/azurecaf/internal\/provider/g' gen.go
#mv gen.go tools
curl -sLOJ https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/resourceDefinition.json
curl -sLOJ https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/resourceDefinition_out_of_docs.json
mv resourceDefinition*.json tools/data
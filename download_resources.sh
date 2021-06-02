#!/bin/bash

# Download the files to make go gen work

curl -sLOJ https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/gen.go
sed -i 's/azurecaf/internal\/provider/g' gen.go
curl -sLOJ https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/resourceDefinition.json
curl -sLOJ https://raw.githubusercontent.com/aztfmod/terraform-provider-azurecaf/master/resourceDefinition_out_of_docs.json
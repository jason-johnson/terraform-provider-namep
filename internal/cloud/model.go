package cloud

type LocationRecord struct {
	RegionName     string `json:"region"`
	DefinitionName string `json:"name"`
	// Name used in Azure
	AzureName    string `json:"azName"`
}
package cloud

type LocationRecord struct {
	RegionName     string `json:"region"`
	DefinitionName string `json:"name"`
	// Name used in Azure
	AzureName    string `json:"azName"`
	ShortName    string `json:"short_name_1"`
	AltShortName string `json:"short_name_2,omitempty"`
}
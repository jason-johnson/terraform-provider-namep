package shared

type NamepConfig struct {
	SliceTokens                     []string
	SliceTokensAvailable            int
	ExtraVariables                  map[string]string
	DefaultLocation                 string
	DefaultResourceNameFormat       string
	DefaultNodashNameFormat         string
	DefaultGlobalResourceNameFormat string
	DefaultGlobalNodashNameFormat   string
	AzureResourceFormats            map[string]string
	CustomResourceFormats           map[string]string
}

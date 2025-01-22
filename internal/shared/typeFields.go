package shared

type TypeFields struct {
	Name            string `tfsdk:"name"`
	Slug            string `tfsdk:"slug"`
	MinLength       int    `tfsdk:"min_length"`
	MaxLength       int    `tfsdk:"max_length"`
	Lowercase       bool   `tfsdk:"lowercase"`
	ValidationRegex string `tfsdk:"validation_regex"`
	DefaultSelector string `tfsdk:"default_selector"`
}

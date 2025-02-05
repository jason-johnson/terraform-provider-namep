variable "config" {
  type = object({
    variables     = map(string)
    variable_maps = map(map(string))
    formats       = map(string)
    types = map(object({
      name             = string
      slug             = string
      min_length       = number
      max_length       = number
      lowercase        = bool
      validation_regex = string
      default_selector = string
    }))
  })
}

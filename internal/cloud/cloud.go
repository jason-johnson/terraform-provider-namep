package cloud

type Cloud interface {
	// Get a shortened name of given location
	short_location(location string) (string, bool)
	// Get an alternate shortened name of given location
	alt_short_location(location string) (string, bool)
	// Get an official slug for a given resource type
	slug(resourceType string) (string, bool)
}

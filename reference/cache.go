package reference

// resolvedFile cache
type resolvedFile struct {
	// File location
	file string
	// Imports resolved within the file
	imports map[string]string
	// Types resolved within the file
	types map[string]typeRef
}

// typeRef cache
type typeRef struct {
	// File containing this type
	file string
	// Type struct body content
	content string
}

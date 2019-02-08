package misc

// StringInSlice check
func StringInSlice(needle string, haystack []string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}

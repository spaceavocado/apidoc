package token

import "fmt"

type dic struct {
	mapping map[int]string
	before  func([]string) []string
	after   func(map[string]string) map[string]string
}

// Map given list of entries into Token meta collection
// based on the underlaying mapping dictionary
func (d *dic) Map(entries []string) map[string]string {
	// Pre processing
	if d.before != nil {
		entries = d.before(entries)
	}
	meta := map[string]string{}
	last := ""
	for i, e := range entries {
		key, ok := d.mapping[i]
		if ok {
			meta[key] = e
			last = key
		} else {
			meta[last] += fmt.Sprintf(" %s", e)
		}
	}
	// Post processing
	if d.after != nil {
		meta = d.after(meta)
	}
	return meta
}

package app

import (
	"errors"

	"github.com/spaceavocado/apidoc/token"
)

type subrouter struct {
	name      string
	url       string
	subrouter string
}

// ResolveSubrouters within the endpoints
// * Reduce the endpoints by subrouter blocks
// * Resolve subrouter url recursively
// * Update endpoint router URL
func resolveSubrouters(endpoints [][]token.Token) ([][]token.Token, error) {
	// Find and reduce subrouters
	filtered := make([][]token.Token, 0, len(endpoints))
	subs := make([]subrouter, 0)
	for _, tokens := range endpoints {
		sub := subrouter{}
		size := len(tokens)
		keys := 0
		if size > 1 && size < 4 {
			for _, t := range tokens {
				if t.Key == "router" && t.Meta["method"] == "" && sub.name == "" {
					sub.name = t.Meta["url"]
					keys++
				} else if t.Key == "routerurl" && sub.url == "" {
					sub.url = t.Meta["value"]
					keys++
				} else if t.Key == "subrouter" && sub.subrouter == "" {
					sub.subrouter = t.Meta["value"]
					keys++
				}
			}
			if keys == size {
				subs = append(subs, sub)
				continue
			}
		}
		filtered = append(filtered, tokens)
	}

	// Resolve subrouters
	subrouters := make(map[string]string, 0)
	var err error
	for _, s := range subs {
		_, err = subrouterTree(subs, s.name, subrouters, 0)
		if err != nil {
			break
		}
	}
	if err != nil {
		return filtered, err
	}

	// Update endpoint router URLs
	for _, e := range filtered {
		sub := ""
		ri := -1
		for i, t := range e {
			if t.Key == "router" {
				ri = i
			} else if t.Key == "subrouter" {
				sub = t.Meta["value"]
			}
		}
		if ri > -1 && sub != "sub" {
			if subURL, ok := subrouters[sub]; ok {
				e[ri].Meta["url"] = subURL + e[ri].Meta["url"]
			}
		}
	}

	return filtered, nil
}

// SubrouterTree recursive resolver
// A subrouter can have an parent so this method is reconstructing the full URL.
// The cycling is soft-locked on 10 inner jumps
func subrouterTree(subs []subrouter, needle string, m map[string]string, depth int) (string, error) {
	if depth >= 10 {
		return "", errors.New("infinite subrouter parent cycling detected")
	}
	if url, ok := m[needle]; ok {
		return url, nil
	}
	for _, s := range subs {
		if s.name == needle {
			if s.subrouter == "" {
				m[needle] = s.url
			} else if url, ok := m[s.subrouter]; ok {
				m[needle] = url + s.url
			} else {
				url, err := subrouterTree(subs, s.subrouter, m, depth+1)
				if err != nil {
					return "", err
				}
				m[needle] = url + s.url
			}
			return m[needle], nil
		}
	}
	return "", nil
}

package app

import (
	"testing"

	"github.com/spaceavocado/apidoc/token"
)

func TestResolveSubrouters(t *testing.T) {
	var test [][]token.Token
	var res [][]token.Token
	var err error

	// Expected
	test = [][]token.Token{
		{
			{Key: "router", Meta: map[string]string{
				"url": "/add", "method": "[get]",
			}},
			{Key: "subrouter", Meta: map[string]string{
				"value": "a",
			}},
		},
		{
			{Key: "router", Meta: map[string]string{
				"url": "b", "method": "",
			}},
			{Key: "subrouter", Meta: map[string]string{
				"value": "c",
			}},
			{Key: "routerurl", Meta: map[string]string{
				"value": "/person",
			}},
		},
		{
			{Key: "router", Meta: map[string]string{
				"url": "c", "method": "",
			}},
			{Key: "routerurl", Meta: map[string]string{
				"value": "/base",
			}},
		},
		{
			{Key: "router", Meta: map[string]string{
				"url": "a", "method": "",
			}},
			{Key: "subrouter", Meta: map[string]string{
				"value": "b",
			}},
			{Key: "routerurl", Meta: map[string]string{
				"value": "/cat",
			}},
		},
	}

	res, err = resolveSubrouters(test)
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
		return
	}
	if len(res) != 1 {
		t.Errorf("Has %d, expected %d reduced endpoints", len(res), 1)
		return
	}
	if res[0][0].Meta["url"] != "/base/person/cat/add" {
		t.Errorf("Has %s, expected %s reduced endpoints", res[0][0].Meta["url"], "/base/person/cat/add")
		return
	}

	// False positive subrouters
	test = [][]token.Token{
		{
			{Key: "router", Meta: map[string]string{
				"url": "/add", "method": "[get]",
			}},
			{Key: "subrouter", Meta: map[string]string{
				"value": "a",
			}},
		},
		{
			{Key: "router", Meta: map[string]string{
				"url": "c", "method": "",
			}},
			{Key: "routerurl", Meta: map[string]string{
				"value": "/base",
			}},
			{Key: "other", Meta: map[string]string{
				"value": "/base",
			}},
		},
		{
			{Key: "subrouter", Meta: map[string]string{
				"value": "b",
			}},
		},
	}

	res, err = resolveSubrouters(test)
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
		return
	}
	if len(res) != 3 {
		t.Errorf("Has %d, expected %d reduced endpoints", len(res), 3)
		return
	}
	if res[0][0].Meta["url"] != "/add" {
		t.Errorf("Has %s, expected %s reduced endpoints", res[0][0].Meta["url"], "/add")
		return
	}

	// Cycling
	test = [][]token.Token{
		{
			{Key: "router", Meta: map[string]string{
				"url": "/add", "method": "[get]",
			}},
			{Key: "subrouter", Meta: map[string]string{
				"value": "a",
			}},
		},
		{
			{Key: "router", Meta: map[string]string{
				"url": "b", "method": "",
			}},
			{Key: "subrouter", Meta: map[string]string{
				"value": "c",
			}},
			{Key: "routerurl", Meta: map[string]string{
				"value": "/person",
			}},
		},
		{
			{Key: "router", Meta: map[string]string{
				"url": "c", "method": "",
			}},
			{Key: "routerurl", Meta: map[string]string{
				"value": "/base",
			}},
			{Key: "subrouter", Meta: map[string]string{
				"value": "b",
			}},
		},
	}

	_, err = resolveSubrouters(test)
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}
}

func TestSubrouterTree(t *testing.T) {
	var test []subrouter
	var m map[string]string
	var err error

	// Valid
	test = []subrouter{
		{"B", "/person", "C"},
		{"C", "/base", ""},
		{"A", "/cat", "B"},
	}
	m = make(map[string]string, 0)
	for _, s := range test {
		_, err = subrouterTree(test, s.name, m, 0)
		if err != nil {
			break
		}
	}
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if m["A"] != "/base/person/cat" {
		t.Errorf("Has %s, expected %s", m["A"], "/base/person/cat")
	}
	if m["B"] != "/base/person" {
		t.Errorf("Has %s, expected %s", m["B"], "/base/person")
	}
	if m["C"] != "/base" {
		t.Errorf("Has %s, expected %s", m["C"], "/base")
	}

	// Cycling prevention
	test = []subrouter{
		{"B", "/person", "C"},
		{"C", "/base", "B"},
		{"A", "/cat", "B"},
	}
	m = make(map[string]string, 0)
	for _, s := range test {
		_, err = subrouterTree(test, s.name, m, 0)
		if err != nil {
			break
		}
	}
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	// Invalid needle
	test = []subrouter{
		{"B", "/person", "C"},
		{"C", "/base", ""},
		{"A", "/cat", "B"},
	}
	m = make(map[string]string, 0)
	for _, s := range test {
		_, err = subrouterTree(test, s.name+"sufix", m, 0)
		if err != nil {
			break
		}
	}
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
}

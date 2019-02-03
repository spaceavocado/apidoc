package token

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spaceavocado/apidoc/extract"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestParse(t *testing.T) {
	p := NewParser(false).(*parser)
	p.typeMapping["missingdict"] = 20

	tests := []extract.Block{
		// Valid
		extract.Block{
			Lines: []string{
				"param id path {int} true Hello World",
			},
		},
		// Invalid
		extract.Block{
			Lines: []string{
				"",
			},
		},
		// Unknown token
		extract.Block{
			Lines: []string{
				"unknown id path {int} true Hello World",
			},
		},
		// Missing dic
		extract.Block{
			Lines: []string{
				"missingdict id path {int} true Hello World",
			},
		},
		// Post processing
		extract.Block{
			Lines: []string{
				"success 200 {string} Comment",
			},
		},
	}

	tokens, err := p.Parse(tests[0])
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(tokens) != 1 {
		t.Errorf("Has %d tokens, expected %d tokens", len(tokens), 1)
		return
	}

	tokens, err = p.Parse(tests[1])
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(tokens) != 0 {
		t.Errorf("Has %d tokens, expected %d tokens", len(tokens), 0)
		return
	}

	tokens, err = p.Parse(tests[2])
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(tokens) != 0 {
		t.Errorf("Has %d tokens, expected %d tokens", len(tokens), 0)
		return
	}

	tokens, err = p.Parse(tests[3])
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(tokens) != 0 {
		t.Errorf("Has %d tokens, expected %d tokens", len(tokens), 0)
		return
	}

	tokens, err = p.Parse(tests[4])
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(tokens) != 1 {
		t.Errorf("Has %d tokens, expected %d tokens", len(tokens), 1)
		return
	}
	if tokens[0].Meta["desc"] != "Comment" {
		t.Errorf("Has \"%s\", expected \"%s\"", tokens[0].Meta["desc"], "Comment")
		return
	}
}

func TestVerbose(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	hook := test.NewGlobal()

	p := NewParser(true).(*parser)
	p.typeMapping["missingdict"] = 20

	test := extract.Block{
		Lines: []string{
			"",
			"unknown id path {int} true Hello World",
			"missingdict id path {int} true Hello World",
		},
	}

	expected := []string{
		"tokenization: cannot tokenize this line:",
		"tokenization: unknown token type:",
		"tokenization: missing token meta dic for type:",
	}

	p.Parse(test)
	if len(hook.Entries) != 3 {
		t.Errorf("Has %d logs, expected %d logs", len(hook.Entries), 3)
		return
	}

	for i, e := range hook.Entries {
		o, err := e.String()
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			return
		}
		if strings.Contains(o, expected[i]) == false {
			t.Errorf("Has %s, expected %s", o, expected[i])
		}
	}
}

// Dictionary mapping
func TestMap(t *testing.T) {
	d := dic{
		mapping: map[int]string{
			0: "key",
			1: "value",
		},
	}

	test := [][]string{
		[]string{
			"tag",
			"path",
		},
		[]string{
			"tag",
			"path",
			"{object}",
		},
	}

	m := d.Map(test[0])
	if len(m) != 2 {
		t.Errorf("Has %d keys, expected %d keys", len(m), 5)
		return
	}
	if m["key"] != test[0][0] {
		t.Errorf("Has \"%s\", expected \"%s\"", m["key"], test[0][0])
		return
	}
	if m["value"] != test[0][1] {
		t.Errorf("Has \"%s\", expected \"%s\"", m["value"], test[0][1])
		return
	}

	m = d.Map(test[1])
	if len(m) != 2 {
		t.Errorf("Has %d keys, expected %d keys", len(m), 5)
		return
	}

	expected := fmt.Sprintf("%s %s", test[1][1], test[1][2])
	if m["value"] != expected {
		t.Errorf("Has \"%s\", expected \"%s\"", m["value"], expected)
		return
	}

	// Preprocessing
	d = dic{
		before: func(input []string) []string {
			out := make([]string, len(input))
			for i, v := range input {
				out[i] = v + "_postfix"
			}
			return out
		},
		mapping: map[int]string{
			0: "key",
			1: "value",
		},
	}
	m = d.Map(test[0])
	if len(m) != 2 {
		t.Errorf("Has %d keys, expected %d keys", len(m), 5)
		return
	}

	expected = fmt.Sprintf("%s_postfix", test[0][0])
	if m["key"] != expected {
		t.Errorf("Has \"%s\", expected \"%s\"", m["key"], expected)
		return
	}

	expected = fmt.Sprintf("%s_postfix", test[0][1])
	if m["value"] != expected {
		t.Errorf("Has \"%s\", expected \"%s\"", m["value"], expected)
		return
	}

	// Postprocessing
	d = dic{
		after: func(input map[string]string) map[string]string {
			out := make(map[string]string, 0)
			for k, v := range input {
				out[k] = v + "_postfix"
			}
			return out
		},
		mapping: map[int]string{
			0: "key",
			1: "value",
		},
	}
	m = d.Map(test[0])
	if len(m) != 2 {
		t.Errorf("Has %d keys, expected %d keys", len(m), 5)
		return
	}

	expected = fmt.Sprintf("%s_postfix", test[0][0])
	if m["key"] != expected {
		t.Errorf("Has \"%s\", expected \"%s\"", m["key"], expected)
		return
	}

	expected = fmt.Sprintf("%s_postfix", test[0][1])
	if m["value"] != expected {
		t.Errorf("Has \"%s\", expected \"%s\"", m["value"], expected)
		return
	}
}

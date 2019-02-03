package reference

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spaceavocado/apidoc/extract"
)

func TestResolve(t *testing.T) {
	r := NewResolver(false).(*resolver)

	// Nothing to resolve
	err := r.Resolve([]extract.Block{
		{
			Lines: []string{""},
		},
	})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	// Resolved body
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				types: map[string]typeRef{
					"person": {
						file:    "github.com/pkg/response/tmp.go",
						content: "",
					},
				},
			},
		},
	}
	r.types = map[string][]string{
		"github.com/pkg/response.person": {
			"Name {string} true Description",
		},
	}

	blocks := []extract.Block{
		{
			File: "github.com/pkg/response/tmp.go",
			Lines: []string{
				"body person",
			},
		},
	}
	err = r.Resolve(blocks)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(blocks[0].Lines) != 2 {
		t.Errorf("Expected %d log entries, got %d", 2, len(blocks[0].Lines))
	}
	if blocks[0].Lines[0] != "bref github.com/pkg/response.person Name {string} true Description" {
		t.Errorf("Expected \"%s\", got \"%s\"", "bref github.com/pkg/response.person Name {string} true Description", blocks[0].Lines[0])
	}
	if blocks[0].Lines[1] != "body github.com/pkg/response.person" {
		t.Errorf("Expected \"%s\", got \"%s\"", "body github.com/pkg/response.person", blocks[0].Lines[1])
	}

	// Resolved body, array
	blocks = []extract.Block{
		{
			File: "github.com/pkg/response/tmp.go",
			Lines: []string{
				"body []person",
			},
		},
	}
	err = r.Resolve(blocks)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(blocks[0].Lines) != 2 {
		t.Errorf("Expected %d log entries, got %d", 2, len(blocks[0].Lines))
	}
	if blocks[0].Lines[0] != "bref github.com/pkg/response.person Name {string} true Description" {
		t.Errorf("Expected \"%s\", got \"%s\"", "bref github.com/pkg/response.person Name {string} true Description", blocks[0].Lines[0])
	}
	if blocks[0].Lines[1] != "body []github.com/pkg/response.person" {
		t.Errorf("Expected \"%s\", got \"%s\"", "body []github.com/pkg/response.person", blocks[0].Lines[1])
	}

	// Invalid ref
	blocks = []extract.Block{
		{
			File: "github.com/pkg/response/tmp.go",
			Lines: []string{
				"body ia_s",
			},
		},
	}
	err = r.Resolve(blocks)
	if err == nil {
		t.Errorf("Expected error got nil")
		return
	}

	// Resolved response
	blocks = []extract.Block{
		{
			File: "github.com/pkg/response/tmp.go",
			Lines: []string{
				"success 200 {object} person",
				"failure 500 {object} person",
				"failure 500 {string} Description",
			},
		},
	}
	err = r.Resolve(blocks)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(blocks[0].Lines) != 5 {
		t.Errorf("Expected %d log entries, got %d", 5, len(blocks[0].Lines))
	}
	if blocks[0].Lines[0] != "sref github.com/pkg/response.person Name {string} true Description" {
		t.Errorf("Expected \"%s\", got \"%s\"", "sref github.com/pkg/response.person Name {string} true Description", blocks[0].Lines[0])
	}
	if blocks[0].Lines[1] != "success 200 {object} github.com/pkg/response.person" {
		t.Errorf("Expected \"%s\", got \"%s\"", "success 200 {object} github.com/pkg/response.person", blocks[0].Lines[1])
	}
	if blocks[0].Lines[2] != "fref github.com/pkg/response.person Name {string} true Description" {
		t.Errorf("Expected \"%s\", got \"%s\"", "fref github.com/pkg/response.person Name {string} true Description", blocks[0].Lines[2])
	}
	if blocks[0].Lines[3] != "failure 500 {object} github.com/pkg/response.person" {
		t.Errorf("Expected \"%s\", got \"%s\"", "failure 500 {object} github.com/pkg/response.person", blocks[0].Lines[3])
	}
	if blocks[0].Lines[4] != "failure 500 {string} Description" {
		t.Errorf("Expected \"%s\", got \"%s\"", "failure 500 {string} Description", blocks[0].Lines[4])
	}

	// Resolved wrappers
	blocks = []extract.Block{
		{
			File: "github.com/pkg/response/tmp.go",
			Lines: []string{
				"swrap person Name",
			},
		},
	}
	err = r.Resolve(blocks)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(blocks[0].Lines) != 2 {
		t.Errorf("Expected %d log entries, got %d", 2, len(blocks[0].Lines))
	}
	if blocks[0].Lines[0] != "swrapref github.com/pkg/response.person Name {string} true true Description" {
		t.Errorf("Expected \"%s\", got \"%s\"", "swrapref github.com/pkg/response.person Name {string} true true Description", blocks[0].Lines[0])
	}
	if blocks[0].Lines[1] != "swrap github.com/pkg/response.person Name" {
		t.Errorf("Expected \"%s\", got \"%s\"", "swrap github.com/pkg/response.person Name", blocks[0].Lines[1])
	}

	// Fail to resolve the reference
	r.packages = map[string]map[string]resolvedFile{
		"x": {
			"x.go": {
				types: map[string]typeRef{},
			},
		},
	}
	r.types = map[string][]string{
		"x": {""},
	}
	err = r.Resolve(blocks)
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	// Root object with own props marked as pointer
	// Resolved body
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				types: map[string]typeRef{
					"person": {
						file:    "github.com/pkg/response/tmp.go",
						content: "",
					},
				},
			},
		},
	}
	r.types = map[string][]string{
		"github.com/pkg/response.person": {
			"Name {string} true Description",
			"Data {object}",
			"Data.field1 {string} false Description",
		},
	}
	blocks = []extract.Block{
		{
			File: "github.com/pkg/response/tmp.go",
			Lines: []string{
				"swrap person Data",
			},
		},
	}
	err = r.Resolve(blocks)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(blocks[0].Lines) != 4 {
		t.Errorf("Expected %d log entries, got %d", 4, len(blocks[0].Lines))
	}
	if blocks[0].Lines[1] != "swrapref github.com/pkg/response.person Data {object} true true" {
		t.Errorf("Expected \"%s\", got \"%s\"", "swrapref github.com/pkg/response.person Data {object} true true", blocks[0].Lines[1])
	}
}

func TestHasExpectedPrefix(t *testing.T) {
	r := NewResolver(false).(*resolver)
	m := r.HasExpectedPrefix("body response.Something")
	if m == (mappingType{}) {
		t.Errorf("Expecting mapping, got nothing")
		return
	}
	if m.prefix != r.prefixMapping["body"].prefix {
		t.Errorf("Expected \"%s\", got \"%s\"", r.prefixMapping["body"].prefix, m.prefix)
	}

	m = r.HasExpectedPrefix("unknown response.Something")
	if m != (mappingType{}) {
		t.Errorf("Expecting nil, got \"%s\"", m.prefix)
		return
	}
}

func TestAddPrefix(t *testing.T) {
	r := NewResolver(false).(*resolver)
	items := []string{"a", "b"}
	r.AddPrefix("prefix_", items)
	for _, e := range items {
		if strings.HasPrefix(e, "prefix_") == false {
			t.Errorf("Invalid prefix, got \"%s\"", e)
			return
		}
	}
}

func TestPkgName(t *testing.T) {
	r := NewResolver(false).(*resolver)
	res := r.PkgName("github.com/pkg/name")
	if res != "github.com\\pkg" {
		t.Errorf("Expected \"%s\", got \"%s\"", "github.com\\pkg", res)
		return
	}
	res = r.PkgName("github.com\\pkg\\name")
	if res != "github.com\\pkg" {
		t.Errorf("Expected \"%s\", got \"%s\"", "github.com\\pkg\\name", res)
		return
	}
}

func TestNormalizePkgName(t *testing.T) {
	r := NewResolver(false).(*resolver)

	tests := []string{
		r.gopath + "/github.com/pkg",
		r.gopath + "\\github.com\\pkg",
		"/github.com/pkg",
	}

	expected := "github.com/pkg"
	for _, c := range tests {
		res := r.NormalizePkgName(c)
		if res != expected {
			t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
			return
		}
	}
}

func TestResolveReference(t *testing.T) {
	r := NewResolver(false).(*resolver)

	// Invalid file
	_, err := r.ResolveReference("", "not-existing/response.go", 0)
	if err == nil {
		t.Errorf("Expected error got nil")
		return
	}

	// Local ref
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				types: map[string]typeRef{
					"person": {
						file:    "github.com/pkg/response/tmp.go",
						content: "",
					},
				},
			},
		},
	}
	r.types = map[string][]string{
		"github.com/pkg/response.person": {
			"Name {string} true Description",
		},
	}
	res, err := r.ResolveReference(
		"person",
		"github.com/pkg/response/tmp.go",
		0,
	)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if res[0] != "github.com/pkg/response.person Name {string} true Description" {
		t.Errorf("Expected \"%s\", got \"%s\"", "github.com/pkg/response.person Name {string} true Description", res[0])
		return
	}

	// External ref, invalid 1
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				types: map[string]typeRef{
					"person": {
						file:    "github.com/pkg/response/tmp.go",
						content: "",
					},
				},
			},
		},
	}
	r.types = map[string][]string{
		"github.com/pkg/response.person": {
			"Name {string} true Description",
		},
	}
	_, err = r.ResolveReference(
		"other.Object",
		"github.com/pkg/response/tmp.go",
		0,
	)
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	// External ref, invalid 2
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				imports: map[string]string{
					"other": ".....",
				},
				types: map[string]typeRef{
					"person": {
						file:    "github.com/pkg/response/tmp.go",
						content: "",
					},
				},
			},
		},
	}
	r.types = map[string][]string{
		"github.com/pkg/response.person": {
			"Name {string} true Description",
		},
	}
	_, err = r.ResolveReference(
		"other.Object",
		"github.com/pkg/response/tmp.go",
		0,
	)
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	// External resolved
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				imports: map[string]string{
					"other": "github.com/pkg/response",
				},
				types: map[string]typeRef{
					"Object": {
						file:    "github.com/pkg/response/tmp.go",
						content: "",
					},
				},
			},
		},
	}
	r.types = map[string][]string{
		"github.com/pkg/response.Object": {
			"Name {string} true Description",
		},
	}
	res, err = r.ResolveReference(
		"other.Object",
		"github.com/pkg/response/tmp.go",
		0,
	)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if res[0] != "github.com/pkg/response.Object Name {string} true Description" {
		t.Errorf("Expected \"%s\", got \"%s\"", "github.com/pkg/response.Object Name {string} true Description", res[0])
		return
	}
}

func TestPkgLoc(t *testing.T) {
	r := NewResolver(false).(*resolver)

	// Existing
	loc, err := r.PkgLoc("response", resolvedFile{
		imports: map[string]string{
			"response": "github.com/project/response",
		},
	})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if loc != "github.com/project/response" {
		t.Errorf("Expected \"%s\", got \"%s\"", "github.com/project/response", loc)
		return
	}

	// Missing
	_, err = r.PkgLoc("other", resolvedFile{
		imports: map[string]string{
			"response": "github.com/project/response",
		},
	})
	if err == nil {
		t.Errorf("Expected error got nil")
		return
	}

	// Verbose
	b := &bytes.Buffer{}
	log.SetOutput(b)
	hook := test.NewGlobal()
	r = NewResolver(true).(*resolver)

	r.PkgLoc("other", resolvedFile{
		imports: map[string]string{
			"response": "github.com/project/response",
		},
	})
	if len(hook.Entries) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(hook.Entries))
	}
	o, err := hook.Entries[0].String()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if strings.Contains(o, "reference resolving: cannot resolve the location of package") == false {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "reference resolving: cannot resolve the location of package", o)
	}
}

func TestReferenceDetails(t *testing.T) {
	r := NewResolver(false).(*resolver)

	// Missing package
	_, err := r.ReferenceDetails("", "x", "", 0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Missing type, verbose
	b := &bytes.Buffer{}
	log.SetOutput(b)
	hook := test.NewGlobal()
	r = NewResolver(true).(*resolver)
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				types: map[string]typeRef{
					"other": {},
				},
			},
		},
	}
	_, err = r.ReferenceDetails("github.com/pkg/response/tmp.go", "github.com/pkg/response", "some", 0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if len(hook.Entries) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(hook.Entries))
	}
	o, err := hook.Entries[0].String()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if strings.Contains(o, "reference resolving: unknown type") == false {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "reference resolving: unknown type", o)
	}

	// Verbose missing package
	hook.Reset()
	r = NewResolver(true).(*resolver)
	_, err = r.ReferenceDetails("", "x", "", 0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if len(hook.Entries) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(hook.Entries))
	}
	o, err = hook.Entries[0].String()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if strings.Contains(o, "reference resolving: unknown package") == false {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "reference resolving: unknown package", o)
	}
}

func TestTypeToParams(t *testing.T) {
	r := NewResolver(false).(*resolver)

	// Cache, no changes, i.e. depth over 0
	r.types = map[string][]string{
		"github.com/pkg/response": {
			"github.com/pkg/response Name {string} true Description",
		},
	}
	res := r.TypeToParams(
		"github.com/pkg/response/tmp.go",
		"github.com/pkg/response",
		"",
		make(map[string]string, 0),
		1,
	)
	if len(res) != 1 {
		t.Errorf("Expected %d lines, got %d", 1, len(res))
	}
	if res[0] != "github.com/pkg/response Name {string} true Description" {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Name {string} true Description", res[0])
	}

	// Cache, no changes, i.e. depth over 0
	r.types = map[string][]string{
		"github.com/pkg/response": {
			"Name {string} true Description",
		},
	}
	res = r.TypeToParams(
		"github.com/pkg/response/tmp.go",
		"github.com/pkg/response",
		"",
		make(map[string]string, 0),
		0,
	)
	if len(res) != 1 {
		t.Errorf("Expected %d lines, got %d", 1, len(res))
	}
	if res[0] != "github.com/pkg/response Name {string} true Description" {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Name {string} true Description", res[0])
	}

	// Valid, base types
	r.types = make(map[string][]string, 0)
	content :=
		`
		// User's name
		Name string
		Age int ` + "`json:\"age\"`" + `
		Height int64 ` + "`apitype:\"int\"`" + `
		Weight int ` + "`required:\"true\"`" + `
	`
	res = r.TypeToParams(
		"github.com/pkg/response/tmp.go",
		"github.com/pkg/response",
		content,
		make(map[string]string, 0),
		0,
	)
	if len(res) != 4 {
		t.Errorf("Expected %d lines, got %d", 4, len(res))
	}
	if res[0] != "github.com/pkg/response Name {string} false \"User's name\"" {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Name {string} false \"User's name\"", res[0])
	}
	if res[1] != "github.com/pkg/response age {int} false " {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response age {int} false ", res[1])
	}
	if res[2] != "github.com/pkg/response Height {int} false " {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Height {int} false ", res[2])
	}
	if res[3] != "github.com/pkg/response Weight {int} true " {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Weight {int} true ", res[3])
	}

	r.types = make(map[string][]string, 0)
	content =
		`
		Name string
	`
	res = r.TypeToParams(
		"github.com/pkg/response/tmp.go",
		"github.com/pkg/response",
		content,
		make(map[string]string, 0),
		1,
	)
	if len(res) != 1 {
		t.Errorf("Expected %d lines, got %d", 4, len(res))
	}
	if res[0] != "Name {string} false " {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "Name {string} false ", res[0])
	}

	// Valid, reference
	r.types = map[string][]string{
		"github.com/pkg/person.Name": {
			"firstname {string} false ",
			"lastname {string} false ",
		},
		"github.com/pkg/person.Details": {
			"age {int} false ",
			"gender {string} false ",
		},
	}
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				imports: map[string]string{
					"person": "github.com/pkg/person",
				},
			},
		},
		"github.com/pkg/person": {
			"github.com/pkg/person/person.go": {
				types: map[string]typeRef{
					"Name":    {},
					"Details": {},
				},
			},
		},
	}

	content =
		`
		Name person.Name
		Details *person.Details
	`
	res = r.TypeToParams(
		"github.com/pkg/response/tmp.go",
		"github.com/pkg/response",
		content,
		make(map[string]string, 0),
		0,
	)
	if len(res) != 6 {
		t.Errorf("Expected %d lines, got %d", 6, len(res))
	}
	if res[0] != "github.com/pkg/response Name {object}" {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Name {object}", res[0])
	}
	if res[1] != "github.com/pkg/response Name.firstname {string} false " {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Name.firstname {string} false ", res[1])
	}
	if res[2] != "github.com/pkg/response Name.lastname {string} false " {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Name.lastname {string} false ", res[2])
	}
	if res[3] != "github.com/pkg/response Details {object}" {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "github.com/pkg/response Details {object}", res[3])
	}

	// Valid, reference, non base depth
	r.types = map[string][]string{
		"github.com/pkg/person.Name": {
			"firstname {string} false ",
			"lastname {string} false ",
		},
		"github.com/pkg/person.Details": {
			"age {int} false ",
			"gender {string} false ",
		},
	}
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				imports: map[string]string{
					"person": "github.com/pkg/person",
				},
			},
		},
		"github.com/pkg/person": {
			"github.com/pkg/person/person.go": {
				types: map[string]typeRef{
					"Name":    {},
					"Details": {},
				},
			},
		},
	}

	res = r.TypeToParams(
		"github.com/pkg/response/tmp.go",
		"github.com/pkg/response",
		content,
		make(map[string]string, 0),
		1,
	)
	if len(res) != 6 {
		t.Errorf("Expected %d lines, got %d", 6, len(res))
	}
	if res[0] != "Name {object}" {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "Name {object}", res[0])
	}

	// Valid, reference, array
	r.types = map[string][]string{
		"github.com/pkg/person.Name": {
			"firstname {string} false ",
			"lastname {string} false ",
		},
		"github.com/pkg/person.Details": {
			"age {int} false ",
			"gender {string} false ",
		},
	}
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"github.com/pkg/response/tmp.go": {
				imports: map[string]string{
					"person": "github.com/pkg/person",
				},
			},
		},
		"github.com/pkg/person": {
			"github.com/pkg/person/person.go": {
				types: map[string]typeRef{
					"Name":    {},
					"Details": {},
				},
			},
		},
	}

	content =
		`
		Name []person.Name
		Details person.Details
	`
	res = r.TypeToParams(
		"github.com/pkg/response/tmp.go",
		"github.com/pkg/response",
		content,
		make(map[string]string, 0),
		1,
	)
	if len(res) != 6 {
		t.Errorf("Expected %d lines, got %d", 6, len(res))
	}
	if res[0] != "Name {[]object}" {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "Name {[]object}", res[0])
	}
}

func TestParseFieldMeta(t *testing.T) {
	r := NewResolver(false).(*resolver)
	// invalid
	res := r.ParseFieldMeta("")
	if len(res) != 0 {
		t.Errorf("Not expected parsing")
		return
	}

	// valid
	res = r.ParseFieldMeta("json:\"name\"")
	if len(res) != 1 {
		t.Errorf("Expected pairs %d, got %d", 1, len(res))
		return
	}
	if m, ok := res["json"]; ok {
		if m != "name" {
			t.Errorf("Expected \"%s\", got \"%s\"", "name", m)
		}
	} else {
		t.Errorf("Invalid parsing")
	}

	// many
	res = r.ParseFieldMeta("json:\"name\" required:\"true\"")
	if len(res) != 2 {
		t.Errorf("Expected pairs %d, got %d", 2, len(res))
		return
	}
}

func TestIsBasicType(t *testing.T) {
	valid := []string{"bool", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "byte", "rune", "float32", "float64", "complex64", "complex128", "object"}
	invalid := []string{"custom"}
	r := NewResolver(false).(*resolver)

	for _, c := range valid {
		if r.IsBasicType(c) == false {
			t.Errorf("%s validated as not basic type", c)
			return
		}
	}
	for _, c := range invalid {
		if r.IsBasicType(c) == true {
			t.Errorf("%s validated as basic type", c)
			return
		}
	}
}

func TestParsePackage(t *testing.T) {
	r := NewResolver(false).(*resolver)

	// Cached
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"tmp": {},
		},
	}
	err := r.ParsePackage("github.com/pkg/response", "github.com/pkg/response")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	// Invalid path
	err = r.ParsePackage("not-existing", "not-existing")
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	// Valid
	err = os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	err = os.MkdirAll("tmp/sub", os.ModePerm)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	files := []string{
		"tmp/tmp.go",
		"tmp/tmp.txt",
	}
	content := "package tmp"
	err = ioutil.WriteFile(files[0], []byte(content), 0644)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	err = ioutil.WriteFile(files[1], []byte(content), 0644)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	defer func() {
		os.Remove(files[0])
		os.Remove(files[1])
		os.Remove("tmp/sub")
		os.Remove("tmp")
	}()

	err = r.ParsePackage("tmp", "tmp")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
}

func TestParseFile(t *testing.T) {
	r := NewResolver(false).(*resolver)

	// Cached
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"tmp": {},
		},
	}
	err := r.ParseFile("github.com/pkg/response", "tmp")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	// Invalid file
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"other": {},
		},
	}
	err = r.ParseFile("github.com/pkg/response", "tmp")
	if err == nil {
		t.Errorf("Unexpected error, got nil")
		return
	}

	// Nothing to resolve
	content := "package tmp"
	file := "tmp"

	err = ioutil.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	defer func() {
		os.Remove(file)
	}()
	err = r.ParseFile("github.com/pkg/response", "tmp")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	// Single import
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"other": {},
		},
	}
	content =
		`
		package tmp
		import "github.com/pkg/request"

	`
	err = ioutil.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	err = r.ParseFile("github.com/pkg/response", "tmp")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	if i, ok := r.packages["github.com/pkg/response"]["tmp"].imports["request"]; ok {
		if i != "github.com/pkg/request" {
			t.Errorf("Expected \"%s\", got \"%s\"", "github.com/pkg/request", i)
		}
	} else {
		t.Errorf("Single import not resolved")
	}

	// Multi import
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"other": {},
		},
	}
	content =
		`
		package tmp
		import (
			"github.com/pkg/request"
			shortcut "github.com/pkg/tools"
		)

	`
	err = ioutil.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	err = r.ParseFile("github.com/pkg/response", "tmp")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	if i, ok := r.packages["github.com/pkg/response"]["tmp"].imports["request"]; ok {
		if i != "github.com/pkg/request" {
			t.Errorf("Expected \"%s\", got \"%s\"", "github.com/pkg/request", i)
		}
	} else {
		t.Errorf("Multi import not resolved")
	}
	if i, ok := r.packages["github.com/pkg/response"]["tmp"].imports["shortcut"]; ok {
		if i != "github.com/pkg/tools" {
			t.Errorf("Expected \"%s\", got \"%s\"", "github.com/pkg/tools", i)
		}
	} else {
		t.Errorf("Multi import not resolved")
	}

	// Types import
	r.packages = map[string]map[string]resolvedFile{
		"github.com/pkg/response": {
			"other": {},
		},
	}
	content =
		`
		package tmp
		
		type Person struct {
			Name string
			Age int
		}

		type Car struct {
			Brand string
			Details CarDetails
		}

	`
	err = ioutil.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	err = r.ParseFile("github.com/pkg/response", "tmp")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	if len(r.packages["github.com/pkg/response"]["tmp"].types) != 2 {
		t.Errorf("Expected %d captured types, got %d", 2, len(r.packages["github.com/pkg/response"]["tmp"].types))
	}
	if tr, ok := r.packages["github.com/pkg/response"]["tmp"].types["Person"]; ok {
		if trimAllSpaces(tr.content) != "NamestringAgeint" {
			t.Errorf("Expected \"%s\", got \"%s\"", "NamestringAgeint", trimAllSpaces(tr.content))
		}
	} else {
		t.Errorf("Type capture failed")
	}
}

func trimAllSpaces(s string) string {
	r := regexp.MustCompile("\\s")
	return r.ReplaceAllString(s, "")
}

package openapi

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spaceavocado/apidoc/token"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestGenerate(t *testing.T) {
	g := NewGenerator(false).(*generator)

	file := "tmp"
	defer func() {
		os.Remove(file)
	}()

	var err error
	var res string
	var expected string

	var read = func() {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		res = string(b)
		g.buffer.Clear()
	}

	// Invalid output folder
	err = g.Generate(
		[]token.Token{token.Token{}},
		[][]token.Token{[]token.Token{token.Token{}}},
		"invalid|folder/file.go")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Invalid file
	err = g.Generate(
		[]token.Token{token.Token{}},
		[][]token.Token{[]token.Token{token.Token{}}},
		"|.go")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Transformations
	g.buffer.Clear()
	g.Generate(
		[]token.Token{
			token.Token{
				Key:  "ver",
				Meta: map[string]string{"value": "1.0"},
			},
		},
		[][]token.Token{
			[]token.Token{
				token.Token{
					Key:  "param",
					Meta: map[string]string{"type": "{object}"},
				},
			},
		},
		file)

	expected = ""
	expected += "info:\n"
	expected += "  version: \"1.0\"\n"
	expected += "paths:\n"
	expected += "openapi: \"3.0.2\""

	read()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}

	// Components
	g.compCache = map[string][]string{
		"github.com/pkg.Peter": []string{
			"Peter:\n",
			"  type: object\n",
			"  properties:\n",
			"    firstname:\n",
			"      type: string\n",
		},
	}
	g.Generate(
		[]token.Token{
			token.Token{
				Key:  "ver",
				Meta: map[string]string{"value": "1.0"},
			},
		},
		[][]token.Token{
			[]token.Token{
				token.Token{},
			},
		},
		file)

	expected = ""
	expected += "info:\n"
	expected += "  version: \"1.0\"\n"
	expected += "components:\n"
	expected += "  schemas:\n"
	expected += "    Peter:\n"
	expected += "      type: object\n"
	expected += "      properties:\n"
	expected += "        firstname:\n"
	expected += "          type: string\n"
	expected += "paths:\n"
	expected += "openapi: \"3.0.2\""

	read()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}

	// Router
	g.compCache = make(map[string][]string, 0)
	g.Generate(
		[]token.Token{
			token.Token{},
		},
		[][]token.Token{
			[]token.Token{
				token.Token{
					Key: "summary",
					Meta: map[string]string{
						"value": "sample",
					},
				},
				token.Token{
					Key: "id",
					Meta: map[string]string{
						"value": "action-id",
					},
				},
				token.Token{
					Key: "desc",
					Meta: map[string]string{
						"value": "lorem",
					},
				},
				token.Token{
					Key: "tag",
					Meta: map[string]string{
						"value": "Dolor",
					},
				},
				token.Token{
					Key: "router",
					Meta: map[string]string{
						"url":    "/sample",
						"method": "[get]",
					},
				},
				token.Token{
					Key: "produce",
					Meta: map[string]string{
						"value": "json",
					},
				},
				token.Token{
					Key: "body",
					Meta: map[string]string{
						"value": "github.com/pkg.Peter",
					},
				},
				token.Token{
					Key: "bref",
					Meta: map[string]string{
						"pkg.type": "github.com/pkg.Peter",
						"key":      "firstname",
						"type":     "{string}",
						"req":      "false",
						"desc":     "Name",
					},
				},
				token.Token{
					Key: "accept",
					Meta: map[string]string{
						"value": "json",
					},
				},
				token.Token{
					Key: "success",
					Meta: map[string]string{
						"code": "200",
						"type": "{string}",
						"desc": "OK",
					},
				},
				token.Token{
					Key: "param",
					Meta: map[string]string{
						"key":  "token",
						"type": "{string}",
						"req":  "true",
						"desc": "Token",
					},
				},
			},
		},
		file)

	expected = ""
	expected += "info:\n"
	expected += "components:\n"
	expected += "  schemas:\n"
	expected += "    Peter:\n"
	expected += "      type: object\n"
	expected += "      properties:\n"
	expected += "        firstname:\n"
	expected += "          description: Name\n"
	expected += "          type: string\n"
	expected += "paths:\n"
	expected += "  /sample:\n"
	expected += "    get:\n"
	expected += "      summary: sample\n"
	expected += "      operationId: action-id\n"
	expected += "      description: lorem\n"
	expected += "      tags:\n"
	expected += "      - Dolor\n"
	expected += "      parameters:\n"
	expected += "      - name: token\n"
	expected += "        description: Token\n"
	expected += "        required: true\n"
	expected += "        schema:\n"
	expected += "          type: string\n"
	expected += "      requestBody:\n"
	expected += "        content:\n"
	expected += "          application/json:\n"
	expected += "            schema:\n"
	expected += "              $ref: \"#/components/schemas/Peter\"\n"
	expected += "      responses:\n"
	expected += "        \"200\":\n"
	expected += "          description: OK\n"
	expected += "          content:\n"
	expected += "            text/plain:\n"
	expected += "              schema:\n"
	expected += "                type: string\n"
	expected += "openapi: \"3.0.2\""

	read()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}
}

func TestMainSection(t *testing.T) {
	g := NewGenerator(false).(*generator)
	g.MainSection([]token.Token{
		token.Token{
			Key:  "title",
			Meta: map[string]string{"value": "title"},
		},
		token.Token{
			Key:  "ver",
			Meta: map[string]string{"value": "1.0"},
		},
		token.Token{
			Key:  "desc",
			Meta: map[string]string{"value": "lorem"},
		},
		token.Token{
			Key:  "contact.name",
			Meta: map[string]string{"value": "name"},
		},
		token.Token{
			Key:  "lic.url",
			Meta: map[string]string{"value": "url"},
		},
		token.Token{
			Key:  "server",
			Meta: map[string]string{"url": "url1", "desc": "lorem 1"},
		},
		token.Token{
			Key:  "server",
			Meta: map[string]string{"url": "url2", "desc": "lorem 2"},
		},
	})

	expected := ""
	expected += "info:\n  title: title\n  version: 1.0\n  description: lorem\n"
	expected += "  contact:\n    name: name\n"
	expected += "  license:\n    url: url\n"
	expected += "servers:\n"
	expected += "  - url: url1\n    description: lorem 1\n"
	expected += "  - url: url2\n    description: lorem 2\n"

	res := g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}
}

func TestParamsSection(t *testing.T) {
	g := NewGenerator(false).(*generator)

	var res string

	// Non array
	g.ParamsSection([]token.Token{
		token.Token{
			Key: "param",
			Meta: map[string]string{
				(g.nameMetaKey): "param1",
				"desc":          "desc1",
				"in":            "path",
				(g.reqMetaKey):  "true",
				(g.typeMetaKey): "int",
			},
		},
	}, 0)

	expected := ""
	expected += "parameters:\n"
	expected += "- name: param1\n"
	expected += "  description: desc1\n"
	expected += "  in: path\n"
	expected += "  required: true\n"
	expected += "  schema:\n"
	expected += "    type: int\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}

	// Array
	g.buffer.Clear()
	g.ParamsSection([]token.Token{
		token.Token{
			Key: "param",
			Meta: map[string]string{
				(g.nameMetaKey): "param1",
				"desc":          "desc1",
				"in":            "path",
				(g.reqMetaKey):  "true",
				(g.typeMetaKey): "array int",
			},
		},
	}, 0)

	expected = ""
	expected += "parameters:\n"
	expected += "- name: param1\n"
	expected += "  description: desc1\n"
	expected += "  in: path\n"
	expected += "  required: true\n"
	expected += "  schema:\n"
	expected += "    type: array\n"
	expected += "    items:\n"
	expected += "      type: int\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}
}

func TestBodySection(t *testing.T) {
	g := NewGenerator(false).(*generator)
	g.compMapping = map[string]string{
		"github.com/pkg.Peter": "Peter",
	}

	var res string
	var mts = []string{
		"application/json",
		"multipart/form-data",
	}

	// Non array
	g.BodySection(token.Token{
		Key: "body",
		Meta: map[string]string{
			"value": "github.com/pkg.Peter",
		},
	}, mts, 0)

	expected := ""
	expected += "requestBody:\n"
	expected += "  content:\n"
	expected += "    application/json:\n"
	expected += "      schema:\n"
	expected += "        $ref: \"#/components/schemas/Peter\"\n"
	expected += "    multipart/form-data:\n"
	expected += "      schema:\n"
	expected += "        $ref: \"#/components/schemas/Peter\"\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}

	// Array
	g.buffer.Clear()
	g.BodySection(token.Token{
		Key: "param",
		Meta: map[string]string{
			"value": "[]github.com/pkg.Peter",
		},
	}, mts, 0)

	expected = ""
	expected += "requestBody:\n"
	expected += "  content:\n"
	expected += "    application/json:\n"
	expected += "      schema:\n"
	expected += "        type: array\n"
	expected += "        items:\n"
	expected += "          $ref: \"#/components/schemas/Peter\"\n"
	expected += "    multipart/form-data:\n"
	expected += "      schema:\n"
	expected += "        type: array\n"
	expected += "        items:\n"
	expected += "          $ref: \"#/components/schemas/Peter\"\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}
}

func TestResponseSection(t *testing.T) {
	g := NewGenerator(false).(*generator)

	// Just run the method, to see if any error occurs
	g.ResponseSection(0, []token.Token{
		token.Token{
			Key: "produce",
			Meta: map[string]string{
				"value": "application/json",
			},
		},
	}, 0)
}

func TestResponse(t *testing.T) {
	g := NewGenerator(false).(*generator)
	g.compMapping = map[string]string{
		"github.com/pkg.Peter": "Peter",
	}
	g.wrappers = map[string][]dataWrapper{
		"success": []dataWrapper{
			dataWrapper{
				lines: make([]string, 0),
			},
		},
	}

	var res string
	var expected string
	var mts = []string{
		"application/json",
		"multipart/form-data",
	}

	// Success, plain
	g.Response("success", mts, 0, []token.Token{
		token.Token{
			Key: "success",
			Meta: map[string]string{
				"code": "\"200\"",
				"type": "string",
				"desc": "lorem",
			},
		},
	}, 0)
	expected = ""
	expected += "  \"200\":\n"
	expected += "    description: lorem\n"
	expected += "    content:\n"
	expected += "      text/plain:\n"
	expected += "        schema:\n"
	expected += "          type: string\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}

	// Success, direct reference
	g.buffer.Clear()
	g.Response("success", mts, 0, []token.Token{
		token.Token{
			Key: "success",
			Meta: map[string]string{
				"code": "\"200\"",
				"type": "object",
				"ref":  "github.com/pkg.Peter",
				"desc": "lorem",
			},
		},
	}, 0)
	expected = ""
	expected += "  \"200\":\n"
	expected += "    description: lorem\n"
	expected += "    content:\n"
	expected += "      application/json:\n"
	expected += "        schema:\n"
	expected += "          $ref: \"#/components/schemas/Peter\"\n"
	expected += "      multipart/form-data:\n"
	expected += "        schema:\n"
	expected += "          $ref: \"#/components/schemas/Peter\"\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}

	// Success, direct array
	g.buffer.Clear()
	g.Response("success", mts, 0, []token.Token{
		token.Token{
			Key: "success",
			Meta: map[string]string{
				"code": "\"200\"",
				"type": "object",
				"ref":  "[]github.com/pkg.Peter",
				"desc": "lorem",
			},
		},
	}, 0)
	expected = ""
	expected += "  \"200\":\n"
	expected += "    description: lorem\n"
	expected += "    content:\n"
	expected += "      application/json:\n"
	expected += "        schema:\n"
	expected += "          type: array\n"
	expected += "          items:\n"
	expected += "            $ref: \"#/components/schemas/Peter\"\n"
	expected += "      multipart/form-data:\n"
	expected += "        schema:\n"
	expected += "          type: array\n"
	expected += "          items:\n"
	expected += "            $ref: \"#/components/schemas/Peter\"\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}

	// Success, Wrapper
	g.compMapping = map[string]string{
		"github.com/pkg.Peter": "Peter",
	}
	g.wrappers = map[string][]dataWrapper{
		"success": []dataWrapper{
			dataWrapper{
				pos: 4,
				lines: []string{
					":\n",
					"  type: object\n",
					"  properties:\n",
					"    data:\n",
					"->4\n",
					"    status:\n",
					"      type: string\n",
				},
			},
		},
	}
	g.buffer.Clear()
	g.Response("success", mts, 0, []token.Token{
		token.Token{
			Key: "success",
			Meta: map[string]string{
				"code": "\"200\"",
				"type": "object",
				"ref":  "github.com/pkg.Peter",
				"desc": "lorem",
			},
		},
	}, 0)
	expected = ""
	expected += "  \"200\":\n"
	expected += "    description: lorem\n"
	expected += "    content:\n"
	expected += "      application/json:\n"
	expected += "        schema:\n"
	expected += "          type: object\n"
	expected += "          properties:\n"
	expected += "            data:\n"
	expected += "              $ref: \"#/components/schemas/Peter\"\n"
	expected += "            status:\n"
	expected += "              type: string\n"
	expected += "      multipart/form-data:\n"
	expected += "        schema:\n"
	expected += "          type: object\n"
	expected += "          properties:\n"
	expected += "            data:\n"
	expected += "              $ref: \"#/components/schemas/Peter\"\n"
	expected += "            status:\n"
	expected += "              type: string\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}

	// Success, Wrapper, array
	g.buffer.Clear()
	g.Response("success", mts, 0, []token.Token{
		token.Token{
			Key: "success",
			Meta: map[string]string{
				"code": "\"200\"",
				"type": "object",
				"ref":  "[]github.com/pkg.Peter",
				"desc": "lorem",
			},
		},
	}, 0)
	expected = ""
	expected += "  \"200\":\n"
	expected += "    description: lorem\n"
	expected += "    content:\n"
	expected += "      application/json:\n"
	expected += "        schema:\n"
	expected += "          type: object\n"
	expected += "          properties:\n"
	expected += "            data:\n"
	expected += "              type: array\n"
	expected += "              items:\n"
	expected += "                $ref: \"#/components/schemas/Peter\"\n"
	expected += "            status:\n"
	expected += "              type: string\n"
	expected += "      multipart/form-data:\n"
	expected += "        schema:\n"
	expected += "          type: object\n"
	expected += "          properties:\n"
	expected += "            data:\n"
	expected += "              type: array\n"
	expected += "              items:\n"
	expected += "                $ref: \"#/components/schemas/Peter\"\n"
	expected += "            status:\n"
	expected += "              type: string\n"

	res = g.buffer.Flush()
	if res != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, res)
	}
}

func TestBufferTokenMeta(t *testing.T) {
	g := NewGenerator(false).(*generator)

	// Valid meta
	g.BufferTokenMeta(
		token.Token{
			Key: "a",
			Meta: map[string]string{
				"meta_a": "value_a",
				"meta_b": "value_b",
				"meta_c": "value_c",
			},
		}, "meta_b", "key", 0)
	res := g.buffer.Flush()
	if res != "key: value_b\n" {
		t.Errorf("Expected \"%s\", got \"%s\"", "key: value_b\n", res)
	}

	// Invalid meta
	g.BufferTokenMeta(
		token.Token{
			Key: "a",
			Meta: map[string]string{
				"meta_a": "value_a",
			},
		}, "meta_b", "key", 0)
	res = g.buffer.Flush()
	if res != "key: value_b\n" {
		t.Errorf("Expected \"%s\", got \"%s\"", "key: value_b\n", res)
	}
}

func TestResolveWrappers(t *testing.T) {
	g := NewGenerator(false).(*generator)

	var expected []string

	// Success
	g.wrappers["success"] = make([]dataWrapper, 0)
	g.wrappers["failure"] = make([]dataWrapper, 0)
	g.ResolveWrappers([]token.Token{
		token.Token{
			Key: g.wrapperSuccessTokenKey,
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				(g.prtMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: g.wrapperSuccessTokenKey,
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "data",
				(g.typeMetaKey): "object",
				(g.reqMetaKey):  "false",
				(g.prtMetaKey):  "true",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: g.wrapperSuccessTokenKey,
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				(g.prtMetaKey):  "false",
				"desc":          "lorem",
			},
		},
	})
	expected = []string{
		":\n",
		"  type: object\n",
		"  properties:\n",
		"    firstname:\n",
		"      description: lorem\n",
		"      type: string\n",
		"    data:\n",
		"-> 0\n",
		"    lastname:\n",
		"      description: lorem\n",
		"      type: string\n",
	}
	if len(g.wrappers["success"][0].lines) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(g.wrappers["success"][0].lines))
		return
	}
	for i, l := range g.wrappers["success"][0].lines {
		if l != expected[i] {
			t.Errorf("Expected \"%s\", got \"%s\"", expected[i], l)
		}
	}

	// Failure
	g.wrappers["success"] = make([]dataWrapper, 0)
	g.wrappers["failure"] = make([]dataWrapper, 0)
	g.ResolveWrappers([]token.Token{
		token.Token{
			Key: g.wrapperErrorTokenKey,
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				(g.prtMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: g.wrapperErrorTokenKey,
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "data",
				(g.typeMetaKey): "object",
				(g.reqMetaKey):  "false",
				(g.prtMetaKey):  "true",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: g.wrapperErrorTokenKey,
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				(g.prtMetaKey):  "false",
				"desc":          "lorem",
			},
		},
	})
	expected = []string{
		":\n",
		"  type: object\n",
		"  properties:\n",
		"    firstname:\n",
		"      description: lorem\n",
		"      type: string\n",
		"    data:\n",
		"-> 0\n",
		"    lastname:\n",
		"      description: lorem\n",
		"      type: string\n",
	}
	if len(g.wrappers["failure"][0].lines) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(g.wrappers["failure"][0].lines))
		return
	}
	for i, l := range g.wrappers["failure"][0].lines {
		if l != expected[i] {
			t.Errorf("Expected \"%s\", got \"%s\"", expected[i], l)
		}
	}
}

func TestResolveComponents(t *testing.T) {
	g := NewGenerator(false).(*generator)

	var expected map[string][]string

	// Already in cache
	g.compCache = map[string][]string{"github.com/pkg.Peter": []string{""}}
	expected = map[string][]string{"github.com/pkg.Peter": []string{""}}
	g.ResolveComponents([]token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "ipsum",
			},
		},
	})
	if len(g.compCache) != len(expected) {
		t.Errorf("Invalid cache modification")
		return
	}
	for k, c := range g.compCache {
		el, ok := expected[k]
		if ok == false {
			t.Errorf("Invalid cache modification")
			return
		}
		for i, l := range el {
			if l != c[i] {
				t.Errorf("Expected \"%s\", got \"%s\"", c[i], l)
			}
		}
	}

	// Unique name
	g.compCache = map[string][]string{"": []string{""}}
	g.compUniqueName = []string{
		"Peter",
	}
	g.ResolveComponents([]token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "ipsum",
			},
		},
	})
	if len(g.compUniqueName) != 2 {
		t.Errorf("Expected %d, got %d", 2, len(g.compUniqueName))
		return
	}
	if g.compUniqueName[1] != "Peter1" {
		t.Errorf("Expected \"%s\", got \"%s\"", "Peter1", g.compUniqueName[1])
	}

	// Cache properly build
	g.compCache = make(map[string][]string, 0)
	expected = map[string][]string{
		"github.com/pkg.Peter": []string{
			"Peter:\n",
			"  type: object\n",
		},
	}
	g.ResolveComponents([]token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "ipsum",
			},
		},
	})
	if len(g.compCache) != len(expected) {
		t.Errorf("Invalid cache modification")
		return
	}
	for k, c := range g.compCache {
		el, ok := expected[k]
		if ok == false {
			t.Errorf("Invalid cache modification")
			return
		}
		for i, l := range el {
			if l != c[i] {
				t.Errorf("Expected \"%s\", got \"%s\"", l, c[i])
			}
		}
	}

	// Reducing
	g.compCache = make(map[string][]string, 0)
	expected = map[string][]string{
		"github.com/pkg.Peter": []string{
			"Peter:\n",
			"  type: object\n",
		},
	}
	g.ResolveComponents([]token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "sref",
			Meta: map[string]string{
				"pkg.type":      "github.com/pkg.Peter",
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "ipsum",
			},
		},
	})
	if len(g.compCache) != len(expected) {
		t.Errorf("Invalid cache modification")
		return
	}
	if len(g.compCache["github.com/pkg.Peter"]) != 9 {
		t.Errorf("Expected %d, got %d", 9, len(g.compCache["github.com/pkg.Peter"]))
	}
}

func TestParseObject(t *testing.T) {
	g := NewGenerator(false).(*generator)

	var res []string
	var expected []string

	// Flat, no array
	res = g.ParseObject("github.com/pkg.Peter", []token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "true",
				"desc":          "impsum",
			},
		},
	}, 0, false)

	expected = []string{
		"Peter:\n",
		"  type: object\n",
		"  required:\n",
		"  - lastname\n",
		"  properties:\n",
		"    firstname:\n",
		"      description: lorem\n",
		"      type: string\n",
		"    lastname:\n",
		"      description: impsum\n",
		"      type: string\n",
	}
	if len(res) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(res))
		return
	}
	for i, l := range res {
		if l != expected[i] {
			t.Errorf("Expected \"%s\", got \"%s\"", expected[i], l)
		}
	}

	// Flat, array
	res = g.ParseObject("github.com/pkg.Peter", []token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "true",
				"desc":          "impsum",
			},
		},
	}, 0, true)

	expected = []string{
		"Peter:\n",
		"  type: array\n",
		"  items:\n",
		"    required:\n",
		"    - lastname\n",
		"    properties:\n",
		"      firstname:\n",
		"        description: lorem\n",
		"        type: string\n",
		"      lastname:\n",
		"        description: impsum\n",
		"        type: string\n",
	}

	if len(res) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(res))
		return
	}
	for i, l := range res {
		if l != expected[i] {
			t.Errorf("Expected \"%s\", got \"%s\"", expected[i], l)
		}
	}

	// With references, no array
	res = g.ParseObject("github.com/pkg.Peter", []token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "details",
				(g.typeMetaKey): "object",
				(g.reqMetaKey):  "false",
				"desc":          "impsum",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "details.age",
				(g.typeMetaKey): "int",
				(g.reqMetaKey):  "true",
				"desc":          "dolor",
			},
		},
	}, 0, false)

	expected = []string{
		"Peter:\n",
		"  type: object\n",
		"  properties:\n",
		"    firstname:\n",
		"      description: lorem\n",
		"      type: string\n",
		"    details:\n",
		"      type: object\n",
		"      required:\n",
		"      - age\n",
		"      properties:\n",
		"        age:\n",
		"          description: dolor\n",
		"          type: int\n",
	}
	if len(res) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(res))
		return
	}
	for i, l := range res {
		if l != expected[i] {
			t.Errorf("Expected \"%s\", got \"%s\"", expected[i], l)
		}
	}

	// With references, inner array
	res = g.ParseObject("github.com/pkg.Peter", []token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "details",
				(g.typeMetaKey): "object",
				(g.reqMetaKey):  "false",
				"desc":          "impsum",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "details.age",
				(g.typeMetaKey): "array int",
				(g.reqMetaKey):  "true",
				"desc":          "dolor",
			},
		},
	}, 0, false)

	expected = []string{
		"Peter:\n",
		"  type: object\n",
		"  properties:\n",
		"    firstname:\n",
		"      description: lorem\n",
		"      type: string\n",
		"    details:\n",
		"      type: object\n",
		"      required:\n",
		"      - age\n",
		"      properties:\n",
		"        age:\n",
		"          description: dolor\n",
		"          type: array\n",
		"          items:\n",
		"            type: int\n",
	}
	if len(res) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(res))
		return
	}
	for i, l := range res {
		if l != expected[i] {
			t.Errorf("Expected \"%s\", got \"%s\"", expected[i], l)
		}
	}

	// Wrapper
	res = g.ParseObject("github.com/pkg.Peter", []token.Token{
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "firstname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "lorem",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "data",
				(g.typeMetaKey): "object",
				(g.reqMetaKey):  "false",
				"desc":          "impsum",
				(g.prtMetaKey):  "true",
			},
		},
		token.Token{
			Key: "bref",
			Meta: map[string]string{
				(g.nameMetaKey): "lastname",
				(g.typeMetaKey): "string",
				(g.reqMetaKey):  "false",
				"desc":          "impsum",
			},
		},
	}, 0, false)

	expected = []string{
		"Peter:\n",
		"  type: object\n",
		"  properties:\n",
		"    firstname:\n",
		"      description: lorem\n",
		"      type: string\n",
		"    data:\n",
		"-> 0\n",
		"    lastname:\n",
		"      description: impsum\n",
		"      type: string\n",
	}
	if len(res) != len(expected) {
		t.Errorf("Expected %d entries, got %d", len(expected), len(res))
		return
	}
	for i, l := range res {
		if l != expected[i] {
			t.Errorf("Expected \"%s\", got \"%s\"", expected[i], l)
		}
	}
}

func TestTokenMeta(t *testing.T) {
	g := NewGenerator(false).(*generator)

	// Found
	res, ok := g.TokenMeta(
		token.Token{
			Key: "a",
			Meta: map[string]string{
				"meta_a": "value_a",
				"meta_b": "value_b",
				"meta_c": "value_c",
			},
		}, "meta_b")
	if ok == false {
		t.Errorf("Expected valid meta, got nil")
		return
	}
	if res != "value_b" {
		t.Errorf("Expected \"%s\", got \"%s\"", "value_b", res)
	}

	// Missing
	res, ok = g.TokenMeta(
		token.Token{
			Key: "a",
			Meta: map[string]string{
				"meta_a": "value_a",
			},
		}, "meta_b")
	if ok {
		t.Errorf("Expected nil, got \"%s\"", res)
		return
	}
}

func TestGetToken(t *testing.T) {
	g := NewGenerator(false).(*generator)

	// Found
	res, ok := g.GetToken([]token.Token{
		token.Token{
			Key: "a",
		},
		token.Token{
			Key: "b",
		},
		token.Token{
			Key: "c",
		},
	}, "b")
	if ok == false {
		t.Errorf("Expected valid token, got nil")
		return
	}
	if res.Key != "b" {
		t.Errorf("Expected \"%s\", got \"%s\"", "b", res.Key)
	}

	// Missing
	res, ok = g.GetToken([]token.Token{
		token.Token{
			Key: "a",
		},
	}, "b")
	if ok {
		t.Errorf("Expected nil, got \"%s\"", res.Key)
		return
	}
}

func TestGetTokens(t *testing.T) {
	g := NewGenerator(false).(*generator)

	var res []token.Token

	// Single key
	res = g.GetTokens([]token.Token{
		token.Token{
			Key: "a",
		},
		token.Token{
			Key: "b",
		},
		token.Token{
			Key: "c",
		},
	}, "b")
	if len(res) != 1 {
		t.Errorf("Expected %d entries, got %d", 1, len(res))
	}
	if res[0].Key != "b" {
		t.Errorf("Expected \"%s\", got \"%s\"", "b", res[0].Key)
	}

	// Many keys
	res = g.GetTokens([]token.Token{
		token.Token{
			Key: "a",
		},
		token.Token{
			Key: "b",
		},
		token.Token{
			Key: "c",
		},
	}, "a", "c")
	if len(res) != 2 {
		t.Errorf("Expected %d entries, got %d", 2, len(res))
	}
	if res[0].Key != "a" {
		t.Errorf("Expected \"%s\", got \"%s\"", "a", res[0].Key)
	}
	if res[1].Key != "c" {
		t.Errorf("Expected \"%s\", got \"%s\"", "c", res[1].Key)
	}
}

func TestTokensByPrefix(t *testing.T) {
	g := NewGenerator(false).(*generator)
	res := g.GetTokensByPrefix([]token.Token{
		token.Token{
			Key: "a",
		},
		token.Token{
			Key: "desired_b",
		},
		token.Token{
			Key: "c",
		},
	}, "desired_")
	if len(res) != 1 {
		t.Errorf("Expected %d entries, got %d", 1, len(res))
	}
	if res[0].Key != "desired_b" {
		t.Errorf("Expected \"%s\", got \"%s\"", "desired_b", res[0].Key)
	}
}

func TestGetRequiredTokens(t *testing.T) {
	g := NewGenerator(false).(*generator)
	res := g.GetRequiredTokens([]token.Token{
		token.Token{
			Key: "a",
			Meta: map[string]string{
				(g.reqMetaKey): "true",
			},
		},
		token.Token{
			Key: "b",
			Meta: map[string]string{
				(g.reqMetaKey): "false",
			},
		},
		token.Token{
			Key: "c",
			Meta: map[string]string{
				"other": "value",
			},
		},
	})
	if len(res) != 1 {
		t.Errorf("Expected %d entries, got %d", 1, len(res))
	}
	if res[0].Key != "a" {
		t.Errorf("Expected \"%s\", got \"%s\"", "a", res[0].Key)
	}
}

func TestParseArray(t *testing.T) {
	g := NewGenerator(false).(*generator)

	var res []string

	// Parse
	res = g.ParseArray(" peter, william ", trsEmpty)
	if len(res) != 2 {
		t.Errorf("Expected %d entries, got %d", 2, len(res))
	}
	if res[0] != "peter" {
		t.Errorf("Expected \"%s\", got \"%s\"", "peter", res)
	}
	if res[1] != "william" {
		t.Errorf("Expected \"%s\", got \"%s\"", "william", res)
	}

	// Parse with transform
	res = g.ParseArray(" peter, william ", func(i string) string {
		return i + "_postfix"
	})
	if len(res) != 2 {
		t.Errorf("Expected %d entries, got %d", 2, len(res))
	}
	if res[0] != "peter_postfix" {
		t.Errorf("Expected \"%s\", got \"%s\"", "peter_postfix", res)
	}
	if res[1] != "william_postfix" {
		t.Errorf("Expected \"%s\", got \"%s\"", "william_postfix", res)
	}
}

func TestComponentRef(t *testing.T) {
	g := NewGenerator(false).(*generator)
	g.compMapping = map[string]string{
		"github.com/pkg.Peter": "Peter",
	}

	var res string

	// Found
	res = g.ComponentRef("github.com/pkg.Peter")
	if strings.Contains(res, "#/components/schemas/Peter") == false {
		t.Errorf("Expected \"%s\", got \"%s\"", "#/components/schemas/Peter", res)
	}

	// Missing
	res = g.ComponentRef("github.com/pkg.Missing")
	if res != "" {
		t.Errorf("Expected \"%s\", got \"%s\"", "", res)
	}

	// Verbose
	b := &bytes.Buffer{}
	log.SetOutput(b)
	hook := test.NewGlobal()
	g = NewGenerator(true).(*generator)
	g.compMapping = map[string]string{
		"github.com/pkg.Peter": "Peter",
	}

	g.ComponentRef("github.com/pkg.Missing")
	if len(hook.Entries) != 1 {
		t.Errorf("Has %d logs, expected %d logs", len(hook.Entries), 1)
		return
	}
	o, err := hook.Entries[0].String()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if strings.Contains(o, "generator: missing component reference") == false {
		t.Errorf("Has %s, expected %s", o, "generator: missing component reference")
	}
}

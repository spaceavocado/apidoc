// Package openapi generates the documentation into OpenAPI format.
// Specification: https://swagger.io/specification/
package openapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spaceavocado/apidoc/misc"
	"github.com/spaceavocado/apidoc/output"
	"github.com/spaceavocado/apidoc/token"
)

type generator struct {
	verbose bool
	// OpenAPI version
	version string

	// Output buffer
	buffer buffer

	// Components unique names
	// References extracted from the documentation might have
	// the same names, therefore this prevents duplications
	compUniqueName []string
	// Name mapping between the package.struct name
	// and clean simplified name, i.e. components have
	// nice struct names in the documentation
	compMapping map[string]string
	// Component buffer cache
	// comp.name -> buffer lines
	compCache map[string][]string
	// Component token keys
	compTokenKeys []string

	// Data wrappers, per type (success, failure)
	// for each endpoint by endpoint index
	wrappers               map[string][]dataWrapper
	wrapperSuccessTokenKey string
	wrapperErrorTokenKey   string

	// Meta key used to determine if the token,
	// param, object prop, is marked as required
	reqMetaKey string
	// Meta key used for property/filed name
	nameMetaKey string
	// Meta key used for data pointer identification
	// within the wrapper object
	prtMetaKey string
	// Meta key used for property/filed name data type
	typeMetaKey string

	// Token meta values transformation mapping
	// TokenKey -> MetaKey -> transformation
	trs map[string]map[string]transformation
}

// DataWrapper structure holds the generated
// data wrapper output annotation. Depth indicates
// where the data pointer is horizontally, pos
// indicates the vertical position.
// The data pointer is the object prop used to
// contain the data object, wrapped in this wrapper.
type dataWrapper struct {
	pos   int
	depth int
	lines []string
}

// Generate the documentation from the given tokens for the
// main section and for the given endpoints, into the file.
func (g *generator) Generate(main []token.Token, endpoints [][]token.Token, file string) error {
	// Transform meta
	for _, t := range main {
		if _, ok := g.trs[t.Key]; ok {
			for k, m := range t.Meta {
				if trs, ok := g.trs[t.Key][k]; ok {
					t.Meta[k] = trs(m)
				}
			}
		}
	}
	for _, e := range endpoints {
		for _, t := range e {
			if _, ok := g.trs[t.Key]; ok {
				for k, m := range t.Meta {
					if trs, ok := g.trs[t.Key][k]; ok {
						t.Meta[k] = trs(m)
					}
				}
			}
		}
	}

	// Resolve main secion
	g.MainSection(main)

	// Wrappers and references
	for _, e := range endpoints {
		g.ResolveWrappers(e)
		g.ResolveComponents(e)
	}

	// Components
	if len(g.compCache) > 0 {
		g.buffer.Label("components", 0)
		g.buffer.Label("schemas", 1)
		for _, c := range g.compCache {
			for _, l := range c {
				g.buffer.Write(l, 2)
			}
		}
	}

	// Resolved endpoints
	// Used to prevent re-generation of sibling endpoints
	resolved := make([]string, 0, len(endpoints))

	// Resolve endpoints
	g.buffer.Label("paths", 0)
	for index, e := range endpoints {
		if t, ok := g.GetToken(e, "router"); ok {
			if m, ok := g.TokenMeta(t, "url"); ok {
				if misc.StringInSlice(m, resolved) {
					continue
				}
				resolved = append(resolved, m)
				g.buffer.Label(m, 1)

				// Get sibling endpoints
				group := [][]token.Token{e}
				if index < len(endpoints)-1 {
					if siblings := g.SiblingEndpoints(m, endpoints[index+1:]); len(siblings) > 0 {
						group = append(group, siblings...)
					}
				}

				// Resolved groupped endpoints
				for _, e := range group {
					// Tags
					tags := make([]string, 0)
					if t, ok := g.GetToken(e, "tag"); ok {
						if m, ok := g.TokenMeta(t, "value"); ok && len(m) > 0 {
							tags = g.ParseArray(m, trsEmpty)
						}
					}

					t, _ := g.GetToken(e, "router")

					// Methods
					if m, ok := g.TokenMeta(t, "method"); ok {
						methods := g.ParseArray(m, trsEmpty)
						for _, method := range methods {
							g.buffer.Label(method, 2)
							if t, ok := g.GetToken(e, "summary"); ok {
								g.BufferTokenMeta(t, "value", "summary", 3)
							}
							if t, ok := g.GetToken(e, "id"); ok {
								g.BufferTokenMeta(t, "value", "operationId", 3)
							}
							if t, ok := g.GetToken(e, "desc"); ok {
								g.BufferTokenMeta(t, "value", "description", 3)
							}

							// Tags
							if len(tags) > 0 {
								g.buffer.Label("tags", 3)
								for _, tag := range tags {
									g.buffer.Line(fmt.Sprintf("- %s", tag), 3)
								}
							}

							// Params
							if params := g.GetTokens(e, "param"); len(params) > 0 {
								g.ParamsSection(params, 3)
							}

							// Body
							if body, ok := g.GetToken(e, "body"); ok {
								if t, ok := g.GetToken(e, "accept"); ok {
									if m, ok := g.TokenMeta(t, "value"); ok && len(m) > 0 {
										mts := g.ParseArray(m, g.trs["accept"]["value"])
										g.BodySection(body, mts, 3)
									}
								}
							}

							// Response
							g.ResponseSection(index, e, 3)
						}
					}
				}
			}
		}
	}

	// OpenAPI version
	g.buffer.Write(fmt.Sprintf("openapi: \"%s\"", g.version), 0)

	// Output folder
	err := os.MkdirAll(filepath.Dir(file), os.ModePerm)
	if err != nil {
		return err
	}

	// Save the output file
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	fp.Write([]byte(g.buffer.Flush()))
	return fp.Close()
}

// MainSection processing
func (g *generator) MainSection(main []token.Token) {
	// Info section
	g.buffer.Label("info", 0)
	if t, ok := g.GetToken(main, "title"); ok {
		g.BufferTokenMeta(t, "value", "title", 1)
	}
	if t, ok := g.GetToken(main, "ver"); ok {
		g.BufferTokenMeta(t, "value", "version", 1)
	}
	if t, ok := g.GetToken(main, "desc"); ok {
		g.BufferTokenMeta(t, "value", "description", 1)
	}

	// Contact
	if col := g.GetTokensByPrefix(main, "contact."); len(col) > 0 {
		g.buffer.Label("contact", 1)
		for _, t := range col {
			if m, ok := t.Meta["value"]; ok {
				name := strings.TrimPrefix(t.Key, "contact.")
				g.buffer.KeyValue(name, trsSafeValue(m), 2)
			}
		}
	}

	// License
	if col := g.GetTokensByPrefix(main, "lic."); len(col) > 0 {
		g.buffer.Label("license", 1)
		for _, t := range col {
			if m, ok := t.Meta["value"]; ok {
				name := strings.TrimPrefix(t.Key, "lic.")
				g.buffer.KeyValue(name, trsSafeValue(m), 2)
			}
		}
	}

	// Servers
	if servers := g.GetTokens(main, "server"); len(servers) > 0 {
		g.buffer.Label("servers", 0)
		for _, t := range servers {
			g.BufferTokenMeta(t, "url", "- url", 1)
			g.BufferTokenMeta(t, "desc", "description", 2)
		}
	}
}

// ParamsSection processing
func (g *generator) ParamsSection(params []token.Token, depth int) {
	g.buffer.Label("parameters", depth)
	for _, t := range params {
		g.BufferTokenMeta(t, g.nameMetaKey, "- name", depth)
		g.BufferTokenMeta(t, "desc", "description", depth+1)
		g.BufferTokenMeta(t, "in", "in", depth+1)
		g.BufferTokenMeta(t, g.reqMetaKey, "required", depth+1)
		g.buffer.Label("schema", depth+1)
		if m, ok := t.Meta[g.typeMetaKey]; ok {
			// Array type
			if strings.HasPrefix(m, "array ") {
				g.buffer.KeyValue("type", "array", depth+2)
				g.buffer.Label("items", depth+2)
				g.buffer.KeyValue("type", strings.TrimPrefix(m, "array "), depth+3)
				// Regular type
			} else {
				g.buffer.KeyValue("type", trsSafeValue(m), depth+2)
			}
		}
	}
}

// BodySection processing
func (g *generator) BodySection(body token.Token, mediaTypes []string, depth int) {
	g.buffer.Label("requestBody", depth)
	g.buffer.Label("content", depth+1)
	for _, mt := range mediaTypes {
		g.buffer.Label(mt, depth+2)
		g.buffer.Label("schema", depth+3)
		if m, ok := body.Meta["value"]; ok {
			// Array type
			if strings.HasPrefix(m, "[]") {
				g.buffer.KeyValue("type", "array", depth+4)
				g.buffer.Label("items", depth+4)
				g.buffer.KeyValue("$ref", g.ComponentRef(strings.TrimPrefix(m, "[]")), depth+5)
				// Regular type
			} else {
				g.buffer.KeyValue("$ref", g.ComponentRef(m), depth+4)
			}
		}
	}
}

// ResponseSection processing
func (g *generator) ResponseSection(eIndex int, tokens []token.Token, depth int) {
	g.buffer.Label("responses", depth)

	// Produces media types
	mts := make([]string, 0)
	if t, ok := g.GetToken(tokens, "produce"); ok {
		if m, ok := g.TokenMeta(t, "value"); ok && len(m) > 0 {
			mts = g.ParseArray(m, g.trs["produce"]["value"])
		}
	}

	g.Response("success", mts, eIndex, tokens, depth)
	g.Response("failure", mts, eIndex, tokens, depth)
}

// Response processing
func (g *generator) Response(respType string, mts []string, eIndex int, tokens []token.Token, depth int) {
	reps := g.GetTokens(tokens, respType)
	for _, t := range reps {
		if m, ok := g.TokenMeta(t, "code"); ok {
			g.buffer.Label(m, depth+1)
			g.BufferTokenMeta(t, "desc", "description", depth+2)
			g.buffer.Label("content", depth+2)
			dataType := t.Meta["type"]

			// Plain text
			if dataType != "object" {
				g.buffer.Label("text/plain", depth+3)
				g.buffer.Label("schema", depth+4)
				g.buffer.KeyValue("type", dataType, depth+5)
				// Component
			} else {
				for _, mt := range mts {
					g.buffer.Label(mt, depth+3)
					g.buffer.Label("schema", depth+4)

					// Wrapper
					if len(g.wrappers[respType][eIndex].lines) > 0 {
						for i, l := range g.wrappers[respType][eIndex].lines {
							// Data pointer
							if g.wrappers[respType][eIndex].pos == i {
								// Array type
								if strings.HasPrefix(t.Meta["ref"], "[]") {
									g.buffer.KeyValue("type", "array", depth+7)
									g.buffer.Label("items", depth+7)
									g.buffer.KeyValue("$ref", g.ComponentRef(strings.TrimPrefix(t.Meta["ref"], "[]")), depth+8)
									// Regular type
								} else {
									g.buffer.KeyValue("$ref", g.ComponentRef(t.Meta["ref"]), depth+7)
								}

								// Wrapper prop
							} else {
								// Skip first line, i.e. name of the wrapper
								if i > 0 {
									g.buffer.Write(l, depth+4)
								}
							}
						}

						// Direct reference
					} else {
						// Array type
						if strings.HasPrefix(t.Meta["ref"], "[]") {
							g.buffer.KeyValue("type", "array", depth+5)
							g.buffer.Label("items", depth+5)
							g.buffer.KeyValue("$ref", g.ComponentRef(strings.TrimPrefix(t.Meta["ref"], "[]")), depth+6)
							// Regular type
						} else {
							g.buffer.KeyValue("$ref", g.ComponentRef(t.Meta["ref"]), depth+5)
						}
					}
				}
			}
		}
	}
}

// ResolveWrappers within the endpoint and stored them into wrappers buffer
func (g *generator) ResolveWrappers(tokens []token.Token) {
	// Success wrapper
	g.wrappers["success"] = append(g.wrappers["success"], dataWrapper{})
	index := len(g.wrappers["success"]) - 1
	if col := g.GetTokens(tokens, g.wrapperSuccessTokenKey); len(col) > 0 {
		// Parse the wrapper object
		g.wrappers["success"][index].lines = g.ParseObject("", col, 0, false)

		// Resolve the data pointer location
		// The parsed indicates the pointer location by this entry:
		// -> X, where X is the depth of the pointer
		for i, l := range g.wrappers["success"][index].lines {
			if strings.HasPrefix(l, "->") {
				g.wrappers["success"][index].pos = i
				prtDepth, _ := strconv.Atoi(strings.Split(l, " ")[1])
				g.wrappers["success"][index].depth = prtDepth
				break
			}
		}
	}

	// Error wrapper
	g.wrappers["failure"] = append(g.wrappers["failure"], dataWrapper{})
	index = len(g.wrappers["failure"]) - 1
	if col := g.GetTokens(tokens, g.wrapperErrorTokenKey); len(col) > 0 {
		// Parse the wrapper object
		g.wrappers["failure"][index].lines = g.ParseObject("", col, 0, false)

		// Resolve the data pointer location
		// The parsed indicates the pointer location by this entry:
		// -> X, where X is the depth of the pointer
		for i, l := range g.wrappers["failure"][index].lines {
			if strings.HasPrefix(l, "->") {
				g.wrappers["failure"][index].pos = i
				prtDepth, _ := strconv.Atoi(strings.Split(l, " ")[1])
				g.wrappers["failure"][index].depth = prtDepth
				break
			}
		}
	}
}

// ResolveComponents structures and store them into the components cache
func (g *generator) ResolveComponents(tokens []token.Token) {
	col := g.GetTokens(tokens, g.compTokenKeys...)
	if len(col) > 0 {
		comps := make(map[string][]token.Token, 0)
		// Prepare cache structure
		for _, t := range col {
			if key, ok := t.Meta["pkg.type"]; ok {
				if _, ok := comps[key]; ok == false {
					comps[key] = make([]token.Token, 0)
				}
				comps[key] = append(comps[key], t)
			}
		}

		for name, compDefs := range comps {
			// Already in cache
			if _, ok := g.compCache[name]; ok {
				continue
			}

			// Produce an unique nice name
			chunks := strings.Split(name, ".")
			niceNameBase := chunks[len(chunks)-1]
			niceName := niceNameBase
			i := 0
			for true {
				taken := false
				for _, n := range g.compUniqueName {
					if n == niceName {
						taken = true
						break
					}
				}
				if taken == false {
					g.compUniqueName = append(g.compUniqueName, niceName)
					break
				}
				i++
				niceName = fmt.Sprintf("%s%d", niceNameBase, i)
			}

			// Set mapping between the reference name and component name
			g.compMapping[name] = niceName

			// At this point the component definitions, i.e. tokens,
			// is mixed collection of bref, sref, fref token , i.e. one
			// object might be used many times in within the endpoint.
			// From parsing perspective, we can grad just one kind of tokens
			// and ignore rest. Therefore, reduce the collection on one
			// type here.
			reduced := make([]token.Token, 0)
			kind := compDefs[0].Type
			for _, t := range compDefs {
				if kind == t.Type {
					reduced = append(reduced, t)
				}
			}

			// Parse component object and store it into the cache
			g.compCache[name] = g.ParseObject(name, reduced, 0, false)
		}
	}
}

// ParseObject from the collection of tokens describing the object, recursively
func (g *generator) ParseObject(name string, data []token.Token, depth int, isArray bool) []string {
	// Normalize the name to a local format, i.e. removed the depth information
	// e.g. website.data.name -> name
	if chunks := strings.Split(name, "."); len(chunks) > 1 {
		name = chunks[len(chunks)-1]
	}

	// Local buffer
	b := buffer{
		indentChar: "  ",
	}

	// Base props
	b.Label(name, depth)
	if isArray {
		b.Line("type: array", depth+1)
		b.Label("items", depth+1)
		depth++
	} else {
		b.Line("type: object", depth+1)
	}

	// Distinct local and inner props
	local := make([]token.Token, 0)
	ref := make([]token.Token, 0)
	for _, t := range data {
		chunks := strings.Split(t.Meta[g.nameMetaKey], ".")
		if len(chunks) == 1 {
			local = append(local, t)
		} else {
			ref = append(ref, t)
		}
	}

	// Required fields
	if req := g.GetRequiredTokens(local); len(req) > 0 {
		b.Label("required", depth+1)
		for _, t := range req {
			if propName, ok := g.TokenMeta(t, g.nameMetaKey); ok {
				b.Line(fmt.Sprintf("- %s", propName), depth+1)
			}
		}
	}

	// Local props
	if len(local) > 0 {
		b.Label("properties", depth+1)
		for _, t := range local {
			metaType := t.Meta[g.typeMetaKey]
			metaArr := false
			if strings.HasPrefix(metaType, "array ") {
				metaArr = true
				metaType = strings.TrimPrefix(metaType, "array ")
			}

			// Reference, inner object
			if metaType == "object" {
				if ptr, ok := t.Meta[g.prtMetaKey]; ok && ptr == "true" {
					b.Label(t.Meta[g.nameMetaKey], depth+2)
					b.Line(fmt.Sprintf("-> %d", depth), 0)
				} else {
					// Flatten the inner ref props, so they will be
					// threadted in deeper parsing as a local props
					prefix := fmt.Sprintf("%s.", t.Meta[g.nameMetaKey])
					reduced := make([]token.Token, 0)
					for _, rt := range ref {
						// Take only the props with the same prefix, i.e. in the same object
						if strings.HasPrefix(rt.Meta[g.nameMetaKey], prefix) {
							rt.Meta[g.nameMetaKey] = strings.TrimPrefix(rt.Meta[g.nameMetaKey], prefix)
							reduced = append(reduced, rt)
						}
					}
					// Parse inner object props
					b.lines = append(b.lines, g.ParseObject(t.Meta[g.nameMetaKey], reduced, depth+2, metaArr)...)
				}
				// Plain props
			} else {
				b.Label(t.Meta[g.nameMetaKey], depth+2)
				if m, ok := t.Meta["desc"]; ok {
					b.KeyValue("description", trsSafeValue(m), depth+3)
				}

				// Array type
				if metaArr {
					b.KeyValue("type", "array", depth+3)
					b.Label("items", depth+3)
					b.KeyValue("type", metaType, depth+4)
					// Regular type
				} else {
					b.KeyValue("type", metaType, depth+3)
				}
			}
		}
	}

	return b.lines
}

// SiblingEndpoints found by the same url
func (g *generator) SiblingEndpoints(url string, endpoints [][]token.Token) [][]token.Token {
	found := make([][]token.Token, 0)
	for _, e := range endpoints {
		if t, ok := g.GetToken(e, "router"); ok {
			if m, ok := g.TokenMeta(t, "url"); ok {
				if m == url {
					found = append(found, e)
				}
			}
		}
	}
	return found
}

// BufferTokenMeta writes a key/value pair into the buffer
// where value is the value of the meta prop
func (g *generator) BufferTokenMeta(t token.Token, meta, key string, indent int) {
	if m, ok := g.TokenMeta(t, meta); ok {
		g.buffer.KeyValue(key, trsSafeValue(m), indent)
	}
}

// TokenMeta by meta key
func (g *generator) TokenMeta(t token.Token, key string) (string, bool) {
	if m, ok := t.Meta[key]; ok {
		return m, true
	}
	return "", false
}

// GetToken by key
func (g *generator) GetToken(col []token.Token, key string) (token.Token, bool) {
	for _, t := range col {
		if t.Key == key {
			return t, true
		}
	}
	return token.Token{}, false
}

// GetTokens by key/s
func (g *generator) GetTokens(col []token.Token, keys ...string) []token.Token {
	found := make([]token.Token, 0)
	for _, t := range col {
		for _, key := range keys {
			if t.Key == key {
				found = append(found, t)
				break
			}
		}
	}
	return found
}

// GetTokensByPrefix by prefix
func (g *generator) GetTokensByPrefix(col []token.Token, prefix string) []token.Token {
	found := make([]token.Token, 0)
	for _, t := range col {
		if strings.HasPrefix(t.Key, prefix) {
			found = append(found, t)
		}
	}
	return found
}

// GetRequiredTokens from the collection
func (g *generator) GetRequiredTokens(tokens []token.Token) []token.Token {
	found := make([]token.Token, 0)
	for _, t := range tokens {
		if r, ok := t.Meta[g.reqMetaKey]; ok && r == "true" {
			found = append(found, t)
		}
	}
	return found
}

// ParseArray from a raw content.
// Expected format is comma separated values.
// The transformation is applied on every item
func (g *generator) ParseArray(content string, trs transformation) []string {
	items := strings.Split(content, ",")
	result := make([]string, len(items))
	for i, c := range items {
		result[i] = trs(strings.TrimSpace(c))
	}
	return result
}

// ComponentRef resolved from the name mapping
func (g *generator) ComponentRef(name string) string {
	if m, ok := g.compMapping[name]; ok {
		return fmt.Sprintf("\"#/components/schemas/%s\"", m)
	}
	if g.verbose {
		log.Warnf("generator: missing component reference \"%s\"", name)
	}
	return ""
}

// NewGenerator instance
func NewGenerator(verbose bool) output.Generator {
	trsTypeClean := trsChain([]transformation{trsArray, trsSpecialChars, trsType})
	return &generator{
		verbose: verbose,
		version: "3.0.2",
		buffer: buffer{
			indentChar: "  ",
		},
		compUniqueName: make([]string, 0),
		compMapping:    make(map[string]string, 0),
		compCache:      make(map[string][]string, 0),
		compTokenKeys: []string{
			"bref", "fref", "sref",
		},
		wrappers: map[string][]dataWrapper{
			"success": make([]dataWrapper, 0),
			"failure": make([]dataWrapper, 0),
		},
		wrapperSuccessTokenKey: "swrapref",
		wrapperErrorTokenKey:   "fwrapref",
		reqMetaKey:             "req",
		nameMetaKey:            "key",
		prtMetaKey:             "ptr",
		typeMetaKey:            "type",
		trs: map[string]map[string]transformation{
			"ver": {
				"value": trsQuote,
			},
			"param": {
				"type": trsTypeClean,
			},
			"success": {
				"type": trsTypeClean,
				"code": trsQuote,
			},
			"failure": {
				"type": trsTypeClean,
				"code": trsQuote,
			},
			"sref": {
				"type": trsTypeClean,
			},
			"fref": {
				"type": trsTypeClean,
			},
			"bref": {
				"type": trsTypeClean,
			},
			"swrapref": {
				"type": trsTypeClean,
			},
			"fwrapref": {
				"type": trsTypeClean,
			},
			"accept": {
				"value": trsMediaType,
			},
			"produce": {
				"value": trsMediaType,
			},
			"router": {
				"method": trsSpecialChars,
			},
		},
	}
}

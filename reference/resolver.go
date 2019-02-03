// Package reference handles resolving of reference types within
// the API documentation. It resolves the struct types recursively.
package reference

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spaceavocado/apidoc/extract"
	log "github.com/sirupsen/logrus"
)

// Resolver of references
type Resolver interface {
	// Resolve endpoints references.
	// The resolved references are injected into the block's lines
	Resolve(endpoints []extract.Block) error
}

// refType expected by the resolver
type refType uint8

// Reference Types
const (
	typeBody refType = iota + 1
	typeResp
	typeWrap
)

// mappingType pair
type mappingType struct {
	prefix string
	t      refType
}

type resolver struct {
	verbose bool
	gopath  string
	// Builtin go primitive types
	builtinTypes []string
	// Resolved packages
	// with all resolved files
	packages map[string]map[string]resolvedFile
	// Resolved types
	types map[string][]string
	// Collection of expected prefixes with
	// mapping to produces reference prefixes
	prefixMapping map[string]mappingType
	// Struct meta fields mapping
	metaMapping    map[string]string
	structRx       *regexp.Regexp
	importSingleRx *regexp.Regexp
	importMultiRx  *regexp.Regexp
	fieldRx        *regexp.Regexp
	metaRx         *regexp.Regexp
	respRx         *regexp.Regexp
	boolRx         *regexp.Regexp
	typeCleanRx    *regexp.Regexp
}

// Resolve endpoints references.
// The resolved references are injected into the lines
func (r *resolver) Resolve(endpoints []extract.Block) error {
	for i, b := range endpoints {
		resolved := make([]string, 0)
		for _, l := range b.Lines {
			mapping := r.HasExpectedPrefix(l)

			// Body/Response reference
			if mapping.t == typeBody || mapping.t == typeResp {
				ref := ""
				// Body reference
				if mapping.t == typeBody {
					if chunks := strings.Split(l, " "); len(chunks) > 1 {
						ref = chunks[1]
					}
					// Response reference
				} else {
					if c := r.respRx.FindStringSubmatch(l); len(c) == 2 {
						ref = c[1]
					}
				}

				// Model detected
				// Resolve the reference
				if ref != "" {
					if strings.HasPrefix(ref, "[]") {
						ref = strings.TrimPrefix(ref, "[]")
					}
					entries, err := r.ResolveReference(ref, b.File, 0)
					if err != nil {
						return err
					}
					if len(entries) > 0 {
						newModel := strings.Split(entries[0], " ")[0]
						l = strings.Replace(l, ref, newModel, 1)
						r.AddPrefix(fmt.Sprintf("%s ", mapping.prefix), entries)
						resolved = append(resolved, entries...)
						resolved = append(resolved, l)
					}
					// No valid reference detected
				} else {
					resolved = append(resolved, l)
				}

				// Wrapper reference
			} else if mapping.t == typeWrap {
				tokens := strings.Split(l, " ")

				// Must have all 3 mandatory sections
				// Expected format: @token reference field-pointer
				if len(tokens) == 3 {
					ref := tokens[1]
					entries, err := r.ResolveReference(ref, b.File, 0)
					if err != nil {
						return err
					}

					for i, e := range entries {
						resolved := false
						chunks := strings.Split(e, " ")
						// Mark the pointer prop inside the wrapper entry
						// tokens[2] is the pointer field
						entries[i] = r.boolRx.ReplaceAllStringFunc(e, func(m string) string {
							resolved = true
							return fmt.Sprintf("%s %v", m, tokens[2] == chunks[1])
						})
						// It the entry is not resolved as an object local prop,
						// there is a chance that the pointer is the root prop object is self.
						// If so, mark it as the pointer
						if resolved == false && strings.HasSuffix(entries[i], fmt.Sprintf("%s {object}", tokens[2])) {
							entries[i] += " true true"
						}
					}

					if len(entries) > 0 {
						ref := strings.Split(entries[0], " ")[0]
						chunks := strings.Split(l, " ")
						r.AddPrefix(fmt.Sprintf("%s ", mapping.prefix), entries)
						resolved = append(resolved, entries...)
						resolved = append(resolved, fmt.Sprintf("%s %s %s", chunks[0], ref, chunks[2]))
					}
				}

				// Primitive type
			} else {
				resolved = append(resolved, l)
			}
		}

		// Updated the lines with the resolved enriched lines
		endpoints[i].Lines = resolved
	}
	return nil
}

// HasExpectedPrefix returns the detected
// expected prefix it the form of its output mapping
func (r *resolver) HasExpectedPrefix(line string) mappingType {
	for e, m := range r.prefixMapping {
		if strings.HasPrefix(line, fmt.Sprintf("%s ", e)) {
			return m
		}
	}
	return mappingType{}
}

// AddPrefix into entries
func (r *resolver) AddPrefix(prefix string, items []string) {
	for i := 0; i < len(items); i++ {
		items[i] = prefix + items[i]
	}
}

// PkgName from the path
func (r *resolver) PkgName(path string) string {
	return filepath.Dir(path)
}

// NormalizePkgName cleared from the GO root path,
// normalized to forward slashes, and without
// the heading slash
func (r *resolver) NormalizePkgName(name string) string {
	name = strings.TrimPrefix(name, r.gopath)
	name = strings.Replace(name, "\\", "/", -1)
	name = strings.TrimPrefix(name, "/")
	return name
}

// PkgLoc by the package prefix/key found in a resolved file
func (r *resolver) PkgLoc(prefix string, fc resolvedFile) (string, error) {
	p, ok := fc.imports[prefix]
	if ok {
		return p, nil
	}
	if r.verbose {
		log.Warnf("reference resolving: cannot resolve the location of package \"%s\" in the file \"%s\"", prefix, fc.file)
	}
	return "", errors.New("not found")
}

// ResolveReference recursively from the local files
// and from the imported packages. It returns the resolved
// documentation lines describing the references type.
func (r *resolver) ResolveReference(ref, file string, depth int) ([]string, error) {
	pkg := r.PkgName(file)

	// Parse package
	err := r.ParsePackage(pkg, pkg)
	if err != nil {
		return []string{}, err
	}

	var prefix string

	// External reference, i.e imported from an other package
	if strings.Contains(ref, ".") == true {
		chunks := strings.Split(ref, ".")
		prefix = chunks[0]
		ref = chunks[1]

		pkg = r.NormalizePkgName(pkg)

		// Resolve package
		external, err := r.PkgLoc(prefix, r.packages[pkg][file])
		if err != nil {
			return []string{}, err
		}

		// Parse package
		err = r.ParsePackage(filepath.Join(r.gopath, external), external)
		if err != nil {
			return []string{}, err
		}

		prefix = external

		// Local reference
	} else {
		chunks := strings.Split(pkg, "/")
		prefix = chunks[len(chunks)-1]
	}

	return r.ReferenceDetails(file, prefix, ref, depth)
}

// ReferenceDetails resolved from the cached type struct content
func (r *resolver) ReferenceDetails(file, pkg, ref string, depth int) ([]string, error) {
	pkg = r.NormalizePkgName(pkg)
	p, ok := r.packages[pkg]
	if ok == false {
		if r.verbose {
			log.Warnf("reference resolving: unknown package \"%s\" in the file \"%s\"", pkg, file)
		}
		return nil, fmt.Errorf("unknown package \"%s\"", pkg)
	}

	// Try to find the type matching the reference
	for _, fc := range p {
		for t, tr := range fc.types {
			if t == ref {
				return r.TypeToParams(tr.file, fmt.Sprintf("%s.%s", pkg, ref), tr.content, fc.imports, depth), nil
			}
		}
	}
	if r.verbose {
		log.Warnf("reference resolving: unknown type \"%s\" in the file \"%s\"", ref, file)
	}
	return nil, fmt.Errorf("unknown ref \"%s\"", ref)
}

// TypeToParams deconstructs the struct type into field line items
// If the type has been already resolved it retruns the cached result
func (r *resolver) TypeToParams(file, pkgname, content string, imports map[string]string, depth int) []string {
	if t, ok := r.types[pkgname]; ok {
		clone := make([]string, len(t))
		prefix := ""
		if strings.HasPrefix(t[0], pkgname) == false && depth == 0 {
			prefix = fmt.Sprintf("%s ", pkgname)
		}
		for i, l := range t {
			clone[i] = fmt.Sprintf("%s%s", prefix, l)
		}
		return clone
	}
	r.types[pkgname] = make([]string, 0)

	lines := strings.Split(strings.Replace(content, "\r\n", "\n", -1), "\n")
	entries := make([]string, 0)
	var desc string

	// Process all entries, i.e. all lines within
	// the struct body
	for _, l := range lines {
		l = strings.TrimSpace(l)

		if l != "" {
			name := ""
			t := ""
			req := "false"
			meta := make(map[string]string, 0)
			c := r.fieldRx.FindStringSubmatch(strings.TrimSpace(l))

			// Desc, i.e. comment above
			if c[0] != "" && c[1] != "" {
				desc = fmt.Sprintf("\"%s\"", c[1])
				continue
			}

			// Case 1: name type `meta`
			if c[0] != "" && c[2] != "" && c[3] != "" && c[4] != "" {
				name = c[2]
				t = c[3]
				meta = r.ParseFieldMeta(c[4])
				// Case 2: name type
			} else if c[0] != "" && c[5] != "" && c[6] != "" {
				name = c[5]
				t = c[6]
			}
			// TODO: Resolve embedded object

			// Remove pointer, decorators
			t = strings.Replace(t, "*", "", -1)
			t = strings.Replace(t, "&", "", -1)

			// Meta overrides
			if m, ok := meta[r.metaMapping["name"]]; ok {
				name = m
			}
			if m, ok := meta[r.metaMapping["type"]]; ok {
				t = m
			}
			if m, ok := meta[r.metaMapping["req"]]; ok {
				req = m
			}

			// Continue only of the name is valid.
			// I.e not empty, not marked as skipped in json
			if name != "-" && name != "" {
				// Base type
				if r.IsBasicType(t) {
					if depth == 0 {
						entries = append(entries, fmt.Sprintf("%s %s {%s} %s %s", pkgname, name, t, req, desc))
					} else {
						entries = append(entries, fmt.Sprintf("%s {%s} %s %s", name, t, req, desc))
					}

					// Recursive reference
				} else {
					rootType := "object"
					if strings.HasPrefix(t, "[]") {
						t = strings.TrimPrefix(t, "[]")
						rootType = "[]object"
					}

					if childEntries, _ := r.ResolveReference(t, file, depth+1); len(childEntries) > 0 {
						if depth == 0 {
							entries = append(entries, fmt.Sprintf("%s %s {%s}", pkgname, name, rootType))
							r.AddPrefix(fmt.Sprintf("%s %s.", pkgname, name), childEntries)
							entries = append(entries, childEntries...)
						} else {
							entries = append(entries, fmt.Sprintf("%s {%s}", name, rootType))
							r.AddPrefix(fmt.Sprintf("%s.", name), childEntries)
							entries = append(entries, childEntries...)
						}
					}
				}
			}

			desc = ""
		}
	}

	r.types[pkgname] = make([]string, len(entries))
	copy(r.types[pkgname], entries)
	return entries
}

// ParseFieldMeta associated with a struct property
func (r *resolver) ParseFieldMeta(meta string) map[string]string {
	output := make(map[string]string, 0)
	caps := r.metaRx.FindAllStringSubmatch(meta, -1)
	for _, c := range caps {
		output[c[1]] = strings.Split(c[2], ",")[0]
	}
	return output
}

// IsBasicType checks the tested type against
// the builtin go types, to determine if it is
// a primitive/basic type
func (r *resolver) IsBasicType(tested string) bool {
	tested = strings.TrimPrefix(tested, "[]")
	for _, t := range r.builtinTypes {
		if tested == t {
			return true
		}
	}
	return false
}

// ParsePackage files into the cache.
// If the package is already parsed, the processing is skipped.
func (r *resolver) ParsePackage(root, name string) error {
	name = r.NormalizePkgName(name)
	_, ok := r.packages[name]
	if ok {
		return nil
	}
	r.packages[name] = make(map[string]resolvedFile, 0)

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path == root {
				return nil
			}
			return filepath.SkipDir
		}

		if strings.HasSuffix(path, ".go") {
			return r.ParseFile(name, path)
		}
		return nil
	})
}

// ParseFile and store into the cache.
// It searches for imports and struct entities.
// Already parsed file is skipped.
func (r *resolver) ParseFile(pkg, file string) error {
	_, ok := r.packages[pkg][file]
	if ok {
		return nil
	}

	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()
	b, err := ioutil.ReadAll(fp)
	content := string(b)

	// Convert interface{} to a pseudo object, to improve RX speed and
	// complexity, and it will have the same meaning in the produces
	// documentation structure.
	content = strings.Replace(content, "interface{}", "object", -1)

	imports := make(map[string]string, 0)

	// Single import
	ic := r.importSingleRx.FindStringSubmatch(content)
	if len(ic) > 0 {
		sections := strings.Split(ic[1], "/")
		// Ignore system packages
		if len(sections) > 1 {
			imports[sections[len(sections)-1]] = ic[1]
		}

		// Multi import
	} else {
		ic = r.importMultiRx.FindStringSubmatch(content)
		if len(ic) > 0 {
			lines := strings.Split(strings.Replace(ic[1], "\r\n", "\n", -1), "\n")
			for _, l := range lines {
				l = strings.TrimSpace(l)
				if l != "" {
					chunks := strings.Split(l, " ")
					if len(chunks) == 2 {
						imports[chunks[0]] = strings.Replace(chunks[1], "\"", "", -1)
					} else {
						l = strings.Replace(l, "\"", "", -1)
						sections := strings.Split(l, "/")
						// Ingore system packages
						if len(sections) > 1 {
							imports[sections[len(sections)-1]] = l
						}
					}
				}
			}
		}
	}

	// Struct. i.e. types
	types := make(map[string]typeRef, 0)
	sc := r.structRx.FindAllStringSubmatch(content, -1)
	for _, c := range sc {
		types[c[1]] = typeRef{
			file:    file,
			content: c[2],
		}
	}

	// Store the cache
	r.packages[pkg][file] = resolvedFile{
		imports: imports,
		types:   types,
		file:    file,
	}

	return nil
}

// NewResolver instance
func NewResolver(verbose bool) Resolver {
	return &resolver{
		verbose:      verbose,
		gopath:       filepath.Join(os.Getenv("GOPATH"), "src"),
		builtinTypes: []string{"bool", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "byte", "rune", "float32", "float64", "complex64", "complex128", "object"},
		packages:     make(map[string]map[string]resolvedFile, 0),
		types:        make(map[string][]string, 0),
		prefixMapping: map[string]mappingType{
			"body":    mappingType{"bref", typeBody},
			"success": mappingType{"sref", typeResp},
			"failure": mappingType{"fref", typeResp},
			"fwrap":   mappingType{"fwrapref", typeWrap},
			"swrap":   mappingType{"swrapref", typeWrap},
		},
		metaMapping: map[string]string{
			"name": "json",
			"type": "apitype",
			"req":  "required",
		},
		structRx:       regexp.MustCompile("type\\s(.*)\\sstruct\\s?{([^}]+)}"),
		importSingleRx: regexp.MustCompile("import \"(.*)\""),
		importMultiRx:  regexp.MustCompile("import \\(([^)]+)\\)"),
		fieldRx:        regexp.MustCompile("^\\/\\/\\s?(.*)|([^\\s]+)\\s+([^\\s]+)\\s+`(.*)`|([^\\s]+)\\s+([^\\s]+)|.*"),
		metaRx:         regexp.MustCompile("([a-z]+)+:\"([^\"]+)\""),
		respRx:         regexp.MustCompile("(?:success|failure).*{object}\\s+([^\\s]+)"),
		boolRx:         regexp.MustCompile("false|true"),
		typeCleanRx:    regexp.MustCompile(".*\\."),
	}
}

package openapi

import (
	"fmt"
	"regexp"
)

// transformation performs a transform operation
type transformation func(input string) string

// trsEmpty is an empty transformation used just as
// a skip transformation
var trsEmpty transformation = func(input string) string {
	return input
}

// trsChain generates a chain of transformation
// from the send transformations
func trsChain(trs []transformation) transformation {
	return func(input string) string {
		for _, t := range trs {
			input = t(input)
		}
		return input
	}
}

// newTrsType replace GO types into OpenAPI supported types
func newTrsType() transformation {
	mapping := map[string]string{
		"byte":    "integer",
		"rune":    "integer",
		"int":     "integer",
		"int8":    "integer",
		"int16":   "integer",
		"int32":   "integer",
		"int64":   "integer",
		"uint":    "integer",
		"uint8":   "integer",
		"uint16":  "integer",
		"uint32":  "integer",
		"uint64":  "integer",
		"uintptr": "integer",
		"float32": "number",
		"float64": "number",
		"bool":    "boolean",
	}
	return func(input string) string {
		updated, ok := mapping[input]
		if ok {
			return updated
		}
		return input
	}
}

// newTrsMediaType replace short form media types
// into the OpenAPI format.
func newTrsMediaType() transformation {
	mapping := map[string]string{
		"text":         "text/plain",
		"html":         "text/html",
		"xml":          "text/xml",
		"json":         "application/json",
		"form":         "application/x-www-form-urlencoded",
		"multipart":    "multipart/form-data",
		"json-api":     "application/vnd.api+json",
		"json-stream":  "application/x-json-stream",
		"octet-stream": "application/octet-stream",
		"png":          "image/png",
		"jpeg":         "image/jpeg",
		"jpg":          "image/jpeg",
		"gif":          "image/gif",
	}
	return func(input string) string {
		updated, ok := mapping[input]
		if ok {
			return updated
		}
		return input
	}
}

// newTrsQuote wrap the input into quotes
func newTrsQuote() transformation {
	return func(input string) string {
		if input[:1] != "\"" {
			input = fmt.Sprintf("\"%s", input)
		}
		if input[len(input)-1:] != "\"" {
			input = fmt.Sprintf("%s\"", input)
		}
		return input
	}
}

// newTrsSpecialChars removes special chars from the input
func newTrsSpecialChars() transformation {
	r := regexp.MustCompile("[{}\\[\\]]")
	return func(input string) string {
		return r.ReplaceAllString(input, "")
	}
}

// newTrsArray updates array signature recognized by generator
// i.e. []type to "array type"
func newTrsArray() transformation {
	r := regexp.MustCompile("^{\\[\\]")
	return func(input string) string {
		return r.ReplaceAllString(input, "{array ")
	}
}

// transformations
var (
	trsType         = newTrsType()
	trsMediaType    = newTrsMediaType()
	trsQuote        = newTrsQuote()
	trsSpecialChars = newTrsSpecialChars()
	trsArray        = newTrsArray()
)

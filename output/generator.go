// Package output handles generation of the output API documentation
// from the tokenized outcome.
package output

import "github.com/spaceavocado/apidoc/token"

// Generator of the documentation
type Generator interface {
	// Generate the documentation from the given tokens for the
	// main section and for the given endpoints, into the file.
	Generate(main []token.Token, endpoints [][]token.Token, file string) error
}

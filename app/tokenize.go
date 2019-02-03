package app

import (
	"github.com/spaceavocado/apidoc/token"
)

// TokenizationResult produced by token parser
type TokenizationResult struct {
	// Main holds tokens for the main API docmentation section
	Main []token.Token
	// Endpoints holds tokens each docmentation endpoint
	Endpoints [][]token.Token
}

// Tokenize the extracted documentation
func (a *App) Tokenize(extract ExtractResult) (TokenizationResult, error) {
	r := TokenizationResult{}

	// Main block
	tokens, err := a.tokenParser.Parse(extract.Main)
	if err != nil {
		return r, err
	}
	r.Main = tokens

	// Endpoints
	for _, b := range extract.Endpoints {
		tokens, err = a.tokenParser.Parse(b)
		if err != nil {
			continue
		}
		r.Endpoints = append(r.Endpoints, tokens)
	}

	return r, nil
}

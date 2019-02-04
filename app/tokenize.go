package app

import (
	"fmt"

	"github.com/spaceavocado/apidoc/token"
)

// RequiredMainTokens to be present in the main block
var requiredMainTokens = []string{"title", "ver"}

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

	// Validate required tokens
	// i.e. the Main must contains those tokens
	for _, rt := range requiredMainTokens {
		found := false
		for _, t := range tokens {
			if t.Key == rt {
				found = true
				break
			}
		}
		if found == false {
			return r, fmt.Errorf("missing required \"%s\" token in the main file", rt)
		}
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

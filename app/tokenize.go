package app

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spaceavocado/apidoc/token"
)

// RequiredMainTokens to be present in the main block
var requiredMainTokens = []string{"title", "ver"}

// RequiredEndpointTokens to be present in the endpoint block
var requiredEndpointTokens = []string{"router", "produce", "success"}

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

// ReduceEndpoints filters out invalid endpoints
func (a *App) ReduceEndpoints(endpoints [][]token.Token) [][]token.Token {
	reduced := make([][]token.Token, 0)
	for _, e := range endpoints {
		valid := false
		// Validate required tokens
		// i.e. must contains those tokens
		for _, rt := range requiredEndpointTokens {
			valid = false
			for _, t := range e {
				if t.Key == rt {
					valid = true
					break
				}
			}
			if valid == false {
				break
			}
		}
		if valid {
			reduced = append(reduced, e)
		} else if a.conf.Verbose {
			log.Warnf("Ivalid endpoint detected")
		}
	}
	return reduced
}

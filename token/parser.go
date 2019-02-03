// Package token handles tokenization of the raw
// API documentation lines.
package token

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spaceavocado/apidoc/extract"
	log "github.com/sirupsen/logrus"
)

// ErrParsing is soft error not causing the failure
// of the whole procedure. e.g. invalid token, etc.
var errParsing = errors.New("parsing error")

// Token produced by the tokenization process.
//
// Meta is a collection of all found and expected token
// meta properties defined by token dictionairy.
type Token struct {
	Type Type
	Key  string
	Meta map[string]string
}

// Type of the token
// from parsing perspective, i.e. shared dictionires
type Type uint8

// Token types
const (
	Value Type = iota
	Param
	Server
	ReqResp
	Method
	Ref
	Wrap
)

// Parser of the tokens
type Parser interface {
	// Parse the raw extracted blocks into tokens
	Parse(b extract.Block) ([]Token, error)
}

type parser struct {
	verbose bool
	// Mapping between expected token identifiers and its type
	typeMapping map[string]Type
	// Token type parsing dictionary
	typeDic         map[Type]dic
	tokenSectionsRx *regexp.Regexp
}

// Parse a raw extracted block into tokens
func (p *parser) Parse(b extract.Block) ([]Token, error) {
	tokens := make([]Token, 0, len(b.Lines))
	for _, line := range b.Lines {
		t, err := p.Tokenize(line)
		if err == errParsing {
			continue
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// Tokenize from a raw API documentation line
func (p *parser) Tokenize(line string) (Token, error) {
	sections := p.tokenSectionsRx.FindAllString(line, -1)
	if len(sections) == 0 {
		if p.verbose {
			log.Warnf("tokenization: cannot tokenize this line: %s", line)
		}
		return Token{}, errParsing
	}

	// Token type
	t, ok := p.typeMapping[sections[0]]
	if ok == false {
		if p.verbose {
			log.Warnf("tokenization: unknown token type: %s", sections[0])
		}
		return Token{}, errParsing
	}

	// Token meta dic
	dic, ok := p.typeDic[t]
	if ok == false {
		if p.verbose {
			log.Warnf("tokenization: missing token meta dic for type: %s", sections[0])
		}
		return Token{}, errParsing
	}

	// Parse token meta
	token := Token{
		Type: t,
		Key:  sections[0],
		Meta: dic.Map(sections[1:]),
	}

	return token, nil
}

// NewParser instance
func NewParser(verbose bool) Parser {
	return &parser{
		verbose: verbose,
		typeMapping: map[string]Type{
			// Main block
			"title":         Value,
			"desc":          Value,
			"terms":         Value,
			"contact.name":  Value,
			"contact.url":   Value,
			"contact.email": Value,
			"lic.name":      Value,
			"lic.url":       Value,
			"ver":           Value,
			"server":        Server,

			// Endpoint
			"summary":  Value,
			"id":       Value,
			"tag":      Value,
			"accept":   Value,
			"produce":  Value,
			"param":    Param,
			"sref":     Ref,
			"swrap":    Value,
			"fref":     Ref,
			"fwrap":    Value,
			"swrapref": Wrap,
			"fwrapref": Wrap,
			"bref":     Ref,
			"body":     Value,
			"success":  ReqResp,
			"failure":  ReqResp,
			"router":   Method,
		},
		typeDic: map[Type]dic{
			Value: dic{
				mapping: map[int]string{
					0: "value",
				},
			},
			Param: dic{
				mapping: map[int]string{
					0: "key",
					1: "in",
					2: "type",
					3: "req",
					4: "desc",
				},
			},
			Server: dic{
				mapping: map[int]string{
					0: "url",
					1: "desc",
				},
			},
			ReqResp: dic{
				mapping: map[int]string{
					0: "code",
					1: "type",
					2: "ref",
					3: "desc",
				},
				after: func(meta map[string]string) map[string]string {
					if meta["type"] == "{string}" || meta["type"] == "string" {
						meta["desc"] = strings.TrimSpace(meta["ref"] + " " + meta["desc"])
						meta["ref"] = ""
					}
					return meta
				},
			},
			Method: dic{
				mapping: map[int]string{
					0: "url",
					1: "method",
				},
			},
			Ref: dic{
				mapping: map[int]string{
					0: "pkg.type",
					1: "key",
					2: "type",
					3: "req",
					4: "desc",
				},
			},
			Wrap: dic{
				mapping: map[int]string{
					0: "pkg.type",
					1: "key",
					2: "type",
					3: "req",
					4: "ptr",
					5: "desc",
				},
			},
		},
		tokenSectionsRx: regexp.MustCompile("\"[^\"]*\"|[^\\s]+"),
	}
}

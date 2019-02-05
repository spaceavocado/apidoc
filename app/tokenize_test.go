package app

import (
	"bytes"
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spaceavocado/apidoc/extract"
	"github.com/spaceavocado/apidoc/token"
)

type mockParser struct {
	returns int
	tokens  [][]token.Token
	err     []error
}

func (p *mockParser) Parse(b extract.Block) ([]token.Token, error) {
	p.returns++
	return p.tokens[p.returns], p.err[p.returns]
}

func TestTokenize(t *testing.T) {
	// Valid
	a := New(Configuration{})
	_, err := a.Tokenize(ExtractResult{
		Main: extract.Block{
			Lines: []string{
				"title Sample API",
				"ver 1.0",
			},
		},
		Endpoints: []extract.Block{
			{
				Lines: []string{
					"summary Sample endpoint",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	// Missing required tokens
	_, err = a.Tokenize(ExtractResult{
		Main: extract.Block{
			Lines: []string{
				"title Sample API",
			},
		},
		Endpoints: make([]extract.Block, 0),
	})
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	// Errors
	a.tokenParser = &mockParser{
		returns: -1,
		tokens:  [][]token.Token{make([]token.Token, 0)},
		err:     []error{errors.New("simulated error")},
	}
	_, err = a.Tokenize(ExtractResult{})
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	a.tokenParser = &mockParser{
		returns: -1,
		tokens: [][]token.Token{
			{
				{Key: "title"},
				{Key: "ver"},
			},
			make([]token.Token, 0),
		},
		err: []error{
			nil,
			errors.New("simulated error"),
		},
	}
	_, err = a.Tokenize(ExtractResult{
		Endpoints: make([]extract.Block, 1),
	})
	if err != nil {
		t.Errorf("Expected error, got nil")
		return
	}
}

func TestReduceEndpoints(t *testing.T) {
	a := New(Configuration{})
	res := a.ReduceEndpoints([][]token.Token{
		// Valid
		{
			{Key: "success"},
			{Key: "produce"},
			{Key: "router"},
			{Key: "summary"},
		},
		// Invalid
		{
			{Key: "success"},
			{Key: "router"},
			{Key: "summary"},
		},
	})
	if len(res) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(res))
	}

	// Verbose
	b := &bytes.Buffer{}
	log.SetOutput(b)
	hook := test.NewGlobal()
	a = New(Configuration{
		Verbose: true,
	})
	res = a.ReduceEndpoints([][]token.Token{
		// Valid
		{
			{Key: "success"},
			{Key: "produce"},
			{Key: "router"},
			{Key: "summary"},
		},
		// Invalid
		{
			{Key: "success"},
			{Key: "router"},
			{Key: "summary"},
		},
	})
	if len(res) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(res))
	}
	if len(hook.Entries) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(hook.Entries))
	}
}

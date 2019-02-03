package extract

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestExtract(t *testing.T) {
	content :=
		`
	// @summary Refresh ID Token
	// @desc Use the refresh token
	// to receive a new ID token.
	// It must be in a valid format.
	`
	file := "tmp"

	err := ioutil.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	defer func() {
		os.Remove(file)
	}()

	e := NewExtractor(false).(*extractor)
	blocks, err := e.Extract(file)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	if len(blocks) == 0 {
		t.Errorf("No blocks have been extracted.")
		return
	}

	_, err = e.Extract("not-existing-file")
	if err == nil {
		t.Errorf("An error is expected, got nil")
		return
	}
}

func TestParse(t *testing.T) {
	// Various indentation for teting
	test := []string{
		`
		// ValidateToken request
			// @summary Validate ID Token
		//   @desc Perform a validation of the ID Token
// @ID validate
		// @tag Token
		// @accept json, multipart/form-data, application/x-www-form-urlencoded
		// 
		// @produce json
		// @body []model.Validate
		// @success 200 {object} response.OK OK
		// @failure 400 {object} response.Error "Bad Request"
		// @failure 500 {string} string "Internal Server Error"
		// @router /validate [post]

		// @summary Refresh ID Token
		// @desc Use the refresh token
		// to receive a new ID token.
		// It must be in a valid format.
		// @accept json
		// @produce json
		// @success 200 {object} token.Token OK
		// all if set
		// and good to go
		// @router /token [post]
		func RefreshToken(a *app.App) request.HandlerFunc {
		}

		// @summary Refresh ID Token
		// @desc Use the refresh token
		// to receive a new ID token.
		// It must be in a valid format.
		
		func test() {}
		`,

		// Gorilla mux handler with methods
		`
		// @summary Refresh ID Token
		// @desc Use the refresh token
		// to receive a new ID token.
		// It must be in a valid format.
		r.HandleFunc("/person/{id:[0-9]+}", GetPerson).Methods("GET")
		`,

		// Gorilla mux handler without methods
		`
		// @summary Refresh ID Token
		// @desc Use the refresh token.
		r.Handle("/person/{id:[0-9]+}", GetPerson)
		`,
	}

	expected := []string{
		"summary Validate ID Token",
		"desc Perform a validation of the ID Token",
		"ID validate",
		"tag Token",
		"accept json, multipart/form-data, application/x-www-form-urlencoded",
		"produce json",
		"body []model.Validate",
		"success 200 {object} response.OK OK",
		"failure 400 {object} response.Error \"Bad Request\"",
		"failure 500 {string} string \"Internal Server Error\"",
		"router /validate [post]",
	}

	e := NewExtractor(false).(*extractor)
	r := bufio.NewReaderSize(strings.NewReader(test[0]), 0)

	blocks, err := e.parse(r, "test.go")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	if len(blocks) == 0 {
		t.Errorf("No blocks have been extracted.")
		return
	}

	if len(blocks) != 3 {
		t.Errorf("Incorrect number of blocks have been captured. Expected: %d, got: %d", 2, len(blocks))
		return
	}

	for i, e := range expected {
		if len(blocks[0].Lines) <= i {
			t.Errorf("Expected: \"%s\", got: null", e)
			break
		}
		if e != blocks[0].Lines[i] {
			t.Errorf("Expected: \"%s\", got: \"%s\"", e, blocks[0].Lines[i])
			break
		}
	}

	if len(expected) != len(blocks[0].Lines) {
		t.Errorf("Has %d additional lines", len(blocks[0].Lines)-len(expected))
	}
	if 6 != len(blocks[1].Lines) {
		t.Errorf("Has %d additional lines", len(blocks[1].Lines)-6)
	}

	// Multiline entry
	if blocks[1].Lines[1] != "desc Use the refresh token to receive a new ID token. It must be in a valid format." {
		t.Errorf("Invalid multiline parsing, has \"%s\", expected \"%s\"", blocks[1].Lines[1], "desc Use the refresh token to receive a new ID token. It must be in a valid format.")
	}
	if blocks[1].Lines[4] != "success 200 {object} token.Token OK all if set and good to go" {
		t.Errorf("Invalid multiline parsing, has \"%s\", expected \"%s\"", blocks[1].Lines[4], "success 200 {object} token.Token OK all if set and good to go")
	}
	if blocks[2].Lines[1] != "desc Use the refresh token to receive a new ID token. It must be in a valid format." {
		t.Errorf("Invalid multiline parsing, has \"%s\", expected \"%s\"", blocks[2].Lines[1], "desc Use the refresh token to receive a new ID token. It must be in a valid format.")
	}

	// Gorilla mux handler
	r = bufio.NewReaderSize(strings.NewReader(test[1]), 0)
	blocks, err = e.parse(r, "test.go")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(blocks) == 0 {
		t.Errorf("No blocks have been extracted.")
		return
	}
	if len(blocks[0].Lines) != 4 {
		t.Errorf("Incorrect number of blocks have been captured. Expected: %d, got: %d", 4, len(blocks[0].Lines))
		return
	}

	r = bufio.NewReaderSize(strings.NewReader(test[2]), 0)
	blocks, err = e.parse(r, "test.go")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(blocks) == 0 {
		t.Errorf("No blocks have been extracted.")
		return
	}
	if len(blocks[0].Lines) != 4 {
		t.Errorf("Incorrect number of blocks have been captured. Expected: %d, got: %d", 4, len(blocks[0].Lines))
		return
	}
	if blocks[0].Lines[2] != "router /person/{id} [get]" {
		t.Errorf("Invalid multiline parsing, has \"%s\", expected \"%s\"", blocks[0].Lines[2], "router /person/{id} [get]")
	}
}

func TestGorillaMuxHandler(t *testing.T) {
	e := NewExtractor(false).(*extractor)
	testBlocks := []Block{
		// A block without router param
		{
			Lines: []string{
				"summary Person",
			},
		},
		// Block with router
		{
			Lines: []string{
				"router /url [get]",
			},
		},
		// Block with set param
		{
			Lines: []string{
				"param id path {int} true ID",
			},
		},
	}

	testCaptures := [][]string{
		{
			"/section/{id:[0-9]+}/{username}",
			"\"GET\", \"POST\"",
		},
		{
			"/section",
			"\"GET\"",
		},
	}

	// Expected injection
	res := e.gorillaMuxHandler(testBlocks[0], testCaptures[0][0], testCaptures[0][1])
	if len(res.Lines) != 4 {
		t.Errorf("Unexpected count of entries, has %d, expected %d", len(res.Lines), 4)
		return
	}
	if res.Lines[1] != "router /section/{id}/{username} [get, post]" {
		t.Errorf("Has %s, expected %s", res.Lines[1], "router /section/{id}/{username} [get, post]")
	}
	if res.Lines[2] != "param id path {string} true" {
		t.Errorf("Has %s, expected %s", res.Lines[2], "param id path {string} true")
	}
	if res.Lines[3] != "param username path {string} true" {
		t.Errorf("Has %s, expected %s", res.Lines[3], "param username path {string} true")
	}

	res = e.gorillaMuxHandler(testBlocks[0], testCaptures[1][0], testCaptures[1][1])
	if len(res.Lines) != 2 {
		t.Errorf("Unexpected count of entries, has %d, expected %d", len(res.Lines), 2)
		return
	}

	// Skip, router is already defined
	res = e.gorillaMuxHandler(testBlocks[1], testCaptures[0][0], testCaptures[0][1])
	if len(res.Lines) != 1 {
		t.Errorf("Unexpected count of entries, has %d, expected %d", len(res.Lines), 1)
		return
	}
	if res.Lines[0] != "router /url [get]" {
		t.Errorf("Has %s, expected %s", res.Lines[0], "router /url [get]")
	}

	// Skip defined param
	res = e.gorillaMuxHandler(testBlocks[2], testCaptures[0][0], testCaptures[0][1])
	if len(res.Lines) != 3 {
		t.Errorf("Unexpected count of entries, has %d, expected %d", len(res.Lines), 3)
		return
	}
	if res.Lines[1] != "router /section/{id}/{username} [get, post]" {
		t.Errorf("Has %s, expected %s", res.Lines[1], "router /section/{id}/{username} [get, post]")
	}
	if res.Lines[2] != "param username path {string} true" {
		t.Errorf("Has %s, expected %s", res.Lines[2], "param username path {string} true")
	}
}

func TestVerbose(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	hook := test.NewGlobal()

	e := NewExtractor(true).(*extractor)
	tests := []string{
		`
		// @summary Refresh ID Token
		// @param id path {int} true User ID
		// @desc Use the refresh token.
		r.Handle("/person/{id:[0-9]+}", GetPerson)
		`,
		`
		// @summary Refresh ID Token
		// @param id path {int} true User ID
		// @router /person/{id} get 
		r.Handle("/person/{id:[0-9]+}", GetPerson)
		`,
	}

	expected := []string{
		"extracting, Handler func: param",
		"extracting, Handler func: router",
	}

	for i, test := range tests {
		hook.Reset()
		r := bufio.NewReaderSize(strings.NewReader(test), 0)
		_, err := e.parse(r, "test.go")
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			break
		}

		if len(hook.Entries) != 1 {
			t.Errorf("Has %d logs, expected %d logs", len(hook.Entries), 1)
			break
		}

		o, err := hook.Entries[0].String()
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			break
		}

		if strings.Contains(o, expected[i]) == false {
			t.Errorf("Has %s, expected %s", o, expected[i])
		}
	}
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read break")
}

func TestReader(t *testing.T) {
	e := NewExtractor(true).(*extractor)
	r := bufio.NewReaderSize(&errorReader{}, 0)
	blocks, err := e.parse(r, "test.go")
	if err == nil {
		t.Errorf("Error is expected, got nil")
		return
	}
	if len(blocks) != 0 {
		t.Errorf("Unexpected parsing")
	}
}

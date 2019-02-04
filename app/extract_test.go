package app

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spaceavocado/apidoc/extract"
)

type mockExtractor struct {
	index int
	data  []extract.Block
}

func (e *mockExtractor) Extract(file string) ([]extract.Block, error) {
	e.index++
	if e.index == 1 {
		return e.data, nil
	}
	return []extract.Block{}, errors.New("simulated error")
}

func TestExtract(t *testing.T) {
	content := []string{
		`
		// @summary Refresh ID Token
		// @desc Use the refresh token
		// to receive a new ID token.
		// It must be in a valid format.
		`,
		"func main(){}",
		"func main(){}",
		`
		// @summary Refresh ID Token
		// @desc Use the refresh token
		// to receive a new ID token.
		// It must be in a valid format.

		code := 3
		// @summary Endpoint 
		`,
	}
	files := []string{
		"tmp1",
		"tmp2",
		"tmp/tmp.go",
		"tmp3",
	}
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	for i, f := range files {
		err := ioutil.WriteFile(f, []byte(content[i]), 0644)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			return
		}
	}
	defer func() {
		for _, f := range files {
			os.Remove(f)
		}
		os.Remove("tmp")
	}()

	// Valid main
	a := New(Configuration{
		MainFile: files[0],
		EndsRoot: "./",
	})
	_, err = a.Extract()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	// Empty main
	a = New(Configuration{
		MainFile: files[1],
		EndsRoot: "./",
	})
	_, err = a.Extract()
	if err.Error() != "no API documentation found in the root file" {
		t.Errorf("Unexpected error %v", err)
		return
	}

	// Invalid main
	a = New(Configuration{
		MainFile: "",
		EndsRoot: "./",
	})
	_, err = a.Extract()
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	// Endpoint file error
	a = New(Configuration{
		MainFile: files[0],
		EndsRoot: "tmp/",
	})
	a.extractor = &mockExtractor{
		data:  make([]extract.Block, 1),
		index: 0,
	}
	_, err = a.Extract()
	if err == nil {
		t.Errorf("Expected error, got nil")
		return
	}

	// Endpoint from the main block
	a = New(Configuration{
		MainFile: files[3],
		EndsRoot: "tmp/",
	})
	res, err := a.Extract()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if len(res.Main.Lines) == 0 || len(res.Endpoints) == 0 {
		t.Errorf("Unexpected 1 Main, 1 Endpoint")
		return
	}
}

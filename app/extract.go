package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spaceavocado/apidoc/extract"
)

// ExtractResult of the documentation extraction
type ExtractResult struct {
	// Main API docmentation block
	Main extract.Block
	// Endpoints API documentation blocks
	Endpoints []extract.Block
}

// Extract the documentation
func (a *App) Extract() (ExtractResult, error) {
	r := ExtractResult{}

	// Root file
	blocks, err := a.extractor.Extract(a.conf.MainFile)
	if err != nil {
		return r, err
	}
	if len(blocks) == 0 {
		return r, errors.New("no API documentation found in the root file")
	}
	r.Main = blocks[0]

	// If there is more than one block
	// Take the first one as the Main block and put other into Endpoints
	if len(blocks) > 1 {
		r.Endpoints = append(r.Endpoints, blocks[1:]...)
	}

	// Endpoint files
	err = filepath.Walk(a.conf.EndsRoot, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || strings.HasSuffix(path, ".go") == false {
			return nil
		}

		blocks, err = a.extractor.Extract(path)
		if err != nil {
			return err
		}
		if len(blocks) != 0 {
			r.Endpoints = append(r.Endpoints, blocks...)
		}
		return nil
	})

	if err != nil {
		return r, err
	}
	return r, nil
}

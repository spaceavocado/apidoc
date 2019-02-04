// Package app contains the app base
package app

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spaceavocado/apidoc/extract"
	"github.com/spaceavocado/apidoc/output"
	"github.com/spaceavocado/apidoc/output/openapi"
	"github.com/spaceavocado/apidoc/reference"
	"github.com/spaceavocado/apidoc/token"
)

const (
	// Version of the APIDoc
	Version string = "beta-0.2.4"
)

// App main structure
type App struct {
	conf        *Configuration
	extractor   extract.Extractor
	tokenParser token.Parser
	refResolver reference.Resolver
	generator   output.Generator
}

// Start the application
func (a *App) Start() {
	// Extract documentation
	eRes, err := a.Extract()
	if err != nil {
		log.WithError(err).Errorf("an error has occurred during the extracting procedure")
		return
	}

	// Resolve references
	err = a.refResolver.Resolve(eRes.Endpoints)
	if err != nil {
		log.WithError(err).Errorf("an error has occurred during the reference resolving procedure")
		return
	}

	// Tokenize
	tRes, err := a.Tokenize(eRes)
	if err != nil {
		log.WithError(err).Errorf("an error has occurred during the tokenization procedure")
		return
	}

	// Generate
	err = a.generator.Generate(tRes.Main, tRes.Endpoints, filepath.Join(a.conf.Output, "openapi.yaml"))
	if err != nil {
		log.WithError(err).Errorf("an error has occurred during the generation of the output")
		return
	}
}

// New application instance
func New(c Configuration) App {
	return App{
		conf:        &c,
		extractor:   extract.NewExtractor(c.Verbose),
		tokenParser: token.NewParser(c.Verbose),
		refResolver: reference.NewResolver(c.Verbose),
		generator:   openapi.NewGenerator(c.Verbose),
	}
}

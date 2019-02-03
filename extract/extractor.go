// Copyright David Horak <info@davidhorak.com>
// All Rights Reserved

// Package extract handles extraction of all raw API documentation lines
// and arrange them into main documentation block and a collection of
// endpoint documentation blocks.
package extract

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Block extracted raw documentation lines
type Block struct {
	// File containing this block
	File string
	// Lines extracted from the file
	Lines []string
}

// Extractor interface
type Extractor interface {
	// Extract the documentation from the file
	Extract(path string) ([]Block, error)
}

type extractor struct {
	verbose             bool
	commentRx           *regexp.Regexp
	apiDocRx            *regexp.Regexp
	gorillaMuxHandlerRx *regexp.Regexp
	pathCleanRx         *regexp.Regexp
	pathParamRx         *regexp.Regexp
}

// Extract the documentation from the file.
// A Block is defined an uniterupted comments section
// in which there is at least one comment with "@" prefix.
func (e *extractor) Extract(file string) ([]Block, error) {
	fp, err := os.Open(file)
	if err != nil {
		return []Block{}, err
	}
	defer fp.Close()
	return e.parse(bufio.NewReader(fp), file)
}

// Parse file content
func (e *extractor) parse(r *bufio.Reader, file string) ([]Block, error) {
	blocks := []Block{}
	isComment := false
	commentBuffer := ""
	block := Block{
		File: file,
	}

	for {
		var err error
		var buffer bytes.Buffer
		var b []byte
		var isPrefix bool
		for {
			b, isPrefix, err = r.ReadLine()
			buffer.Write(b)
			if isPrefix == false || err != nil {
				break
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return blocks, err
		}
		line := buffer.String()

		// Comment line
		if cm := e.commentRx.FindStringSubmatch(line); len(cm) > 0 {
			if !isComment {
				isComment = true
				if len(block.Lines) > 0 {
					block = Block{
						File: file,
					}
				}
			}

			// API comment
			if m := e.apiDocRx.FindStringSubmatch(cm[1]); len(m) > 0 {
				block.Lines = append(block.Lines, m[1])
				if len(block.Lines) > 2 && len(commentBuffer) > 0 {
					block.Lines[len(block.Lines)-2] += commentBuffer
				}
				commentBuffer = ""
				// Multiline comment
			} else if comment := strings.TrimSpace(cm[1]); comment != "" {
				commentBuffer += fmt.Sprintf(" %s", comment)
			}

			// Break the block
		} else if isComment {
			isComment = false
			if len(block.Lines) > 0 {
				if len(commentBuffer) > 0 {
					block.Lines[len(block.Lines)-1] += commentBuffer
				}
				commentBuffer = ""
				blocks = append(blocks, block)
			}

			// Gorilla mux router
			if m := e.gorillaMuxHandlerRx.FindStringSubmatch(line); len(m) > 0 {
				url := ""
				methods := ""
				if m[3] != "" {
					url = m[3]
				} else {
					url = m[1]
					methods = m[2]
				}
				blocks[len(blocks)-1] = e.gorillaMuxHandler(blocks[len(blocks)-1], url, methods)
			}
		}
	}

	return blocks, nil
}

// GorillaMuxHandler parsing resolver
// It tries to inject the information form the handler function signature
// into the block. Current supported inputs are router path, params from
// the path, and the methods
func (e *extractor) gorillaMuxHandler(b Block, url, methods string) Block {
	// If the block already contains a router annotation
	// skip this processing, since it gas higher priority
	for _, l := range b.Lines {
		if strings.HasPrefix(l, "router ") {
			if e.verbose {
				log.Warnf("extracting, Handler func: router \"%s\" is already defined in the endpoint annotation, skipped.", url)
			}
			return b
		}
	}

	if methods == "" {
		methods = "get"
	} else {
		methods = strings.Replace(methods, "\"", "", -1)
		methods = strings.ToLower(methods)
	}
	url = e.pathCleanRx.ReplaceAllString(url, "")
	b.Lines = append(b.Lines, fmt.Sprintf("router %s [%s]", url, methods))

	params := make([]string, 0)
	if m := e.pathParamRx.FindAllStringSubmatch(url, -1); len(m) > 0 {
		for _, sm := range m {
			params = append(params, sm[1])
		}
	}

	for _, p := range params {
		skip := false
		for _, l := range b.Lines {
			// If the param is already defined, ignore the parsed one from URL
			if strings.HasPrefix(l, fmt.Sprintf("param %s ", p)) {
				if e.verbose {
					log.Warnf("extracting, Handler func: param \"%s\" defined it the handler url \"%s\" is already defined in the endpoint annotation, skipped.", p, url)
				}
				skip = true
				continue
			}
		}
		if skip == false {
			b.Lines = append(b.Lines, fmt.Sprintf("param %s path {string} true", p))
		}
	}

	return b
}

// NewExtractor instance
func NewExtractor(verbose bool) Extractor {
	return &extractor{
		verbose:             verbose,
		commentRx:           regexp.MustCompile("^\\s*\\/\\/\\s*(.*)"),
		apiDocRx:            regexp.MustCompile("^@([^\\s].*)"),
		gorillaMuxHandlerRx: regexp.MustCompile("(?:HandleFunc|Handle)\\(\"([^\"]+)\".*\\.Methods\\(([^\\)]+)\\)|(?:HandleFunc|Handle)\\(\"([^\"]+)\""),
		pathCleanRx:         regexp.MustCompile(":[^}]+"),
		pathParamRx:         regexp.MustCompile("{([^}]+)}"),
	}
}

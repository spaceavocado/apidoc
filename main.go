// Copyright © 2019 David Horak <info@davidhorak.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package main is the main file of the API Documentation Generator.
// It executes the root CMD.
package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spaceavocado/apidoc/cmd"
	"github.com/spaceavocado/apidoc/misc"
)

func main() {
	log.SetFormatter(&misc.PlainLogFormatter{})
	if err := cmd.RootCmd().Execute(); err != nil {
		log.WithError(err).Error("unexpected error")
	}
}

// Package misc is a collection of miscellaneous funcions
package misc

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// PlainLogFormatter for github.com/sirupsen/logrus logger
type PlainLogFormatter struct{}

// Format the log entry
// It ingores the entry data not relevant for APIDoc logging
func (f *PlainLogFormatter) Format(entry *log.Entry) ([]byte, error) {
	switch entry.Level {
	case log.WarnLevel:
		entry.Message = fmt.Sprintf("WARNING: %s", entry.Message)
	case log.ErrorLevel:
		fallthrough
	case log.FatalLevel:
		fallthrough
	case log.PanicLevel:
		entry.Message = fmt.Sprintf("ERROR: %s", entry.Message)
	default:
	}

	log := []byte(entry.Message)

	for k, v := range entry.Data {
		log = append(log, []byte(fmt.Sprintf(":\n  %s: %+v", k, v))...)
	}

	return append(log, '\n'), nil
}

package misc

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestFormat(t *testing.T) {
	f := PlainLogFormatter{}

	tests := []log.Entry{
		{
			Level:   log.WarnLevel,
			Message: "msg",
		},
		{
			Level:   log.ErrorLevel,
			Message: "msg",
		},
		{
			Level:   log.PanicLevel,
			Message: "msg",
		},
		{
			Level:   log.FatalLevel,
			Message: "msg",
		},
		{
			Level:   log.InfoLevel,
			Message: "msg",
		},
		{
			Level:   log.InfoLevel,
			Message: "msg",
			Data: map[string]interface{}{
				"error": errors.New("err msg"),
			},
		},
	}

	expected := []string{
		"WARNING: msg\n",
		"ERROR: msg\n",
		"ERROR: msg\n",
		"ERROR: msg\n",
		"msg\n",
		"msg:\n  error: err msg\n",
	}

	for i, c := range tests {
		b, err := f.Format(&c)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
			return
		}
		if expected[i] != string(b) {
			t.Errorf("Expected \"%s\" error, got \"%s\"", expected[i], string(b))
		}
	}
}

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestMain(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)

	// Unexpected command error
	hook := test.NewGlobal()
	os.Args = []string{"apidoc", "missing-command"}
	main()
	if len(hook.Entries) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(hook.Entries))
	}
	_, err := hook.Entries[0].String()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if strings.Contains(b.String(), "unknown command") == false {
		t.Errorf("Unexpected unknown command error, got nil")
		return
	}

	// Runs the command
	hook.Reset()
	os.Args = []string{"apidoc", "-m", "not-existing-file"}
	main()
	if len(hook.Entries) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(hook.Entries))
	}
	o, err := hook.Entries[0].String()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}
	if strings.Contains(o, "an error has occurred during the extracting procedure") == false {
		t.Errorf("Expected \"%s\" error, got \"%s\"", "an error has occurred during the extracting procedure", o)
	}
}

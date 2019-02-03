package cmd

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cobra"
)

func TestVersion(t *testing.T) {
	versionCmd.Run(&cobra.Command{}, []string{""})
}

func TestRootCmd(t *testing.T) {
	b := &bytes.Buffer{}
	log.SetOutput(b)
	hook := test.NewGlobal()
	cmd := RootCmd()
	cmd.Run(&cobra.Command{}, []string{""})
	if len(hook.Entries) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(hook.Entries))
	}

	hook.Reset()
	c := cobra.Command{}
	c.PersistentFlags().StringP("main", "m", "not-existing-file", "")
	c.PersistentFlags().StringP("endpoints", "e", "./", "")
	c.PersistentFlags().StringP("output", "o", "docs/api", "")
	c.PersistentFlags().BoolP("verbose", "v", false, "")

	cmd = RootCmd()
	cmd.Run(&c, []string{""})
	if len(hook.Entries) != 1 {
		t.Errorf("Expected %d log entries, got %d", 1, len(hook.Entries))
	}
}

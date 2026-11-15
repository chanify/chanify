//go:build !test

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionEntrypoint(t *testing.T) {
	original := os.Args
	config := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(config, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	os.Args = []string{"chanify", "--config", config, "version", "--format", "{{.Client.Name}} {{.Client.Version}}"}
	t.Cleanup(func() { os.Args = original })
	main()
}

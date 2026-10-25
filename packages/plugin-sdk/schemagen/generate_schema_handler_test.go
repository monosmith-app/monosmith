package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesThenFindsItUpToDate(t *testing.T) {
	out := filepath.Join(t.TempDir(), "schema", "protocol.schema.json")
	if err := run("../protocol", out, true); err == nil || !strings.Contains(err.Error(), "not up to date") {
		t.Fatalf("-check on a missing file: %v", err)
	}
	if err := run("../protocol", out, false); err != nil {
		t.Fatal(err)
	}
	if err := run("../protocol", out, true); err != nil {
		t.Fatalf("-check on a fresh file: %v", err)
	}
	committed, err := os.ReadFile("../schema/protocol.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	written, _ := os.ReadFile(out)
	if strings.ReplaceAll(string(committed), "\r\n", "\n") != string(written) {
		t.Error("the command does not write what is committed")
	}
}

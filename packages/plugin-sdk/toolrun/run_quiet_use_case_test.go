package toolrun

import (
	"bytes"
	"strings"
	"testing"
)

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	previous := Log
	Log = &buf
	t.Cleanup(func() { Log = previous })
	return &buf
}

func TestQuietHidesWhatAToolPrintsWhenItSucceeds(t *testing.T) {
	t.Setenv(VerboseEnv, "")
	log := captureLog(t)
	if err := Quiet(t.TempDir(), []string{"go", "version"}); err != nil {
		t.Fatal(err)
	}
	if text := log.String(); !strings.Contains(text, "Running go version") || strings.Contains(text, "go version go") {
		t.Fatalf("%q", text)
	}
}

func TestQuietShowsWhatAToolPrintedWhenItFails(t *testing.T) {
	t.Setenv(VerboseEnv, "")
	log := captureLog(t)
	if err := Quiet(t.TempDir(), []string{"go", "definitely-not-a-command"}); err == nil {
		t.Fatal("a failing command must fail")
	}
	if !strings.Contains(log.String(), "definitely-not-a-command") {
		t.Fatalf("the tool's output is the reason, and was hidden: %q", log.String())
	}
}

func TestQuietStreamsWhenVerbose(t *testing.T) {
	t.Setenv(VerboseEnv, "1")
	log := captureLog(t)
	if err := Quiet(t.TempDir(), []string{"go", "version"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(log.String(), "go version go") {
		t.Fatalf("%q", log.String())
	}
}

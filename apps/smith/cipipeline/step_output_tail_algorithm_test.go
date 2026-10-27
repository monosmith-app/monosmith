package cipipeline

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func numbered(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

func TestTheTailWriterKeepsTheEndOfALongOutput(t *testing.T) {
	w := &tailWriter{}
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(w, "line %d\n", i)
	}
	got := lastLines(w.String(), 2)
	if got != "line 19998\nline 19999" {
		t.Errorf("the end survives trimming: %q", got)
	}
	if len(w.buf) > 2*tailKeep {
		t.Errorf("memory is bounded: %d bytes", len(w.buf))
	}
}

func TestAFailedStepShowsItsOutputTailAndTheLogFileItNames(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	log := filepath.Join(t.TempDir(), "debug-0.log")
	if err := os.WriteFile(log, []byte(numbered(100)), 0o644); err != nil {
		t.Fatal(err)
	}
	output := numbered(50) + "npm error A complete log of this run can be found in: " + log + "\n"

	var out bytes.Buffer
	msg := failureWithOutput(&out, "Install (exit 1)", output)

	if !strings.HasPrefix(msg, "Install (exit 1)\n") || !strings.Contains(msg, "complete log of this run") || strings.Contains(msg, "line 20\n") {
		t.Errorf("the failure carries the last %d lines of the output: %q", outputTailLines, msg)
	}
	shown := out.String()
	if !strings.HasPrefix(shown, "::group::log file "+log+"\n") || !strings.Contains(shown, "line 100\n") || strings.Contains(shown, "line 60\n") || !strings.HasSuffix(shown, "::endgroup::\n") {
		t.Errorf("the log file's last %d lines sit in a group: %q", logFileLines, shown)
	}
}

func TestAFailureWithoutAReadableLogFileIsJustTheTail(t *testing.T) {
	var out bytes.Buffer
	missing := "npm error A complete log of this run can be found in: " + filepath.Join(t.TempDir(), "gone.log") + "\n"
	if msg := failureWithOutput(&out, "Install (exit 1)", missing); !strings.Contains(msg, "gone.log") || out.Len() != 0 {
		t.Errorf("an unreadable file prints nothing: %q %q", msg, out.String())
	}
	if msg := failureWithOutput(&out, "Install (exit 1)", ""); msg != "Install (exit 1)" {
		t.Errorf("no output, no tail: %q", msg)
	}
}

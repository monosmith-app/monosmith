package cipipeline

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
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
	msg := failureWithOutput(&out, "Install", "Install (exit 1)", output)

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
	if msg := failureWithOutput(&out, "Install", "Install (exit 1)", missing); !strings.Contains(msg, "gone.log") || strings.Contains(out.String(), "log file") {
		t.Errorf("an unreadable file prints no log-file group: %q %q", msg, out.String())
	}
	if msg := failureWithOutput(&out, "Install", "Install (exit 1)", ""); msg != "Install (exit 1)" {
		t.Errorf("no output, no tail: %q", msg)
	}
}

func TestAFailedStepPrintsItsWholeOutputInAGroupNamedAfterIt(t *testing.T) {
	cases := []struct{ name, env, open, shut string }{
		{"github", "GITHUB_ACTIONS", "::group::Install tool: full output\n", "::endgroup::\n"},
		{"azure", "TF_BUILD", "##[group]Install tool: full output\n", "##[endgroup]\n"},
		{"plain", "", "Install tool: full output\n", ""},
	}
	whole := numbered(100)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("GITHUB_ACTIONS", "")
			t.Setenv("TF_BUILD", "")
			if c.env != "" {
				t.Setenv(c.env, "true")
			}
			var out bytes.Buffer
			msg := failureWithOutput(&out, "Install tool", "Install tool (exit 1)", whole)
			shown := out.String()
			if !strings.HasPrefix(shown, c.open) || !strings.HasSuffix(shown, c.shut) || !strings.Contains(shown, "line 1\n") || !strings.Contains(shown, "line 100\n") {
				t.Errorf("the whole output sits in a group named after the step: %q", shown)
			}
			if strings.Contains(msg, "line 20\n") || !strings.Contains(msg, "line 100") {
				t.Errorf("the error keeps only the last %d lines: %q", outputTailLines, msg)
			}
		})
	}
}

func TestABlockingStepThatFailsReportsItsOutputInAGroup(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	steps := []protocol.Step{{Name: "Install tool", Command: []string{"go", "tool", "no-such-tool-for-smith"}, Blocking: true}}
	var out bytes.Buffer
	_, err := runSteps(context.Background(), t.TempDir(), steps, nil, &out)
	if err == nil || !strings.Contains(err.Error(), "Install tool (exit") {
		t.Fatalf("a failed blocking step stops the phase: %v", err)
	}
	shown := out.String()
	if !strings.Contains(shown, "::group::Install tool: full output\n") || !strings.Contains(shown, "no-such-tool-for-smith") || !strings.HasSuffix(shown, "::endgroup::\n") {
		t.Errorf("the output of the failed step sits in a group: %q", shown)
	}
}

func TestTheCapturedOutputIsBoundedToTheTailBuffer(t *testing.T) {
	w := &tailWriter{}
	_, _ = w.Write([]byte(strings.Repeat("x", tailKeep+10)))
	if len(w.String()) != tailKeep {
		t.Errorf("captured %d bytes", len(w.String()))
	}
}

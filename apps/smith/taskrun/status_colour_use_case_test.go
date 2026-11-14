package taskrun

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

func TestStatusColourFollowsTheTerminalDecisionAndNeverTheToolColourAHostHandsOn(t *testing.T) {
	env := func(kv ...string) func(string) string {
		m := map[string]string{}
		for i := 0; i+1 < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return func(k string) string { return m[k] }
	}
	terminal := func(int) bool { return true }
	cases := []struct {
		name     string
		getenv   func(string) string
		asJSON   bool
		events   bool
		out      *os.File
		fileTerm func(int) bool
		want     bool
	}{
		{"a terminal", env(), false, false, os.Stdout, terminal, true},
		{"NO_COLOR", env("NO_COLOR", "1"), false, false, os.Stdout, terminal, false},
		{"a pipe", env(), false, false, nil, terminal, false},
		{"FORCE_COLOR on a pipe", env("FORCE_COLOR", "1"), false, false, nil, terminal, true},
		{"the interface child: tool colour on, FORCE_COLOR=0, a pipe", env(termcolour.ToolColourEnv, "1", "FORCE_COLOR", "0"), false, false, nil, terminal, false},
		{"--json", env(), true, false, os.Stdout, terminal, false},
		{"--json-events", env("FORCE_COLOR", "1"), false, true, nil, terminal, false},
	}
	for _, c := range cases {
		var out interface {
			Write([]byte) (int, error)
		} = &bytes.Buffer{}
		if c.out != nil {
			out = c.out
		}
		if got := statusColour(c.getenv, c.asJSON, c.events, out, c.fileTerm); got != c.want {
			t.Errorf("%s: statusColour = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestARunPaintsItsOwnLinesButNotTheLinesOfATool(t *testing.T) {
	f := newFixture(t, "colour")
	g, err := taskgraph.Build(f.pg, []string{"core"}, []string{"build"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	Run(context.Background(), g, Options{Root: f.root, Files: f.files, Projects: f.pg, Storage: f.store, Out: &out, ToolVersion: "test", StatusColour: true})
	text := out.String()
	if !strings.Contains(text, termcolour.Wrap("▶", termcolour.Dim)) || !strings.Contains(text, termcolour.Wrap("✓", termcolour.Green)) {
		t.Errorf("the start and the end of the task are not painted:\n%q", text)
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, " │ ") && termcolour.HasSGR(line) {
			t.Errorf("a tool line was painted by the status paint: %q", line)
		}
	}
}

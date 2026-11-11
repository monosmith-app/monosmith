package taskrun

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// runColour runs core:build (the helper's "colour" mode) with colour on or
// off and returns the summary, the text and the task.output lines.
func (f *fixture) runColour(t *testing.T, colour bool) (Summary, string, []string) {
	t.Helper()
	g, err := taskgraph.Build(f.pg, []string{"core"}, []string{"build"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	got := &collected{}
	s := Run(context.Background(), g, Options{Root: f.root, Files: f.files, Projects: f.pg, Storage: f.store,
		Out: &out, ToolVersion: "test", Colour: colour, Events: runevents.NewRunStream("r1", got)})
	var lines []string
	for _, e := range got.events {
		if e.Event == runevents.TaskOutput {
			lines = append(lines, e.Line)
		}
	}
	return s, out.String(), lines
}

func TestColourAsksTheToolAndPaintsWhatItCannotColour(t *testing.T) {
	f := newFixture(t, "colour")
	s, out, events := f.runColour(t, true)
	if !s.OK() {
		t.Fatalf("%s\n%s", s, out)
	}
	for _, want := range []string{
		"core:build │ " + termcolour.Wrap("--- FAIL: TestX (0.00s)", termcolour.Red), // smith paint
		"core:build │ \x1b[31mred by the tool\x1b[0m",                                // the tool's own colour, cursor codes gone
		"core:build │ FORCE_COLOR=1 NO_COLOR=\n",                                     // the tool was asked for colour
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%q", want, out)
		}
	}
	if strings.Contains(out, "\x1b[2K") || strings.Contains(out, "\x1b[1G") {
		t.Errorf("a cursor code passed:\n%q", out)
	}
	for _, line := range events {
		if strings.Contains(line, "\x1b") {
			t.Errorf("an event line must stay plain: %q", line)
		}
	}
}

func TestPlainStaysPlainAndAsksTheToolForPlainText(t *testing.T) {
	f := newFixture(t, "colour")
	s, out, _ := f.runColour(t, false)
	if !s.OK() {
		t.Fatalf("%s\n%s", s, out)
	}
	if strings.Contains(out, "\x1b") {
		t.Errorf("a plain destination got an escape code:\n%q", out)
	}
	if !strings.Contains(out, "core:build │ FORCE_COLOR=0 NO_COLOR=1\n") || !strings.Contains(out, "core:build │ red by the tool\n") {
		t.Errorf("plain run:\n%s", out)
	}
}

func TestTheCacheStoresThePlainLogAndAReplayIsPaintedAgain(t *testing.T) {
	f := newFixture(t, "colour")
	if s, out, _ := f.runColour(t, true); !s.OK() {
		t.Fatalf("%s\n%s", s, out)
	}
	_ = os.RemoveAll(filepath.Join(f.root, "libs", "core", "out"))
	s, plain, _ := f.runColour(t, false)
	if s.Count(Cached) != 1 {
		t.Fatalf("colour must not change the key: %s\n%s", s, plain)
	}
	if strings.Contains(plain, "\x1b") || !strings.Contains(plain, "core:build │ red by the tool\n") {
		t.Errorf("a replay to a plain destination must have no codes:\n%q", plain)
	}
	s, painted, _ := f.runColour(t, true)
	if s.Count(Cached) != 1 || !strings.Contains(painted, termcolour.Wrap("--- FAIL: TestX (0.00s)", termcolour.Red)) {
		t.Errorf("a replay to a colour destination gets smith paint:\n%q", painted)
	}
	entry, err := taskcache.Lookup(f.store, s.Results[0].Key)
	if err != nil {
		t.Fatal(err)
	}
	if log, _ := taskcache.Restore(f.store, f.root, "libs/core", []string{"out"}, entry); bytes.Contains(log, []byte("\x1b")) {
		t.Errorf("the cached log must be plain: %q", log)
	}
}

func TestTheColourVariablesNeverJoinTheCacheKey(t *testing.T) {
	f := newFixture(t, "core")
	g, _ := taskgraph.Build(f.pg, []string{"web"}, []string{"build"})
	plainKey, plainIn := KeyOf(g, Options{Root: f.root, Files: f.files, Projects: f.pg, ToolVersion: "test"}, "web:build")
	colourKey, colourIn := KeyOf(g, Options{Root: f.root, Files: f.files, Projects: f.pg, ToolVersion: "test", Colour: true}, "web:build")
	if plainKey != colourKey {
		t.Fatalf("the key changed with colour: %s vs %s", plainKey, colourKey)
	}
	r := &runner{opts: Options{Root: f.root, Colour: true}}
	env := strings.Join(r.env(g.Tasks["web:build"], colourKey), "\n")
	if !strings.Contains(env, "FORCE_COLOR=1") || !strings.Contains(env, termcolour.ToolEntry(true)) {
		t.Fatalf("the task env lacks the colour entries")
	}
	for k := range colourIn.Env {
		if strings.Contains(strings.Join(termcolour.Env(true), " "), k+"=") || k == termcolour.ToolColourEnv {
			t.Errorf("%s joined KeyInputs.Env", k)
		}
	}
	if len(plainIn.Env) != len(colourIn.Env) {
		t.Errorf("KeyInputs.Env differs: %v vs %v", plainIn.Env, colourIn.Env)
	}
}

func TestToolColourOfAWriterThatIsNoTerminalFollowsOnlyTheEnvironment(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "")
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("TF_BUILD", "")
	t.Setenv("GITLAB_CI", "")
	t.Setenv(termcolour.ToolColourEnv, "")
	if ToolColour(&bytes.Buffer{}) {
		t.Error("a buffer is no terminal")
	}
	t.Setenv("GITHUB_ACTIONS", "true")
	if !ToolColour(&bytes.Buffer{}) {
		t.Error("GitHub Actions renders colour")
	}
	t.Setenv("NO_COLOR", "1")
	if ToolColour(&bytes.Buffer{}) {
		t.Error("NO_COLOR wins")
	}
}

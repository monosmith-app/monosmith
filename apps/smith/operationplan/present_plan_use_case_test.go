package operationplan

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func samplePlan() Plan {
	var p Plan
	p.Title = "add package billing"
	p.Add("Project", Line{Action: Create, Path: "packages/billing/", Note: "node plugin"})
	p.Add("Lint, added by eslint", Line{Action: Create, Path: "packages/billing/eslint.config.mjs"})
	p.Add("Project", Line{Action: Change, Path: "smith.toml"})
	return p
}

func TestRenderGroupsLinesByFeature(t *testing.T) {
	var b bytes.Buffer
	Render(&b, samplePlan())
	got := b.String()
	for _, want := range []string{"Plan: add package billing", "  Project\n", "create  packages/billing/  (node plugin)", "  Lint, added by eslint", "change  smith.toml"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Count(got, "  Project\n") != 1 {
		t.Fatalf("lines of one heading share one group:\n%s", got)
	}
}

func TestRenderEmptyPlanWritesNothing(t *testing.T) {
	var b bytes.Buffer
	Render(&b, Plan{Title: "nothing"})
	if b.Len() != 0 {
		t.Fatalf("got %q", b.String())
	}
}

func TestShouldAskOnlyForAPersonAtATerminal(t *testing.T) {
	cases := []struct {
		name string
		env  Environment
		yes  bool
		want bool
	}{
		{"terminal", Environment{Terminal: true}, false, true},
		{"terminal with --yes", Environment{Terminal: true}, true, false},
		{"no terminal", Environment{}, false, false},
		{"terminal in CI", Environment{Terminal: true, CI: true}, false, false},
	}
	for _, c := range cases {
		if got := ShouldAsk(c.env, c.yes); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestPresentNeverBlocksWithoutATerminal(t *testing.T) {
	var out bytes.Buffer
	// An input that would block forever proves the question is not read.
	ok, err := Present(&out, blockingReader{}, samplePlan(), Options{Ask: true, Env: Environment{CI: true}})
	if err != nil || !ok {
		t.Fatalf("got %v, %v", ok, err)
	}
	if strings.Contains(out.String(), "Go ahead?") {
		t.Fatal("asked in CI")
	}
}

func TestPresentAsksInATerminal(t *testing.T) {
	env := Environment{Terminal: true}
	var out bytes.Buffer
	ok, err := Present(&out, strings.NewReader("y\n"), samplePlan(), Options{Ask: true, Env: env})
	if err != nil || !ok || !strings.Contains(out.String(), "Go ahead?") {
		t.Fatalf("yes: %v %v\n%s", ok, err, out.String())
	}
	ok, err = Present(&out, strings.NewReader("n\n"), samplePlan(), Options{Ask: true, Env: env})
	if ok || !errors.Is(err, ErrDeclined) {
		t.Fatalf("no: %v %v", ok, err)
	}
	ok, err = Present(&out, blockingReader{}, samplePlan(), Options{Ask: true, Yes: true, Env: env})
	if !ok || err != nil {
		t.Fatalf("--yes: %v %v", ok, err)
	}
}

func TestPresentDryRunStops(t *testing.T) {
	var out bytes.Buffer
	ok, err := Present(&out, blockingReader{}, samplePlan(), Options{DryRun: true, Ask: true, Env: Environment{Terminal: true}})
	if ok || err != nil || !strings.Contains(out.String(), "nothing was changed") {
		t.Fatalf("%v %v\n%s", ok, err, out.String())
	}
}

func TestWithOwnersReadsTheRecord(t *testing.T) {
	root := t.TempDir()
	lock := "schema = 1\n\n[[file]]\npath = \"tsconfig.json\"\nowner = \"you\"\nrenderer = \"r\"\nversion = 1\nmode = \"file\"\nhash = \"h\"\nwanted = \"h\"\n"
	if err := os.WriteFile(filepath.Join(root, "smith.lock"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}
	var p Plan
	p.Add("g", Line{Action: Change, Path: "tsconfig.json"})
	p.Add("g", Line{Action: Change, Path: "other.json"})
	got := WithOwners(root, p).Groups[0].Lines
	if got[0].Owner != "you" || got[1].Owner != "" {
		t.Fatalf("%+v", got)
	}
}

type blockingReader struct{}

func (blockingReader) Read([]byte) (int, error) { select {} }

package derive

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
)

func TestReclaimGivesAnEjectedFileBackToSmithWithItsDiff(t *testing.T) {
	root := t.TempDir()
	rec := &derivelock.Record{}
	f := refs("../core")
	derive(t, root, rec, Decision{}, f)
	derive(t, root, rec, Decision{Eject: []string{f.Path}}, f)
	if e := rec.Find(f.Path, f.Renderer); e == nil || e.Owner != derivelock.OwnerYou {
		t.Fatalf("not ejected: %+v", e)
	}
	path := filepath.Join(root, filepath.FromSlash(f.Path))
	if err := os.WriteFile(path, []byte(`{"references":[{"path":"../mine"}],"extra":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := PlanFor(root, *rec, []Proposal{{Plugin: "node", File: f}})
	if err != nil {
		t.Fatal(err)
	}
	changes, err := PlanReclaim(root, plan, []string{f.Path})
	if err != nil {
		t.Fatal(err)
	}
	var shown bytes.Buffer
	PrintReclaimDiffs(&shown, []string{f.Path}, changes)
	if !strings.Contains(shown.String(), "will be overwritten") || !strings.Contains(shown.String(), "../mine") || !strings.Contains(shown.String(), "../core") {
		t.Fatalf("the diff shows both sides:\n%s", shown.String())
	}
	if got := read(t, root, f.Path); !strings.Contains(got, "../mine") {
		t.Fatalf("planning must not write: %s", got)
	}

	out, err := Apply(root, plan, rec, Decision{Reclaim: []string{f.Path}})
	if err != nil || len(out.Reclaimed) != 1 {
		t.Fatalf("outcome %+v, error %v", out, err)
	}
	if got := read(t, root, f.Path); !strings.Contains(got, "../core") || strings.Contains(got, "../mine") || !strings.Contains(got, `"extra"`) {
		t.Errorf("derived keys are back and the file's other keys stay: %s", got)
	}
	if e := rec.Find(f.Path, f.Renderer); e == nil || e.Owner != derivelock.OwnerSmith {
		t.Errorf("smith owns it again: %+v", e)
	}
	plan, _ = derive(t, root, rec, Decision{}, f)
	if plan.Items[0].Status != StatusCurrent {
		t.Errorf("after the reclaim derive finds the file %s", plan.Items[0].Status)
	}
}

func TestReclaimTakesADriftedFileBackAndRefusesAFileSmithDoesNotDerive(t *testing.T) {
	root := t.TempDir()
	rec := &derivelock.Record{}
	f := refs("../core")
	derive(t, root, rec, Decision{}, f)
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(f.Path)), []byte(`{"references":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, out := derive(t, root, rec, Decision{Reclaim: []string{f.Path}}, f)
	if plan.Items[0].Status != StatusDrift || len(out.Reclaimed) != 1 || len(out.Held) != 0 {
		t.Fatalf("status %s, outcome %+v", plan.Items[0].Status, out)
	}
	if got := read(t, root, f.Path); !strings.Contains(got, "../core") {
		t.Errorf("the drift is gone: %s", got)
	}
	if _, err := Apply(root, plan, rec, Decision{Reclaim: []string{"other/tsconfig.json"}}); err == nil {
		t.Error("a file smith does not derive cannot be reclaimed")
	}
}

package derive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func refs(paths ...string) protocol.DerivedFile {
	var list []any
	for _, p := range paths {
		list = append(list, map[string]any{"path": p})
	}
	return protocol.DerivedFile{Project: "web", Path: "web/tsconfig.lib.json", Mode: protocol.DeriveKeys, Format: "json", Renderer: "node.tsconfig.references", Version: 1, Keys: map[string]any{"references": list}}
}

func derive(t *testing.T, root string, rec *derivelock.Record, d Decision, f protocol.DerivedFile) (Plan, Outcome) {
	t.Helper()
	plan, err := PlanFor(root, *rec, []Proposal{{Plugin: "node", File: f}})
	if err != nil {
		t.Fatal(err)
	}
	out, err := Apply(root, plan, rec, d)
	if err != nil {
		t.Fatal(err)
	}
	return plan, out
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestOwnedFileFollowsItsInputsButNeverOverwritesAHandEdit(t *testing.T) {
	root := t.TempDir()
	rec := &derivelock.Record{}
	f := refs("../core")

	// First run creates the file and records it.
	plan, out := derive(t, root, rec, Decision{}, f)
	if plan.Items[0].Status != StatusMissing || len(out.Written) != 1 {
		t.Fatalf("first run: %+v", plan.Items[0])
	}
	if e := rec.Find(f.Path, f.Renderer); e == nil || e.Owner != derivelock.OwnerSmith {
		t.Fatalf("not recorded: %+v", rec)
	}

	// Same inputs: current, nothing to do.
	plan, _ = derive(t, root, rec, Decision{}, f)
	if plan.Items[0].Status != StatusCurrent || len(plan.Pending()) != 0 {
		t.Fatalf("second run: %v", plan.Items[0].Status)
	}

	// The graph changed: stale, rewritten.
	f2 := refs("../core", "../util")
	plan, out = derive(t, root, rec, Decision{}, f2)
	if plan.Items[0].Status != StatusStale || len(out.Written) != 1 || !strings.Contains(read(t, root, f.Path), "../util") {
		t.Fatalf("stale run: %v", plan.Items[0].Status)
	}

	// A hand edit is drift, whatever the inputs do.
	edited := strings.Replace(read(t, root, f.Path), "../util", "../hand", 1)
	if err := os.WriteFile(filepath.Join(root, "web", "tsconfig.lib.json"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	f3 := refs("../core", "../util", "../more")
	plan, out = derive(t, root, rec, Decision{}, f3)
	if plan.Items[0].Status != StatusDrift || len(out.Held) != 1 || read(t, root, f.Path) != edited {
		t.Fatalf("drift run: %v, file changed: %v", plan.Items[0].Status, read(t, root, f.Path) != edited)
	}

	// A hand edit that equals the render is no drift: the record is refreshed.
	same, err := PlanFor(root, *rec, []Proposal{{File: refs("../core", "../hand")}})
	if err != nil || same.Items[0].Status != StatusUnrecorded {
		t.Fatalf("a hand edit equal to the render is not drift: %v %v", same.Items[0].Status, err)
	}

	// --adopt keeps the edit as the baseline against the current inputs.
	_, out = derive(t, root, rec, Decision{Adopt: []string{f.Path}}, f3)
	if len(out.Adopted) != 1 {
		t.Fatalf("adopt: %+v", out)
	}
	plan, _ = derive(t, root, rec, Decision{}, f3)
	if plan.Items[0].Status != StatusCurrent || read(t, root, f.Path) != edited {
		t.Fatalf("an adopted edit stands while the inputs stay: %v", plan.Items[0].Status)
	}
	plan, _ = derive(t, root, rec, Decision{}, refs("../changed"))
	if plan.Items[0].Status != StatusDrift {
		t.Fatalf("an adopted edit meets changed inputs: %v", plan.Items[0].Status)
	}

	// --replace overwrites after the diff, --eject stops touching the file.
	_, out = derive(t, root, rec, Decision{Replace: []string{f.Path}}, refs("../changed"))
	if len(out.Replaced) != 1 || !strings.Contains(read(t, root, f.Path), "../changed") {
		t.Fatalf("replace: %+v", out)
	}
	_, out = derive(t, root, rec, Decision{Eject: []string{f.Path}}, refs("../changed"))
	if len(out.Ejected) != 1 {
		t.Fatalf("eject: %+v", out)
	}
	plan, _ = derive(t, root, rec, Decision{}, refs("../elsewhere"))
	if plan.Items[0].Status != StatusEjected || strings.Contains(read(t, root, f.Path), "elsewhere") {
		t.Fatalf("ejected file was touched: %v", plan.Items[0].Status)
	}
}

func TestAdoptNeedsAKeysFileThatStillHasItsKeys(t *testing.T) {
	root := t.TempDir()
	rec := &derivelock.Record{}
	f := refs("../core")
	f.Keys["composite"] = true
	derive(t, root, rec, Decision{}, f)
	if err := os.WriteFile(filepath.Join(root, "web", "tsconfig.lib.json"), []byte(`{"references": [{"path": "../hand"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanFor(root, *rec, []Proposal{{File: f}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(root, plan, rec, Decision{Adopt: []string{f.Path}}); err == nil || !strings.Contains(err.Error(), "lost a derived key") {
		t.Fatalf("adopt of a file that lost its key: %v", err)
	}
}

func TestAnExistingUnrecordedFileIsRecordedWhenEqualAndHeldWhenDifferent(t *testing.T) {
	root := t.TempDir()
	rec := &derivelock.Record{}
	f := refs("../core")
	if err := os.MkdirAll(filepath.Join(root, "web"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "web", "tsconfig.lib.json"), []byte(`{"references":[{"path":"../other"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, out := derive(t, root, rec, Decision{}, f)
	if plan.Items[0].Status != StatusDiverged || len(out.Held) != 1 || len(rec.Files) != 0 {
		t.Fatalf("diverged: %v", plan.Items[0].Status)
	}
	if err := os.WriteFile(filepath.Join(root, "web", "tsconfig.lib.json"), []byte(`{"extends":"x","references":[{"path":"../core"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, out = derive(t, root, rec, Decision{}, f)
	if plan.Items[0].Status != StatusUnrecorded || len(out.Recorded) != 1 || len(rec.Files) != 1 {
		t.Fatalf("unrecorded: %v", plan.Items[0].Status)
	}
}

func TestADecisionOnAFileSmithDoesNotDeriveIsAnError(t *testing.T) {
	rec := &derivelock.Record{}
	if _, err := Apply(t.TempDir(), Plan{}, rec, Decision{Eject: []string{"nope.json"}}); err == nil {
		t.Fatal("want an error")
	}
}

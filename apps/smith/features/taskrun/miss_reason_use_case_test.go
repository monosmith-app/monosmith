package taskrun

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/features/taskcache"
)

func TestACacheMissSaysWhyTheTaskRan(t *testing.T) {
	f := newFixture(t, "core")
	s, out := f.run(t)
	if !s.OK() {
		t.Fatalf("first run: %s\n%s", s, out)
	}
	for _, r := range s.Results {
		if r.Miss == nil || r.Miss.Reason != runevents.MissNoEntry {
			t.Fatalf("%s never ran before: %+v", r.Task, r.Miss)
		}
	}
	_ = os.WriteFile(filepath.Join(f.root, "libs", "core", "src.txt"), []byte("v2"), 0o644)
	s, _ = f.run(t)
	byTask := map[string]*runevents.Miss{}
	for _, r := range s.Results {
		byTask[r.Task] = r.Miss
	}
	if m := byTask["core:build"]; m == nil || m.Reason != runevents.MissInputs || m.Files != 1 || len(m.Changed) != 0 {
		t.Fatalf("core's source changed: %+v", m)
	}
	if m := byTask["web:build"]; m == nil || m.Reason != runevents.MissInputs || m.Files != 0 || len(m.Changed) == 0 || m.Changed[0] != "upstream" {
		t.Fatalf("web runs because core changed: %+v", m)
	}
	_ = os.RemoveAll(filepath.Join(f.root, "apps", "web", "out"))
	s, _ = f.run(t)
	for _, r := range s.Results {
		if r.Miss != nil {
			t.Fatalf("a hit has no miss: %s %+v", r.Task, r.Miss)
		}
	}
	if e := finishedEvent(Result{Task: "a:b", Status: Succeeded, Miss: &runevents.Miss{Reason: runevents.MissKey}}); e.Miss == nil || e.Miss.Reason != runevents.MissKey {
		t.Fatalf("the event carries the miss: %+v", e)
	}
}

func TestTheDifferencesGiveTheReason(t *testing.T) {
	got := missFrom([]taskcache.Difference{{What: "command"}, {What: "file", Name: "a"}, {What: "file", Name: "b"}, {What: "env", Name: "X"}, {What: "env", Name: "Y"}})
	want := &runevents.Miss{Reason: runevents.MissInputs, Files: 2, Changed: []string{"command", "env"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%+v", got)
	}
	if got := missFrom(nil); got.Reason != runevents.MissKey {
		t.Fatalf("no difference with another key: %+v", got)
	}
}

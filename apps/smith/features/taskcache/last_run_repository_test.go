package taskcache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLastRunFollowsTheLatestCachedRun(t *testing.T) {
	ws := t.TempDir()
	s, err := NewLocalStorage(filepath.Join(ws, ".cache"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := LastRun(s, "web:build"); !errors.Is(err, ErrNoLastRun) {
		t.Fatalf("a task never cached has no last run: %v", err)
	}
	_ = os.MkdirAll(filepath.Join(ws, "apps", "web"), 0o755)
	in := &KeyInputs{Task: "web:build", Files: []FileDigest{{"apps/web/a.ts", "1"}}}
	for _, key := range []string{"aaa111", "bbb222"} {
		if err := Save(s, ws, key, "web:build", "apps/web", nil, []byte(key), 0, in); err != nil {
			t.Fatal(err)
		}
	}
	e, key, err := LastRun(s, "web:build")
	if err != nil || key != "bbb222" || e.Inputs == nil || e.Inputs.Files[0].Path != "apps/web/a.ts" {
		t.Fatalf("%v %s %+v", err, key, e.Inputs)
	}

	// A hit on an older entry makes it the last run again.
	if err := Remember(s, "web:build", "aaa111"); err != nil {
		t.Fatal(err)
	}
	if _, key, _ := LastRun(s, "web:build"); key != "aaa111" {
		t.Fatalf("last %s", key)
	}

	// A pruned entry is no run to compare with.
	_ = s.Delete(AC, "aaa111")
	if _, _, err := LastRun(s, "web:build"); !errors.Is(err, ErrNoLastRun) {
		t.Fatalf("pruned: %v", err)
	}
}

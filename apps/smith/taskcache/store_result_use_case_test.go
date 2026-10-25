package taskcache

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSaveLookupRestore(t *testing.T) {
	ws := t.TempDir()
	s, _ := NewLocalStorage(filepath.Join(ws, ".smith", "cache"))
	dist := filepath.Join(ws, "apps", "web", "dist")
	_ = os.MkdirAll(filepath.Join(dist, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(dist, "main.js"), []byte("console.log(1)"), 0o644)
	_ = os.WriteFile(filepath.Join(dist, "sub", "x.js"), []byte("x"), 0o644)

	if err := Save(s, ws, "key1", "web:build", "apps/web", []string{"dist", "missing"}, []byte("built!\n"), time.Second, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Lookup(s, "nope"); !errors.Is(err, ErrMiss) {
		t.Fatalf("want miss, got %v", err)
	}
	e, err := Lookup(s, "key1")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Outputs) != 2 || e.Duration != time.Second {
		t.Fatalf("entry %+v", e)
	}

	_ = os.WriteFile(filepath.Join(dist, "stale.js"), []byte("stale"), 0o644)
	_ = os.Remove(filepath.Join(dist, "main.js"))
	log, err := Restore(s, ws, "apps/web", []string{"dist"}, e)
	if err != nil {
		t.Fatal(err)
	}
	if string(log) != "built!\n" {
		t.Fatalf("log %q", log)
	}
	body, _ := os.ReadFile(filepath.Join(dist, "main.js"))
	if string(body) != "console.log(1)" {
		t.Fatalf("main.js %q", body)
	}
	if _, err := os.Stat(filepath.Join(dist, "stale.js")); err == nil {
		t.Fatal("restore must replace the output directory, not merge into it")
	}

	_ = os.WriteFile(filepath.Join(ws, "apps", "web", "file-output"), nil, 0o644)
	if err := Save(s, ws, "key2", "web:x", "apps/web", []string{"file-output"}, nil, 0, nil); err == nil || !strings.Contains(err.Error(), "directory") {
		t.Fatalf("a file output must be rejected, got %v", err)
	}
}

func TestPruneKeepsRecentEntries(t *testing.T) {
	ws := t.TempDir()
	s, _ := NewLocalStorage(filepath.Join(ws, "cache"))
	for i, key := range []string{"aaa", "bbb", "ccc"} {
		dir := filepath.Join(ws, key)
		_ = os.MkdirAll(dir, 0o755)
		_ = os.WriteFile(filepath.Join(dir, "f"), []byte(strings.Repeat(key, 100)), 0o644)
		if err := Save(s, ws, key, key, key, []string{"."}, []byte(key), 0, nil); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(time.Duration(i-10) * time.Hour)
		p, _ := s.path(AC, key)
		_ = os.Chtimes(p, old, old)
	}
	res, err := Prune(s, 1200) // the entries and last-run pointers count, not only the blobs
	if err != nil {
		t.Fatal(err)
	}
	if res.Entries != 1 {
		t.Fatalf("result %+v", res)
	}
	if _, err := Lookup(s, "aaa"); !errors.Is(err, ErrMiss) {
		t.Fatal("the oldest entry must go")
	}
	for _, k := range []string{"bbb", "ccc"} {
		if _, err := Lookup(s, k); err != nil {
			t.Fatalf("%s must survive: %v", k, err)
		}
	}
}

func TestKeyIsCanonical(t *testing.T) {
	a := Key(KeyInputs{Task: "t", Command: []string{"a", "b"}, Env: map[string]string{"X": "1", "Y": "2"}, Files: []FileDigest{{"b", "2"}, {"a", "1"}}, Deps: map[string]string{"p:x": "z", "p:y": "y"}})
	b := Key(KeyInputs{Task: "t", Command: []string{"a", "b"}, Env: map[string]string{"Y": "2", "X": "1"}, Files: []FileDigest{{"a", "1"}, {"b", "2"}}, Deps: map[string]string{"p:y": "y", "p:x": "z"}})
	if a != b {
		t.Fatal("ordering must not change the key")
	}
	c := Key(KeyInputs{Task: "t", Command: []string{"ab"}, Env: map[string]string{"X": "1", "Y": "2"}, Files: []FileDigest{{"a", "1"}, {"b", "2"}}, Deps: map[string]string{"p:x": "y", "p:y": "z"}})
	if a == c {
		t.Fatal("argv boundaries must be part of the key")
	}
}

func TestSelectInputs(t *testing.T) {
	files := []string{"package-lock.json", "README.md", "apps/web/src/a.ts", "apps/web/dist/a.js", "apps/web/src/a.spec.ts", "apps/webx/b.ts"}
	got := SelectInputs(files, "apps/web", []string{"**/*", "!**/*.spec.ts", "{workspaceRoot}/package-lock.json"}, []string{"dist"})
	if !reflect.DeepEqual(got, []string{"package-lock.json", "apps/web/src/a.ts"}) {
		t.Fatalf("got %v", got)
	}
	if !reflect.DeepEqual(ProjectFiles(files, "apps/web"), []string{"apps/web/src/a.ts", "apps/web/dist/a.js", "apps/web/src/a.spec.ts"}) {
		t.Fatal("project files")
	}
}

// An output nested past Windows' 260-character MAX_PATH saves and restores:
// smith works with absolute paths, which Go extends with \?\ on Windows.
func TestDeepOutputsRoundTrip(t *testing.T) {
	ws := t.TempDir()
	s, _ := NewLocalStorage(filepath.Join(ws, ".smith", "cache"))
	deep := filepath.Join(ws, "apps", "web", "dist")
	for i := 0; i < 12; i++ {
		deep = filepath.Join(deep, "a-rather-long-directory-name")
	}
	if len(deep) < 300 {
		t.Fatalf("path is only %d characters", len(deep))
	}
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deep, "chunk.js"), []byte("deep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(s, ws, "deepkey", "web:build", "apps/web", []string{"dist"}, nil, 0, nil); err != nil {
		t.Fatal(err)
	}
	e, err := Lookup(s, "deepkey")
	if err != nil || len(e.Outputs) != 1 || strings.Contains(e.Outputs[0].Path, `\`) {
		t.Fatalf("entry %+v %v; output paths are POSIX", e.Outputs, err)
	}
	if err := os.RemoveAll(filepath.Join(ws, "apps", "web", "dist")); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(s, ws, "apps/web", []string{"dist"}, e); err != nil {
		t.Fatal(err)
	}
	if body, err := os.ReadFile(filepath.Join(deep, "chunk.js")); err != nil || string(body) != "deep" {
		t.Fatalf("restored %q %v", body, err)
	}
}

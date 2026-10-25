package cachecommand

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// workspaceWithCache makes a workspace whose cache holds `tasks` results of
// `size` bytes each, and moves into it. It uses no session, so no plugin
// starts and SMITH_WORKSPACE_ROOT cannot send the test elsewhere.
func workspaceWithCache(t *testing.T, config string, tasks, size int) (root string, store *taskcache.LocalStorage) {
	t.Helper()
	root = t.TempDir()
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	t.Setenv("SMITH_CACHE_DIR", "")
	t.Setenv("SMITH_CACHE_REMOTE", "")
	t.Setenv("SMITH_CACHE_KEY", "")
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n\n"+config), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	store, err := taskcache.NewLocalStorage(filepath.Join(root, ".smith", "cache"))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < tasks; i++ {
		project := filepath.Join(root, "p", string(rune('a'+i)))
		out := filepath.Join(project, "dist")
		if err := os.MkdirAll(out, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "bundle.js"), bytes.Repeat([]byte{byte('a' + i)}, size), 0o644); err != nil {
			t.Fatal(err)
		}
		key := strings.Repeat(string(rune('a'+i)), 64)
		task := string(rune('a'+i)) + ":build"
		if err := taskcache.Save(store, root, key, task, "p/"+string(rune('a'+i)), []string{"dist"}, []byte("built "+task), time.Second, nil); err != nil {
			t.Fatal(err)
		}
		past := time.Now().Add(-time.Duration(tasks-i) * time.Hour)
		for _, kind := range []taskcache.Kind{taskcache.AC} {
			p := filepath.Join(store.Root(), string(kind), key[:2], key)
			_ = os.Chtimes(p, past, past)
		}
	}
	return root, store
}

func run(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	cmd := Command()
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), err
}

func TestStatusCommandPrintsTheCacheAndItsJSON(t *testing.T) {
	root, _ := workspaceWithCache(t, "[cache]\nmax_size_mb = 1\n", 3, 400<<10)
	out, err := run(t, "status")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, ".smith", "cache")
	for _, want := range []string{"local cache  " + dir, "  entries     3", "of 1.0 MB", "  last prune  never", "  remote      none", "✖ the cache is over its limit", "largest entries", "c:build"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	js, err := run(t, "status", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal([]byte(js), &report); err != nil {
		t.Fatalf("%v\n%s", err, js)
	}
	if report["entries"] != float64(3) || report["limitBytes"] != float64(1<<20) || report["usedPercent"].(float64) < 100 || report["remote"].(map[string]any)["configured"] != false {
		t.Fatalf("report %v", report)
	}
}

func TestStatusWithARemoteShowsSchemeAndHostNeverCredentials(t *testing.T) {
	workspaceWithCache(t, "", 1, 10)
	t.Setenv("SMITH_CACHE_REMOTE", "https://user:s3cret@cache.example.com/team?token=abc")
	out, err := run(t, "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "https://cache.example.com (off: set SMITH_CACHE_KEY to use it)") || strings.Contains(out, "s3cret") || strings.Contains(out, "token") {
		t.Fatalf("output:\n%s", out)
	}
	t.Setenv("SMITH_CACHE_KEY", "team-secret")
	t.Setenv("SMITH_CACHE_READ_ONLY", "1")
	out, _ = run(t, "status")
	if !strings.Contains(out, "(active, read-only)") || strings.Contains(out, "team-secret") {
		t.Fatalf("output:\n%s", out)
	}
}

func TestTheCacheDirectoryCanBeOverriddenByTheEnvironment(t *testing.T) {
	_, _ = workspaceWithCache(t, "", 1, 10)
	elsewhere := filepath.Join(t.TempDir(), "other cache")
	t.Setenv("SMITH_CACHE_DIR", elsewhere)
	out, err := run(t, "status")
	if err != nil || !strings.Contains(out, elsewhere) || !strings.Contains(out, "none yet") {
		t.Fatalf("err %v, output:\n%s", err, out)
	}
	if _, err := os.Stat(elsewhere); err == nil {
		t.Fatal("status must not create the directory")
	}
}

func TestOutsideAWorkspaceTheCommandsExplainWhatToDo(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	t.Chdir(t.TempDir())
	for _, sub := range []string{"status", "prune", "verify"} {
		_, err := run(t, sub)
		var p *problem.Error
		if !errors.As(err, &p) || len(p.Remedy) == 0 {
			t.Errorf("%s: want a problem with a remedy, got %v", sub, err)
		}
	}
}

func TestPruneCommandDryRunThenForReal(t *testing.T) {
	_, store := workspaceWithCache(t, "[cache]\nmax_size_mb = 1\n", 4, 400<<10)
	dry, err := run(t, "prune", "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(dry, "Would remove 2 entries and 4 blobs") || !strings.Contains(dry, "would be left in 2 entries") {
		t.Fatalf("dry run:\n%s", dry)
	}
	if status, _ := cachemaintenance.Inspect(filepath.Dir(store.Root()), 0); status.Entries != 4 {
		t.Fatal("a dry run removes nothing")
	}
	real, err := run(t, "prune")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(real, "Pruned the cache: removed 2 entries and 4 blobs") {
		t.Fatalf("prune:\n%s", real)
	}
	status, _ := cachemaintenance.Inspect(filepath.Dir(store.Root()), 1<<20)
	if status.Entries != 2 || status.OverLimit() || status.LastPrune == nil {
		t.Fatalf("after the prune: %+v", status)
	}
	again, _ := run(t, "prune")
	if !strings.HasPrefix(again, "Nothing to prune") {
		t.Fatalf("again:\n%s", again)
	}
}

func TestPruneCommandTakesALimitOfItsOwn(t *testing.T) {
	_, store := workspaceWithCache(t, "", 4, 400<<10)
	out, err := run(t, "prune", "--max-size-mb", "1", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var outcome cachemaintenance.PruneOutcome
	if err := json.Unmarshal([]byte(out), &outcome); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if outcome.LimitBytes != 1<<20 || outcome.Entries != 2 || outcome.DryRun {
		t.Fatalf("outcome %+v", outcome)
	}
	if status, _ := cachemaintenance.Inspect(filepath.Dir(store.Root()), 0); status.Entries != 2 {
		t.Fatalf("status %+v", status)
	}
	_, err = run(t, "prune", "--max-size-mb", "-3")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "not a size") {
		t.Fatalf("a negative size: %v", err)
	}
}

func TestClearRefusesWithoutAnswerAndDeletesWithYes(t *testing.T) {
	_, store := workspaceWithCache(t, "", 2, 100)
	dir := filepath.Dir(store.Root())
	_, err := run(t, "clear")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "without a terminal") || !strings.Contains(strings.Join(p.Remedy, " "), "smith cache clear --yes") {
		t.Fatalf("without a terminal: %v", err)
	}
	if status, _ := cachemaintenance.Inspect(dir, 0); status.Entries != 2 {
		t.Fatal("a refused clear deletes nothing")
	}
	out, err := run(t, "clear", "--yes")
	if err != nil || !strings.HasPrefix(out, "Deleted the local cache at "+dir+": 2 entries and") {
		t.Fatalf("err %v, output:\n%s", err, out)
	}
	if status, _ := cachemaintenance.Inspect(dir, 0); status.Exists {
		t.Fatal("the cache must be gone")
	}
	out, err = run(t, "clear")
	if err != nil || !strings.Contains(out, "already empty") {
		t.Fatalf("clearing nothing: %v\n%s", err, out)
	}
}

func TestClearNeverTouchesTheRemote(t *testing.T) {
	workspaceWithCache(t, "", 1, 10)
	// A remote that would fail loudly if anything tried to reach it.
	t.Setenv("SMITH_CACHE_REMOTE", "http://127.0.0.1:1/never")
	t.Setenv("SMITH_CACHE_KEY", "k")
	if _, err := run(t, "clear", "--yes"); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyCommandReportsDamageAndFixesIt(t *testing.T) {
	_, store := workspaceWithCache(t, "", 3, 100)
	dir := filepath.Dir(store.Root())
	clean, err := run(t, "verify")
	if err != nil || !strings.Contains(clean, "no problems") {
		t.Fatalf("a healthy cache: %v\n%s", err, clean)
	}
	// Corrupt one output blob.
	var victim string
	_ = store.Walk(taskcache.CAS, func(key string, size int64, _ time.Time) error {
		if size == 100 && victim == "" {
			victim = key
		}
		return nil
	})
	if err := os.WriteFile(filepath.Join(store.Root(), "cas", victim[:2], victim), []byte("rotten"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "verify")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Error(), "→ run `smith cache verify --fix`") {
		t.Fatalf("damage must fail with a remedy, got %v", err)
	}
	if !strings.Contains(out, "corrupt-blob") || !strings.Contains(out, "damaged-entry") {
		t.Fatalf("output:\n%s", out)
	}
	js, err := run(t, "verify", "--json")
	if err == nil {
		t.Fatal("the JSON form fails too while problems remain")
	}
	var report cachemaintenance.VerifyReport
	if e := json.Unmarshal([]byte(js), &report); e != nil || len(report.Problems) != 2 {
		t.Fatalf("report %+v (%v)\n%s", report, e, js)
	}

	fixed, err := run(t, "verify", "--fix")
	if err != nil || strings.Count(fixed, "(removed)") != 2 {
		t.Fatalf("fix: %v\n%s", err, fixed)
	}
	if status, _ := cachemaintenance.Inspect(dir, 0); status.Entries != 2 {
		t.Fatalf("status %+v", status)
	}
	if _, err := run(t, "verify"); err != nil {
		t.Fatalf("after the fix: %v", err)
	}
}

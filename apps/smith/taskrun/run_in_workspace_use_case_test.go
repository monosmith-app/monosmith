package taskrun

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// cachedSession is a session on a temp workspace, with no plugins and no git:
// pruneAfterRun needs only the root and the cache settings.
func cachedSession(t *testing.T, maxSizeMB int) (*session.Session, *taskcache.LocalStorage) {
	t.Helper()
	t.Setenv("SMITH_CACHE_DIR", "")
	root := t.TempDir()
	var cfg workspace.Config
	cfg.Cache.Dir = ".smith/cache"
	cfg.Cache.MaxSizeMB = maxSizeMB
	store, err := taskcache.NewLocalStorage(filepath.Join(root, ".smith", "cache"))
	if err != nil {
		t.Fatal(err)
	}
	return &session.Session{Root: root, Config: cfg}, store
}

// store puts one blob through the meter, as a task that finished does.
func store(t *testing.T, meter *cachemaintenance.WriteMeter, content []byte) string {
	t.Helper()
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])
	if err := meter.Put(taskcache.CAS, digest, bytes.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	return digest
}

func TestARunThatStoredNothingDoesNotTouchTheCacheDirectory(t *testing.T) {
	s, local := cachedSession(t, 1)
	meter := cachemaintenance.NewWriteMeter(local)
	var warnings bytes.Buffer
	pruneAfterRun(s, meter, &warnings)
	if _, err := os.Stat(filepath.Join(local.Root(), "prune-state.json")); err == nil {
		t.Fatal("a run that only replayed results has nothing to prune")
	}
	if warnings.Len() != 0 {
		t.Fatalf("warnings: %s", warnings.String())
	}
}

func TestARunThatStoredResultsPrunesTheLocalStoreToTheLimit(t *testing.T) {
	s, local := cachedSession(t, 1)
	// Old results from earlier runs fill the store past 1 MB.
	for i := 0; i < 3; i++ {
		content := bytes.Repeat([]byte{byte(i + 1)}, 600<<10)
		sum := sha256.Sum256(content)
		digest := hex.EncodeToString(sum[:])
		if err := local.Put(taskcache.CAS, digest, bytes.NewReader(content)); err != nil {
			t.Fatal(err)
		}
		long := time.Now().Add(-3 * time.Hour)
		_ = os.Chtimes(filepath.Join(local.Root(), "cas", digest[:2], digest), long, long)
	}
	meter := cachemaintenance.NewWriteMeter(local)
	store(t, meter, []byte("what this run stored"))
	var warnings bytes.Buffer
	pruneAfterRun(s, meter, &warnings)
	if warnings.Len() != 0 {
		t.Fatalf("warnings: %s", warnings.String())
	}
	status, err := cachemaintenance.Inspect(filepath.Join(s.Root, ".smith", "cache"), 1<<20)
	if err != nil || status.LastPrune == nil {
		t.Fatalf("status %+v (%v)", status, err)
	}
	if status.Blobs > 1 {
		t.Fatalf("the unused old blobs should have been swept: %+v", status)
	}
}

func TestAPruneFailureIsAWarningWithARemedyAndNeverAFailedRun(t *testing.T) {
	s, local := cachedSession(t, 1)
	meter := cachemaintenance.NewWriteMeter(local)
	store(t, meter, []byte("stored"))
	// Break the store so that the record of the prune cannot be written.
	if err := os.RemoveAll(filepath.Join(local.Root(), "tmp")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local.Root(), "tmp"), []byte("in the way"), 0o644); err != nil {
		t.Fatal(err)
	}
	var warnings bytes.Buffer
	pruneAfterRun(s, meter, &warnings) // returns nothing: it cannot fail the run
	got := warnings.String()
	if !strings.HasPrefix(got, "smith: warning: could not prune the cache to its limit") || !strings.Contains(got, "→ run `smith cache prune`") {
		t.Fatalf("warning:\n%s", got)
	}
}

func TestWithNoLimitTheRunNeverPrunes(t *testing.T) {
	for _, mb := range []int{0, -1} {
		s, local := cachedSession(t, mb)
		meter := cachemaintenance.NewWriteMeter(local)
		store(t, meter, bytes.Repeat([]byte("x"), 2<<20))
		var warnings bytes.Buffer
		pruneAfterRun(s, meter, &warnings)
		if _, err := os.Stat(filepath.Join(local.Root(), "prune-state.json")); err == nil || warnings.Len() != 0 {
			t.Fatalf("max_size_mb %d: a cache with no limit is left alone (warnings %q)", mb, warnings.String())
		}
	}
}

package cachemaintenance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// seedEntry stores one cached result the way taskcache.Save does: the output
// blobs and the log first, then the entry that names them. It returns the
// entry's key and the digests of its blobs.
func seedEntry(t testing.TB, store taskcache.Storage, task string, outputs ...[]byte) (key string, blobs []string) {
	t.Helper()
	entry := taskcache.Entry{Schema: taskcache.EntrySchema, Task: task, Created: time.Now().UTC()}
	for i, content := range outputs {
		digest := putBlob(t, store, content)
		blobs = append(blobs, digest)
		entry.Outputs = append(entry.Outputs, taskcache.OutputFile{Path: fmt.Sprintf("dist/%d", i), Digest: digest, Mode: 0o644})
	}
	entry.Log = putBlob(t, store, []byte("log of "+task))
	blobs = append(blobs, entry.Log)
	raw, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(task + string(bytes.Join(outputs, nil))))
	key = hex.EncodeToString(sum[:])
	if err := store.Put(taskcache.AC, key, bytes.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	if err := taskcache.Remember(store, task, key); err != nil {
		t.Fatal(err)
	}
	return key, blobs
}

func putBlob(t testing.TB, store taskcache.Storage, content []byte) string {
	t.Helper()
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])
	if err := store.Put(taskcache.CAS, digest, bytes.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	return digest
}

// age sets when a stored file was last used.
func age(t testing.TB, store *taskcache.LocalStorage, kind taskcache.Kind, key string, ago time.Duration) {
	t.Helper()
	when := time.Now().Add(-ago)
	if err := os.Chtimes(filepath.Join(store.Root(), string(kind), key[:2], key), when, when); err != nil {
		t.Fatal(err)
	}
}

func newStoreAt(dir string) (*taskcache.LocalStorage, error) { return taskcache.NewLocalStorage(dir) }

func newStore(t testing.TB) (*taskcache.LocalStorage, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "cache")
	store, err := taskcache.NewLocalStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	return store, dir
}

func TestInspectCountsEntriesBlobsAndTheLargestTasks(t *testing.T) {
	store, dir := newStore(t)
	seedEntry(t, store, "web:build", bytes.Repeat([]byte("a"), 3000))
	seedEntry(t, store, "web:build", bytes.Repeat([]byte("b"), 1000))
	seedEntry(t, store, "api:test", bytes.Repeat([]byte("c"), 500))

	status, err := Inspect(dir, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Exists || status.Entries != 3 || status.Blobs != 5 || status.BrokenEntries != 0 {
		t.Fatalf("status %+v", status)
	}
	if status.Bytes < 3000+1000+500 || status.OverLimit() {
		t.Fatalf("bytes %d over %v", status.Bytes, status.OverLimit())
	}
	if len(status.Largest) != 2 || status.Largest[0].Task != "web:build" || status.Largest[0].Entries != 2 || status.Largest[0].LargestBytes < 3000 {
		t.Fatalf("largest %+v", status.Largest)
	}
	if status.LastPrune != nil {
		t.Fatalf("a cache nobody pruned has no last prune: %v", status.LastPrune)
	}
}

func TestInspectOfACacheThatDoesNotExistIsEmptyAndCreatesNothing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "never")
	status, err := Inspect(dir, 100)
	if err != nil || status.Exists || status.Entries != 0 || len(status.Largest) != 0 {
		t.Fatalf("status %+v, err %v", status, err)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("looking at a cache must not create it")
	}
}

func TestInspectListsOnlyTheFiveLargestTasks(t *testing.T) {
	store, dir := newStore(t)
	for i := 1; i <= 8; i++ {
		seedEntry(t, store, fmt.Sprintf("p%d:build", i), bytes.Repeat([]byte{byte('a' + i)}, i*100))
	}
	status, _ := Inspect(dir, 0)
	if len(status.Largest) != 5 || status.Largest[0].Task != "p8:build" || status.Largest[4].Task != "p4:build" {
		t.Fatalf("largest %+v", status.Largest)
	}
}

func TestInspectCountsEntriesALookupWouldMiss(t *testing.T) {
	store, dir := newStore(t)
	_, blobs := seedEntry(t, store, "web:build", []byte("out"))
	if err := store.Delete(taskcache.CAS, blobs[0]); err != nil {
		t.Fatal(err)
	}
	garbage := sha256.Sum256([]byte("x"))
	if err := store.Put(taskcache.AC, hex.EncodeToString(garbage[:]), bytes.NewReader([]byte("not json"))); err != nil {
		t.Fatal(err)
	}
	status, _ := Inspect(dir, 0)
	if status.Entries != 2 || status.BrokenEntries != 2 {
		t.Fatalf("status %+v", status)
	}
}

func TestInspectReportsOverTheLimitAndTheLastPrune(t *testing.T) {
	store, dir := newStore(t)
	seedEntry(t, store, "web:build", bytes.Repeat([]byte("a"), 5000))
	when := time.Date(2026, 10, 25, 9, 41, 0, 0, time.UTC)
	if err := saveState(store.Root(), pruneState{LastPrune: when}); err != nil {
		t.Fatal(err)
	}
	status, _ := Inspect(dir, 1000)
	if !status.OverLimit() || status.LastPrune == nil || !status.LastPrune.Equal(when) {
		t.Fatalf("status %+v", status)
	}
	if status, _ := Inspect(dir, 0); status.OverLimit() {
		t.Fatal("no limit is never over it")
	}
}

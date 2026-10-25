package taskcache

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// savedStore holds one entry per name, each with an output of `size` bytes,
// last used `i` hours ago counting from the first name (oldest).
func savedStore(t *testing.T, names []string, size int) (*LocalStorage, string) {
	t.Helper()
	ws := t.TempDir()
	s, err := NewLocalStorage(filepath.Join(ws, "cache"))
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range names {
		dir := filepath.Join(ws, name)
		_ = os.MkdirAll(dir, 0o755)
		_ = os.WriteFile(filepath.Join(dir, "f"), bytes.Repeat([]byte(name[:1]), size), 0o644)
		key := fmt.Sprintf("%064s", name)
		if err := Save(s, ws, key, name, name, []string{"."}, []byte("log "+name), 0, nil); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(time.Duration(i-len(names)-5) * time.Hour)
		p, _ := s.path(AC, key)
		_ = os.Chtimes(p, old, old)
		for _, blob := range blobsOf(t, s, key) {
			bp, _ := s.path(CAS, blob)
			_ = os.Chtimes(bp, old, old)
		}
	}
	return s, ws
}

func blobsOf(t *testing.T, s Storage, key string) []string {
	t.Helper()
	e, err := Lookup(s, key)
	if err != nil {
		t.Fatal(err)
	}
	out := []string{e.Log}
	for _, o := range e.Outputs {
		out = append(out, o.Digest)
	}
	return out
}

func TestPruneWithDryRunReportsTheSameAsTheRealPruneAndRemovesNothing(t *testing.T) {
	s, _ := savedStore(t, []string{"aaa", "bbb", "ccc", "ddd"}, 1000)
	dry, err := PruneWith(s, PruneOptions{MaxBytes: 3200, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	var before int
	_ = s.Walk(AC, func(string, int64, time.Time) error { before++; return nil })
	if before != 4 {
		t.Fatal("a dry run removes nothing")
	}
	if dry.Entries != 2 || dry.Blobs != 4 || dry.SizeAfter > 3200 || dry.SizeBefore <= dry.SizeAfter || dry.Kept != 2 {
		t.Fatalf("dry run %+v", dry)
	}
	if dry.Tasks["aaa"] != 1 || dry.Tasks["bbb"] != 1 {
		t.Fatalf("the oldest tasks go: %v", dry.Tasks)
	}
	real, err := PruneWith(s, PruneOptions{MaxBytes: 3200})
	if err != nil {
		t.Fatal(err)
	}
	if real.Entries != dry.Entries || real.Blobs != dry.Blobs || real.Freed != dry.Freed || real.SizeAfter != dry.SizeAfter {
		t.Fatalf("real %+v differs from dry %+v", real, dry)
	}
}

func TestPruneWithAfterwardsTheStoreIsUnderTheLimitIncludingEntriesAndPointers(t *testing.T) {
	s, _ := savedStore(t, []string{"aaa", "bbb", "ccc", "ddd", "eee"}, 800)
	const limit = 3000
	res, err := PruneWith(s, PruneOptions{MaxBytes: limit})
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, kind := range []Kind{AC, CAS, Last} {
		_ = s.Walk(kind, func(_ string, size int64, _ time.Time) error { total += size; return nil })
	}
	if total > limit || total != res.SizeAfter {
		t.Fatalf("on disk %d, reported %d, limit %d", total, res.SizeAfter, limit)
	}
}

func TestPruneWithGraceKeepsWhatWasUsedRecently(t *testing.T) {
	s, _ := savedStore(t, []string{"aaa", "bbb"}, 1000)
	// Both entries were used a moment ago: with a grace, neither may go even
	// though the limit is far below their size.
	now := time.Now()
	for _, name := range []string{"aaa", "bbb"} {
		_ = s.Touch(AC, fmt.Sprintf("%064s", name))
		for _, b := range blobsOf(t, s, fmt.Sprintf("%064s", name)) {
			_ = s.Touch(CAS, b)
		}
	}
	res, err := PruneWith(s, PruneOptions{MaxBytes: 10, Grace: 5 * time.Minute, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if res.Entries != 0 || res.Blobs != 0 || res.Kept != 2 {
		t.Fatalf("result %+v", res)
	}
	later, err := PruneWith(s, PruneOptions{MaxBytes: 10, Grace: 5 * time.Minute, Now: now.Add(10 * time.Minute)})
	if err != nil || later.Entries != 2 {
		t.Fatalf("once the grace has passed they go: %+v (%v)", later, err)
	}
}

func TestPruneWithOrphanGraceKeepsABlobNoEntryNamesYetAndSweepsTheRest(t *testing.T) {
	s, _ := savedStore(t, []string{"aaa"}, 100)
	fresh := digest([]byte("being saved right now"))
	if err := s.Put(CAS, fresh, bytes.NewReader([]byte("being saved right now"))); err != nil {
		t.Fatal(err)
	}
	stale := digest([]byte("left behind long ago"))
	if err := s.Put(CAS, stale, bytes.NewReader([]byte("left behind long ago"))); err != nil {
		t.Fatal(err)
	}
	long := time.Now().Add(-time.Hour)
	p, _ := s.path(CAS, stale)
	_ = os.Chtimes(p, long, long)

	res, err := PruneWith(s, PruneOptions{MaxBytes: 1 << 30, OrphanGrace: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := s.Has(CAS, fresh); !ok {
		t.Fatal("a blob written a moment ago may belong to a save that has not written its entry yet")
	}
	if ok, _ := s.Has(CAS, stale); ok || res.Blobs != 1 {
		t.Fatalf("an old unused blob goes: %+v", res)
	}
}

// failingDelete is a store whose Delete is refused for chosen keys, as Windows
// does for a file another process has open.
type failingDelete struct {
	*LocalStorage
	refuse map[string]bool
}

func (f failingDelete) Delete(kind Kind, key string) error {
	if f.refuse[key] {
		return errors.New("The process cannot access the file because it is being used by another process")
	}
	return f.LocalStorage.Delete(kind, key)
}

func TestPruneWithSkipsAFileItCannotRemoveInsteadOfFailing(t *testing.T) {
	s, _ := savedStore(t, []string{"aaa", "bbb", "ccc"}, 1000)
	held := fmt.Sprintf("%064s", "aaa")
	res, err := PruneWith(failingDelete{s, map[string]bool{held: true}}, PruneOptions{MaxBytes: 2500})
	if err != nil {
		t.Fatalf("a held file is not a failure: %v", err)
	}
	if res.Skipped == 0 {
		t.Fatalf("result %+v", res)
	}
	if ok, _ := s.Has(AC, held); !ok {
		t.Fatal("the entry that could not be removed is still there")
	}
	for _, b := range blobsOf(t, s, held) {
		if ok, _ := s.Has(CAS, b); !ok {
			t.Fatal("and so are its blobs: an entry must never lose its blobs to a prune")
		}
	}
}

func TestWalkIgnoresADirectoryThatVanishes(t *testing.T) {
	s, _ := savedStore(t, []string{"aaa"}, 10)
	if err := os.RemoveAll(filepath.Join(s.Root(), "cas")); err != nil {
		t.Fatal(err)
	}
	if err := s.Walk(CAS, func(string, int64, time.Time) error { return nil }); err != nil {
		t.Fatalf("a store being cleared is empty, not an error: %v", err)
	}
}

func TestPutOfABlobThatIsAlreadyThereCountsAsAUse(t *testing.T) {
	s, _ := NewLocalStorage(filepath.Join(t.TempDir(), "cache"))
	content := []byte("shared by many tasks")
	key := digest(content)
	if err := s.Put(CAS, key, bytes.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	long := time.Now().Add(-48 * time.Hour)
	p, _ := s.path(CAS, key)
	_ = os.Chtimes(p, long, long)
	if err := s.Put(CAS, key, bytes.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(p)
	if time.Since(info.ModTime()) > time.Minute {
		t.Fatal("storing a blob again must refresh it, or a prune could sweep it before the new entry names it")
	}
}

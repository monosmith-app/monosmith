package cachemaintenance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

func kinds(r VerifyReport) map[string]int {
	out := map[string]int{}
	for _, p := range r.Problems {
		out[p.Kind]++
	}
	return out
}

func TestVerifyOfAHealthyStoreFindsNothing(t *testing.T) {
	store, dir := newStore(t)
	seedEntry(t, store, "web:build", []byte("one"), []byte("two"))
	seedEntry(t, store, "api:test", []byte("three"))
	report, err := Verify(dir, false, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Problems) != 0 || report.Entries != 2 || report.Blobs != 5 || report.Remaining() != 0 {
		t.Fatalf("report %+v", report)
	}
}

func TestVerifyOfACacheThatDoesNotExistFindsNothingAndCreatesNothing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "never")
	report, err := Verify(dir, true, time.Now())
	if err != nil || len(report.Problems) != 0 {
		t.Fatalf("report %+v, err %v", report, err)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("verifying must not create the cache")
	}
}

// corruptStore builds a store with one of each kind of damage and returns the
// entries that must survive.
func corruptStore(t *testing.T) (dir string, store *taskcache.LocalStorage, survivor string) {
	t.Helper()
	store, dir = newStore(t)
	survivor, _ = seedEntry(t, store, "good:build", []byte("fine"))

	// A blob whose content no longer matches its name, and the entry that names it.
	_, blobs := seedEntry(t, store, "bad:build", []byte("this will rot"))
	if err := os.WriteFile(filepath.Join(store.Root(), "cas", blobs[0][:2], blobs[0]), []byte("rotten"), 0o644); err != nil {
		t.Fatal(err)
	}
	// An entry whose blob is gone.
	_, blobs = seedEntry(t, store, "gone:build", []byte("lost"))
	if err := store.Delete(taskcache.CAS, blobs[0]); err != nil {
		t.Fatal(err)
	}
	// An entry that is not an entry.
	garbage := sha256.Sum256([]byte("garbage"))
	if err := store.Put(taskcache.AC, hex.EncodeToString(garbage[:]), bytes.NewReader([]byte("{not json"))); err != nil {
		t.Fatal(err)
	}
	// A blob nothing uses, and a write that died.
	orphan := putBlob(t, store, []byte("orphan"))
	age(t, store, taskcache.CAS, orphan, time.Hour)
	if err := os.WriteFile(filepath.Join(store.Root(), "tmp", "put-dead"), []byte("half"), 0o644); err != nil {
		t.Fatal(err)
	}
	long := time.Now().Add(-time.Hour)
	_ = os.Chtimes(filepath.Join(store.Root(), "tmp", "put-dead"), long, long)
	return dir, store, survivor
}

func TestVerifyFindsEveryKindOfDamage(t *testing.T) {
	dir, _, _ := corruptStore(t)
	report, err := Verify(dir, false, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	got := kinds(report)
	if got[CorruptBlob] != 1 || got[DamagedEntry] != 1 || got[MissingBlob] != 1 || got[UnreadableEntry] != 1 || got[OrphanBlob] < 1 || got[LeftoverTmp] != 1 {
		t.Fatalf("problems %v\n%+v", got, report.Problems)
	}
	if report.Remaining() != len(report.Problems) {
		t.Fatal("without --fix nothing is fixed")
	}
	for _, p := range report.Problems {
		if p.Fixed || p.Key == "" || p.Detail == "" {
			t.Fatalf("problem %+v", p)
		}
	}
}

func TestVerifyFixRemovesWhatIsBrokenAndKeepsTheRest(t *testing.T) {
	dir, store, survivor := corruptStore(t)
	report, err := Verify(dir, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if report.Remaining() != 0 || len(report.Problems) == 0 {
		t.Fatalf("a fix that works leaves nothing: %+v", report.Problems)
	}
	for _, p := range report.Problems {
		if !p.Fixed {
			t.Fatalf("not fixed: %+v", p)
		}
	}
	if !present(t, store, survivor) {
		t.Fatal("a healthy entry must survive a fix")
	}
	again, err := Verify(dir, false, time.Now())
	if err != nil || len(again.Problems) != 0 || again.Entries != 1 {
		t.Fatalf("after the fix: %+v (%v)", again, err)
	}
	if _, err := taskcache.Lookup(store, survivor); err != nil {
		t.Fatalf("the surviving entry still hits: %v", err)
	}
}

func TestVerifyLeavesAlonePartsOfAWriteThatIsStillRunning(t *testing.T) {
	store, dir := newStore(t)
	putBlob(t, store, []byte("a blob whose entry is about to be written"))
	if err := os.WriteFile(filepath.Join(store.Root(), "tmp", "put-live"), []byte("half"), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Verify(dir, true, time.Now())
	if err != nil || len(report.Problems) != 0 {
		t.Fatalf("report %+v, err %v", report, err)
	}
	if _, err := os.Stat(filepath.Join(store.Root(), "tmp", "put-live")); err != nil {
		t.Fatal("a fresh tmp file must not be removed by --fix")
	}
}

func TestVerifyRemovesAnUnreadableEntryEvenWhenNothingElseIsWrong(t *testing.T) {
	store, dir := newStore(t)
	garbage := hex.EncodeToString(sha256.New().Sum([]byte("x"))[:32])
	if err := store.Put(taskcache.AC, garbage, bytes.NewReader([]byte("{"))); err != nil {
		t.Fatal(err)
	}
	report, _ := Verify(dir, false, time.Now())
	if kinds(report)[UnreadableEntry] != 1 {
		t.Fatalf("report %+v", report)
	}
	report, _ = Verify(dir, true, time.Now())
	if report.Remaining() != 0 {
		t.Fatalf("report %+v", report)
	}
	if ok, _ := store.Has(taskcache.AC, garbage); ok {
		t.Fatal("the unreadable entry must be gone")
	}
}

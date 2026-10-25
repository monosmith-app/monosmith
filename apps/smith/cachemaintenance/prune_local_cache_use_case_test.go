package cachemaintenance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

const childDirVariable = "SMITH_CACHE_TEST_PRUNER_DIR"

// seedOld fills a store with `n` entries of `size` bytes, each last used a
// while ago, oldest first.
func seedOld(t testing.TB, store *taskcache.LocalStorage, n, size int) []string {
	t.Helper()
	var keys []string
	for i := 0; i < n; i++ {
		key, blobs := seedEntry(t, store, fmt.Sprintf("p%d:build", i), bytes.Repeat([]byte{byte(i + 1)}, size))
		age(t, store, taskcache.AC, key, time.Duration(n-i)*time.Hour)
		for _, b := range blobs {
			age(t, store, taskcache.CAS, b, time.Duration(n-i)*time.Hour)
		}
		keys = append(keys, key)
	}
	return keys
}

func present(t testing.TB, store *taskcache.LocalStorage, key string) bool {
	t.Helper()
	ok, err := store.Has(taskcache.AC, key)
	if err != nil {
		t.Fatal(err)
	}
	return ok
}

func TestAfterRunWithNoLimitReadsAndWritesNothing(t *testing.T) {
	store, dir := newStore(t)
	seedOld(t, store, 3, 1000)
	for _, limit := range []int64{0, -1} {
		checked, _, err := AfterRun(dir, limit, 50<<20, time.Now())
		if err != nil || checked {
			t.Fatalf("limit %d: checked %v, err %v", limit, checked, err)
		}
	}
	if _, err := os.Stat(filepath.Join(store.Root(), stateFile)); err == nil {
		t.Fatal("with no limit there is nothing to remember")
	}
}

func TestAfterRunPrunesTheOldestEntriesOnTheFirstCheck(t *testing.T) {
	store, dir := newStore(t)
	keys := seedOld(t, store, 5, 4000)

	checked, result, err := AfterRun(dir, 10000, 100, time.Now())
	if err != nil || !checked {
		t.Fatalf("checked %v, err %v", checked, err)
	}
	if result.Entries == 0 || result.SizeAfter > 10000 {
		t.Fatalf("result %+v", result)
	}
	if present(t, store, keys[0]) || !present(t, store, keys[4]) {
		t.Fatal("the least recently used entry goes first, the newest stays")
	}
	st := loadState(store.Root())
	if st.LastPrune.IsZero() || st.SizeBytes != result.SizeAfter || st.AddedBytes != 0 {
		t.Fatalf("state %+v", st)
	}
}

func TestAfterRunDoesNotWalkTheStoreAgainWithinTheHour(t *testing.T) {
	store, dir := newStore(t)
	seedOld(t, store, 2, 1000)
	now := time.Now()
	if checked, _, err := AfterRun(dir, 1<<20, 10, now); err != nil || !checked {
		t.Fatalf("first run: checked %v, err %v", checked, err)
	}
	// The store now grows far past the limit behind smith's back; a small run a
	// minute later must not notice, because it does not look.
	extra := seedOld(t, store, 6, 400000)
	checked, _, err := AfterRun(dir, 1<<20, 2<<20, now.Add(time.Minute))
	if err != nil || checked {
		t.Fatalf("second run: checked %v, err %v", checked, err)
	}
	for _, k := range extra {
		if !present(t, store, k) {
			t.Fatal("a run that does not check removes nothing")
		}
	}
	if st := loadState(store.Root()); st.AddedBytes != 2<<20 {
		t.Fatalf("the total of what runs stored must grow: %+v", st)
	}
	// An hour later it looks again, and finds what grew.
	checked, result, err := AfterRun(dir, 1<<20, 10, now.Add(time.Hour+time.Second))
	if err != nil || !checked || result.Entries == 0 {
		t.Fatalf("an hour later: checked %v, result %+v, err %v", checked, result, err)
	}
}

func TestAfterRunChecksSoonerAfterABigRunThatLeavesTheStoreOverTheLimit(t *testing.T) {
	store, dir := newStore(t)
	now := time.Now()
	if _, _, err := AfterRun(dir, 1<<20, 10, now); err != nil {
		t.Fatal(err)
	}
	seedOld(t, store, 4, 400000)
	checked, result, err := AfterRun(dir, 1<<20, BigRunBytes, now.Add(time.Minute))
	if err != nil || !checked || result.Entries == 0 {
		t.Fatalf("checked %v, result %+v, err %v", checked, result, err)
	}
}

func TestAfterRunLooksAgainSoonWhenWhatKeepsTheStoreOverTheLimitIsRecent(t *testing.T) {
	store, dir := newStore(t)
	// Entries used just now are protected, so the prune cannot reach the limit.
	for i := 0; i < 3; i++ {
		seedEntry(t, store, fmt.Sprintf("p%d:build", i), bytes.Repeat([]byte{byte(i + 1)}, 5000))
	}
	now := time.Now()
	checked, result, err := AfterRun(dir, 6000, 100, now)
	if err != nil || !checked || result.SizeAfter <= 6000 {
		t.Fatalf("checked %v, result %+v, err %v", checked, result, err)
	}
	later := now.Add(RetryAfter + time.Second)
	if st := loadState(store.Root()); !shouldCheck(st, later, 1, 6000) {
		t.Fatalf("an hour is too long to wait: %+v", st)
	}
}

func TestAfterRunNeverDeletesWhatWasJustStored(t *testing.T) {
	store, dir := newStore(t)
	keys := seedOld(t, store, 3, 3000)
	fresh, blobs := seedEntry(t, store, "web:build", bytes.Repeat([]byte("n"), 3000))
	if _, _, err := AfterRun(dir, 5000, 3000, time.Now()); err != nil {
		t.Fatal(err)
	}
	if !present(t, store, fresh) {
		t.Fatal("the entry this run stored must survive")
	}
	for _, b := range blobs {
		if ok, _ := store.Has(taskcache.CAS, b); !ok {
			t.Fatal("and so must its blobs")
		}
	}
	if present(t, store, keys[0]) {
		t.Fatal("older entries go instead")
	}
}

func TestAfterRunSkipsWhenAnotherProcessIsPruning(t *testing.T) {
	store, dir := newStore(t)
	seedOld(t, store, 4, 4000)
	release, ok := lockPrune(store.Root(), time.Now())
	if !ok {
		t.Fatal("the lock is free")
	}
	checked, _, err := AfterRun(dir, 1000, 100, time.Now())
	if err != nil || checked {
		t.Fatalf("while another process prunes: checked %v, err %v", checked, err)
	}
	release()
	if checked, _, err := AfterRun(dir, 1000, 100, time.Now()); err != nil || !checked {
		t.Fatalf("after it finishes: checked %v, err %v", checked, err)
	}
}

func TestALockLeftByAProcessThatDiedIsTakenOver(t *testing.T) {
	store, _ := newStore(t)
	now := time.Now()
	release, ok := lockPrune(store.Root(), now)
	if !ok {
		t.Fatal("the lock is free")
	}
	defer release()
	if _, again := lockPrune(store.Root(), now.Add(time.Minute)); again {
		t.Fatal("a fresh lock is held")
	}
	old := now.Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(store.Root(), lockFile), old, old); err != nil {
		t.Fatal(err)
	}
	release2, again := lockPrune(store.Root(), now)
	if !again {
		t.Fatal("a stale lock is taken over")
	}
	release2()
}

func TestAfterRunFailsWithAnErrorTheCallerCanWarnAbout(t *testing.T) {
	store, dir := newStore(t)
	seedOld(t, store, 2, 100)
	// A file where the tmp directory should be: the state cannot be written.
	if err := os.RemoveAll(filepath.Join(store.Root(), "tmp")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.Root(), "tmp"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := AfterRun(dir, 1<<20, 10, time.Now()); err == nil {
		t.Fatal("an unwritable state is an error, which the run turns into a warning")
	}
}

func TestPruneNowDryRunRemovesNothingAndSaysWhatWouldGo(t *testing.T) {
	store, dir := newStore(t)
	keys := seedOld(t, store, 4, 4000)
	out, err := PruneNow(dir, 9000, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !out.DryRun || !out.Exists || out.Entries == 0 || out.SizeAfter > 9000 || out.SizeBefore <= 9000 || out.Freed == 0 || len(out.Tasks) != out.Entries {
		t.Fatalf("outcome %+v", out)
	}
	for _, k := range keys {
		if !present(t, store, k) {
			t.Fatal("a dry run removes nothing")
		}
	}
	if st := loadState(store.Root()); !st.LastPrune.IsZero() {
		t.Fatal("a dry run is not a prune")
	}
	real, err := PruneNow(dir, 9000, false, time.Now())
	if err != nil || real.Entries != out.Entries || real.SizeAfter != out.SizeAfter {
		t.Fatalf("the real prune must do what the dry run said: %+v vs %+v (%v)", real, out, err)
	}
	if present(t, store, keys[0]) {
		t.Fatal("the oldest entry must go")
	}
}

func TestPruneNowWithoutALimitRemovesOnlyUnusedBlobs(t *testing.T) {
	store, dir := newStore(t)
	keys := seedOld(t, store, 2, 100)
	orphan := putBlob(t, store, []byte("nobody wants me"))
	age(t, store, taskcache.CAS, orphan, time.Hour)
	out, err := PruneNow(dir, 0, false, time.Now())
	if err != nil || out.Entries != 0 || out.Blobs != 1 {
		t.Fatalf("outcome %+v, err %v", out, err)
	}
	for _, k := range keys {
		if !present(t, store, k) {
			t.Fatal("entries stay")
		}
	}
}

func TestPruneNowOfACacheThatDoesNotExistDoesNothing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "never")
	out, err := PruneNow(dir, 100, false, time.Now())
	if err != nil || out.Exists {
		t.Fatalf("outcome %+v, err %v", out, err)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("pruning must not create the cache")
	}
}

func TestPruneNowLeavesABlobAWriterHasNotNamedYet(t *testing.T) {
	store, dir := newStore(t)
	seedOld(t, store, 1, 100)
	inFlight := putBlob(t, store, []byte("a task is still saving this"))
	if _, err := PruneNow(dir, 0, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	if ok, _ := store.Has(taskcache.CAS, inFlight); !ok {
		t.Fatal("a blob written a moment ago may belong to a save in progress")
	}
}

func TestPruneNowSweepsOnlyOldLeftoversFromTmp(t *testing.T) {
	store, dir := newStore(t)
	seedOld(t, store, 1, 100)
	oldTmp := filepath.Join(store.Root(), "tmp", "put-old")
	newTmp := filepath.Join(store.Root(), "tmp", "put-new")
	for _, p := range []string{oldTmp, newTmp} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	long := time.Now().Add(-time.Hour)
	_ = os.Chtimes(oldTmp, long, long)
	if _, err := PruneNow(dir, 1<<20, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldTmp); err == nil {
		t.Fatal("a write that died an hour ago is swept")
	}
	if _, err := os.Stat(newTmp); err != nil {
		t.Fatal("a write that may be running is not")
	}
}

// TestTwoSmithProcessesPruneAndStoreAtTheSameTime runs a second process that
// prunes in a loop, with a limit that removes everything it may, while this one
// stores and looks up entries. Neither may fail, and nothing may be left
// half-removed.
func TestTwoSmithProcessesPruneAndStoreAtTheSameTime(t *testing.T) {
	if testing.Short() {
		t.Skip("starts a second process")
	}
	store, dir := newStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestChildPrunerLoop$")
	child.Env = append(os.Environ(), childDirVariable+"="+dir)
	var childOut bytes.Buffer
	child.Stdout, child.Stderr = &childOut, &childOut
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- child.Wait() }()

	hits := 0
	for i := 0; i < 60; i++ {
		key, _ := seedEntry(t, store, fmt.Sprintf("p%d:build", i), bytes.Repeat([]byte{byte(i)}, 2000))
		if _, err := taskcache.Lookup(store, key); err == nil {
			hits++
		} else if !errors.Is(err, taskcache.ErrMiss) {
			t.Fatalf("lookup %d: %v", i, err)
		}
	}
	if err := <-done; err != nil {
		t.Fatalf("the pruning process failed: %v\n%s", err, childOut.String())
	}
	report, err := Verify(dir, false, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range report.Problems {
		if p.Kind == CorruptBlob || p.Kind == UnreadableEntry || p.Kind == DamagedEntry {
			t.Fatalf("pruning beside a writer damaged the store: %+v", p)
		}
	}
	t.Logf("%d of 60 lookups hit while the other process pruned; %d problems of the harmless kind", hits, len(report.Problems))
}

// TestChildPrunerLoop is the second process of the test above; it does nothing
// when run on its own.
func TestChildPrunerLoop(t *testing.T) {
	dir := os.Getenv(childDirVariable)
	if dir == "" {
		t.Skip("only runs as the second process of TestTwoSmithProcessesPruneAndStoreAtTheSameTime")
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 2; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 80; i++ {
				if _, err := PruneNow(dir, 1, false, time.Now()); err != nil {
					t.Errorf("prune: %v", err)
					return
				}
				time.Sleep(2 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
}

// BenchmarkAfterRunOnALargeStore measures what a run pays for the prune check on
// a store of several thousand entries: the cheap path (the hour has not passed)
// and the walk (the first check, or the hourly one).
func BenchmarkAfterRunOnALargeStore(b *testing.B) {
	dir := filepath.Join(b.TempDir(), "cache")
	store, err := taskcache.NewLocalStorage(dir)
	if err != nil {
		b.Fatal(err)
	}
	const entries = 4000
	for i := 0; i < entries; i++ {
		seedEntry(b, store, fmt.Sprintf("p%d:build", i), []byte(fmt.Sprintf("output %d %s", i, bytes.Repeat([]byte("x"), 2000))))
	}
	limit := int64(1 << 40)
	now := time.Now()
	if _, _, err := AfterRun(dir, limit, 1, now); err != nil {
		b.Fatal(err)
	}
	b.Run("within the hour", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, _, err := AfterRun(dir, limit, 1<<10, now.Add(time.Minute)); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("the walk", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, _, err := AfterRun(dir, limit, 1<<10, now.Add(time.Duration(i+2)*(CheckInterval+time.Second))); err != nil {
				b.Fatal(err)
			}
		}
	})
}

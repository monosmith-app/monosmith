package cachemaintenance

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// AfterRun keeps the local store under limit bytes after a run that stored
// `added` bytes of results. It is called on every such run and is cheap: unless
// shouldCheck says the store is due for a look, it only adds to the running
// total in the state file. It reports whether it walked the store and, if it
// pruned, what that removed. A limit of zero or less means no limit and reads
// nothing at all. The caller treats an error as a warning, never a failure.
func AfterRun(dir string, limit, added int64, now time.Time) (checked bool, result taskcache.PruneResult, err error) {
	if limit <= 0 {
		return false, taskcache.PruneResult{}, nil
	}
	store, ok, err := openExisting(dir)
	if err != nil || !ok {
		return false, taskcache.PruneResult{}, err
	}
	root := store.Root()
	st := loadState(root)
	if !shouldCheck(st, now, added, limit) {
		st.AddedBytes += added
		return false, taskcache.PruneResult{}, saveState(root, st)
	}
	release, locked := lockPrune(root, now)
	if !locked {
		// Another smith is pruning right now; its result will do.
		return false, taskcache.PruneResult{}, nil
	}
	defer release()
	result, err = taskcache.PruneWith(store, taskcache.PruneOptions{MaxBytes: limit, Grace: InUseGrace, OrphanGrace: OrphanGrace, Now: now})
	if err != nil {
		return true, result, err
	}
	sweepTmp(root, now, InUseGrace)
	next := pruneState{LastCheck: now, LastPrune: now, SizeBytes: result.SizeAfter}
	if result.SizeAfter > limit {
		// What kept it over is protected for a few minutes: look again then,
		// not in an hour.
		next.LastCheck = now.Add(RetryAfter - CheckInterval)
	}
	return true, result, saveState(root, next)
}

// PruneNow is smith cache prune: it prunes the store to limit bytes (zero or
// less means no limit, which removes only what no entry uses). With dryRun it
// removes nothing and reports what would go.
func PruneNow(dir string, limit int64, dryRun bool, now time.Time) (PruneOutcome, error) {
	out := PruneOutcome{Dir: dir, LimitBytes: limit, DryRun: dryRun}
	store, ok, err := openExisting(dir)
	if err != nil || !ok {
		return out, err
	}
	out.Exists = true
	max := limit
	if max <= 0 {
		max = 1<<63 - 1
	}
	result, err := taskcache.PruneWith(store, taskcache.PruneOptions{MaxBytes: max, DryRun: dryRun, OrphanGrace: OrphanGrace, Now: now})
	if err != nil {
		return out, err
	}
	out.Entries, out.Blobs, out.Freed = result.Entries, result.Blobs, result.Freed
	out.SizeBefore, out.SizeAfter, out.Kept, out.Skipped, out.Tasks = result.SizeBefore, result.SizeAfter, result.Kept, result.Skipped, result.Tasks
	if dryRun {
		return out, nil
	}
	sweepTmp(store.Root(), now, InUseGrace)
	return out, saveState(store.Root(), pruneState{LastCheck: now, LastPrune: now, SizeBytes: result.SizeAfter})
}

// openExisting opens the store without creating it: a workspace that never
// cached anything has nothing to look after.
func openExisting(dir string) (*taskcache.LocalStorage, bool, error) {
	if _, err := os.Stat(filepath.Join(dir, "v1")); errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	store, err := taskcache.NewLocalStorage(dir)
	return store, err == nil, err
}

// sweepTmp removes files a killed process left in tmp/. A younger file may be
// a write in progress.
func sweepTmp(root string, now time.Time, grace time.Duration) {
	entries, err := os.ReadDir(filepath.Join(root, "tmp"))
	if err != nil {
		return
	}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) < grace {
			continue
		}
		_ = os.Remove(filepath.Join(root, "tmp", e.Name()))
	}
}

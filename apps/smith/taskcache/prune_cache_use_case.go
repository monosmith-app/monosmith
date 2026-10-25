package taskcache

import (
	"encoding/json"
	"sort"
	"time"
)

// PruneResult reports what pruning removed, or would remove in a dry run.
type PruneResult struct {
	// Entries and Blobs are the action-cache entries and content-addressed
	// blobs removed.
	Entries, Blobs int
	// Freed is the bytes of the removed blobs and entries.
	Freed int64
	// SizeBefore and SizeAfter are the bytes of the entries and blobs before
	// and after (for a dry run, as they would be).
	SizeBefore, SizeAfter int64
	// Kept is how many entries remain.
	Kept int
	// Skipped counts files that were due to go but could not be removed (a
	// file another process holds open on Windows); the next prune retries them.
	Skipped int
	// Tasks counts the removed entries by task.
	Tasks map[string]int
}

// PruneOptions tune PruneWith.
type PruneOptions struct {
	// MaxBytes is the size the entries and blobs must fit.
	MaxBytes int64
	// DryRun removes nothing and reports what would go.
	DryRun bool
	// Grace protects what another process may be using: an entry used, and a
	// blob written or used, within this long is neither removed nor swept.
	// Zero protects nothing.
	Grace time.Duration
	// OrphanGrace protects a blob no entry references yet: a process writes
	// its blobs before the entry that names them. It applies on top of Grace.
	OrphanGrace time.Duration
	// Now is the clock; zero means time.Now().
	Now time.Time
}

// Prune keeps the cache under maxBytes: it removes the least recently used
// action-cache entries until the blobs the rest reference fit, then sweeps
// every blob no remaining entry references.
func Prune(s Storage, maxBytes int64) (PruneResult, error) {
	return PruneWith(s, PruneOptions{MaxBytes: maxBytes})
}

// PruneWith is Prune with a dry run and a grace period. It is safe beside other
// smith processes: a file that vanishes while it runs is ignored, a file it
// cannot remove is skipped, and nothing used or written within the grace is
// touched, so it never removes what another process is reading or writing.
func PruneWith(s Storage, opts PruneOptions) (PruneResult, error) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	type entry struct {
		key      string
		size     int64
		lastUsed time.Time
		task     string
		blobs    []string
	}
	var entries []entry
	err := s.Walk(AC, func(key string, size int64, last time.Time) error {
		e := entry{key: key, size: size, lastUsed: last}
		if r, err := s.Get(AC, key); err == nil {
			var parsed Entry
			if json.NewDecoder(r).Decode(&parsed) == nil {
				e.task = parsed.Task
				e.blobs = append(e.blobs, parsed.Log)
				for _, o := range parsed.Outputs {
					e.blobs = append(e.blobs, o.Digest)
				}
			}
			_ = r.Close()
		}
		entries = append(entries, e)
		return nil
	})
	if err != nil {
		return PruneResult{}, err
	}
	type blob struct {
		size int64
		last time.Time
	}
	blobs := map[string]blob{}
	var blobBytes int64
	if err := s.Walk(CAS, func(key string, size int64, last time.Time) error {
		blobs[key] = blob{size, last}
		blobBytes += size
		return nil
	}); err != nil {
		return PruneResult{}, err
	}
	// The last/ pointers are a few dozen bytes each and are never removed, but
	// they count toward the size, so a cache pruned to the limit is under it.
	var pointerBytes int64
	if err := s.Walk(Last, func(_ string, size int64, _ time.Time) error { pointerBytes += size; return nil }); err != nil && err != ErrNotSupported {
		return PruneResult{}, err
	}
	sort.Slice(entries, func(i, j int) bool {
		if !entries[i].lastUsed.Equal(entries[j].lastUsed) {
			return entries[i].lastUsed.After(entries[j].lastUsed)
		}
		return entries[i].key < entries[j].key
	})
	protected := func(last time.Time) bool { return opts.Grace > 0 && now.Sub(last) < opts.Grace }
	named := map[string]bool{}
	for _, e := range entries {
		for _, b := range e.blobs {
			named[b] = true
		}
	}

	result := PruneResult{Tasks: map[string]int{}}
	for _, e := range entries {
		result.SizeBefore += e.size
	}
	result.SizeBefore += blobBytes + pointerBytes
	referenced := map[string]bool{}
	total := pointerBytes
	for _, e := range entries {
		add := e.size
		for _, b := range e.blobs {
			if !referenced[b] {
				add += blobs[b].size
			}
		}
		if total+add > opts.MaxBytes && !protected(e.lastUsed) {
			if !opts.DryRun {
				if err := s.Delete(AC, e.key); err != nil {
					// Another process holds it: keep it, as if it were in use.
					result.Skipped++
					for _, b := range e.blobs {
						referenced[b] = true
					}
					total += add
					result.Kept++
					continue
				}
			}
			result.Entries++
			result.Freed += e.size
			result.Tasks[e.task]++
			continue
		}
		for _, b := range e.blobs {
			referenced[b] = true
		}
		total += add
		result.Kept++
	}
	for key, b := range blobs {
		if referenced[key] || protected(b.last) || (!named[key] && opts.OrphanGrace > 0 && now.Sub(b.last) < opts.OrphanGrace) {
			if !referenced[key] {
				total += b.size
			}
			continue
		}
		if !opts.DryRun {
			if err := s.Delete(CAS, key); err != nil {
				result.Skipped++
				total += b.size
				continue
			}
		}
		result.Blobs++
		result.Freed += b.size
	}
	result.SizeAfter = total
	return result, nil
}

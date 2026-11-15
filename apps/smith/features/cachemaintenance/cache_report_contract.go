package cachemaintenance

import "time"

// InUseGrace is how long a file written or used is left alone by the prune
// that follows a run, in case another smith process is reading or writing it.
const InUseGrace = 5 * time.Minute

// OrphanGrace is how long a blob that no entry names yet is left alone by any
// prune or check: a process writes its blobs before the entry that names them.
const OrphanGrace = time.Minute

// TaskUsage is the size of the largest cached result of one task.
type TaskUsage struct {
	Task    string `json:"task"`
	Entries int    `json:"entries"`
	// LargestBytes is the blobs the task's largest entry names.
	LargestBytes int64 `json:"largestBytes"`
}

// Status is what the local cache holds.
type Status struct {
	Dir string `json:"dir"`
	// Exists is false before the first cached run.
	Exists  bool  `json:"exists"`
	Entries int   `json:"entries"`
	Blobs   int   `json:"blobs"`
	Bytes   int64 `json:"bytes"`
	// LimitBytes is [cache] max_size_mb in bytes; 0 means no limit.
	LimitBytes int64 `json:"limitBytes"`
	// LastPrune is nil when the cache was never pruned.
	LastPrune *time.Time `json:"lastPrune,omitempty"`
	// BrokenEntries are entries a lookup would miss: unreadable, or naming a
	// blob that is gone. (smith cache verify also checks the blobs' content.)
	BrokenEntries int `json:"brokenEntries"`
	// Largest are the five tasks with the largest entries.
	Largest []TaskUsage `json:"largest"`
}

// OverLimit reports whether the cache is bigger than its limit.
func (s Status) OverLimit() bool { return s.LimitBytes > 0 && s.Bytes > s.LimitBytes }

// Problem kinds that smith cache verify reports.
const (
	CorruptBlob     = "corrupt-blob"
	DamagedEntry    = "damaged-entry"
	MissingBlob     = "missing-blob"
	UnreadableEntry = "unreadable-entry"
	OrphanBlob      = "orphan-blob"
	LeftoverTmp     = "leftover-tmp"
)

// Problem is one thing wrong in the store.
type Problem struct {
	Kind   string `json:"kind"`
	Key    string `json:"key"`
	Detail string `json:"detail"`
	// Fixed is true when a fix removed it.
	Fixed bool `json:"fixed,omitempty"`
}

// VerifyReport is the result of checking the store.
type VerifyReport struct {
	Dir      string    `json:"dir"`
	Entries  int       `json:"entries"`
	Blobs    int       `json:"blobs"`
	Problems []Problem `json:"problems"`
}

// Remaining counts the problems no fix removed.
func (r VerifyReport) Remaining() int {
	n := 0
	for _, p := range r.Problems {
		if !p.Fixed {
			n++
		}
	}
	return n
}

// ClearResult is what clearing the local cache removed.
type ClearResult struct {
	Dir     string `json:"dir"`
	Entries int    `json:"entries"`
	Blobs   int    `json:"blobs"`
	Bytes   int64  `json:"bytes"`
}

// PruneOutcome is the result of pruning the local cache.
type PruneOutcome struct {
	Dir        string `json:"dir"`
	LimitBytes int64  `json:"limitBytes"`
	DryRun     bool   `json:"dryRun"`
	// Exists is false before the first cached run.
	Exists bool `json:"exists"`
	// Entries and Blobs are removed (or, in a dry run, would be).
	Entries    int            `json:"entries"`
	Blobs      int            `json:"blobs"`
	Freed      int64          `json:"freedBytes"`
	SizeBefore int64          `json:"sizeBeforeBytes"`
	SizeAfter  int64          `json:"sizeAfterBytes"`
	Kept       int            `json:"entriesKept"`
	Skipped    int            `json:"skipped"`
	Tasks      map[string]int `json:"tasks,omitempty"`
}

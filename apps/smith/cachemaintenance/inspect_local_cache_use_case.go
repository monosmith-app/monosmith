package cachemaintenance

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// largestShown is how many tasks Status lists.
const largestShown = 5

// Inspect measures the local cache in dir against a limit (zero: none). It
// stats and reads entries but never hashes a blob, so it stays cheap enough for
// smith doctor and smith info; Verify is the full check. A directory that does
// not exist yet is an empty cache.
func Inspect(dir string, limit int64) (Status, error) {
	status := Status{Dir: dir, LimitBytes: limit, Largest: []TaskUsage{}}
	store, ok, err := openExisting(dir)
	if err != nil || !ok {
		return status, err
	}
	status.Exists = true
	if st := loadState(store.Root()); !st.LastPrune.IsZero() {
		last := st.LastPrune
		status.LastPrune = &last
	}
	blobSize := map[string]int64{}
	if err := store.Walk(taskcache.CAS, func(key string, size int64, _ time.Time) error {
		blobSize[key] = size
		status.Blobs++
		status.Bytes += size
		return nil
	}); err != nil {
		return status, err
	}
	perTask := map[string]*TaskUsage{}
	err = store.Walk(taskcache.AC, func(key string, size int64, _ time.Time) error {
		status.Entries++
		status.Bytes += size
		entry, state := readEntry(store, key)
		if state == entryGone {
			status.Entries-- // removed since the walk saw it
			status.Bytes -= size
			return nil
		}
		if state == entryBad {
			status.BrokenEntries++
			return nil
		}
		var named int64
		broken := false
		for _, digest := range entryBlobs(entry) {
			s, present := blobSize[digest]
			if !present {
				broken = true
			}
			named += s
		}
		if broken {
			status.BrokenEntries++
		}
		usage := perTask[entry.Task]
		if usage == nil {
			usage = &TaskUsage{Task: entry.Task}
			perTask[entry.Task] = usage
		}
		usage.Entries++
		usage.LargestBytes = max(usage.LargestBytes, named)
		return nil
	})
	if err != nil {
		return status, err
	}
	if err := store.Walk(taskcache.Last, func(_ string, size int64, _ time.Time) error {
		status.Bytes += size
		return nil
	}); err != nil {
		return status, err
	}
	for _, u := range perTask {
		status.Largest = append(status.Largest, *u)
	}
	sort.Slice(status.Largest, func(i, j int) bool {
		a, b := status.Largest[i], status.Largest[j]
		if a.LargestBytes != b.LargestBytes {
			return a.LargestBytes > b.LargestBytes
		}
		return a.Task < b.Task
	})
	if len(status.Largest) > largestShown {
		status.Largest = status.Largest[:largestShown]
	}
	return status, nil
}

type entryState int

const (
	entryOK entryState = iota
	// entryGone: removed since it was listed.
	entryGone
	// entryBad: present but not an entry.
	entryBad
)

// readEntry decodes an action-cache entry.
func readEntry(store taskcache.Storage, key string) (taskcache.Entry, entryState) {
	r, err := store.Get(taskcache.AC, key)
	if errors.Is(err, taskcache.ErrMiss) {
		return taskcache.Entry{}, entryGone
	}
	if err != nil {
		return taskcache.Entry{}, entryBad
	}
	defer func() { _ = r.Close() }()
	var e taskcache.Entry
	if json.NewDecoder(r).Decode(&e) != nil || e.Log == "" {
		return taskcache.Entry{}, entryBad
	}
	return e, entryOK
}

// entryBlobs are the digests an entry names: its log and its output files.
func entryBlobs(e taskcache.Entry) []string {
	out := []string{e.Log}
	for _, o := range e.Outputs {
		out = append(out, o.Digest)
	}
	return out
}

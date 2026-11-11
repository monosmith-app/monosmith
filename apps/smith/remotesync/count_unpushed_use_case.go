package remotesync

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

// Pending is how many local entries the remote cache does not hold.
type Pending struct {
	// Count is the number of entries found missing on the remote.
	Count int `json:"count"`
	// AtLeast is true when the count stopped early (the limit or the time ran
	// out): the remote lacks Count entries or more.
	AtLeast bool `json:"atLeast,omitempty"`
	// Checked is how many local entries were asked about.
	Checked int `json:"checked"`
}

// CountLimits bound the work of CountUnpushed, so a large cache cannot make
// the caller slow.
type CountLimits struct {
	// MaxEntries is the most local entries asked about (0: no limit).
	MaxEntries int
	// Parallel is how many requests run at once (0: 8).
	Parallel int
}

// CountUnpushed asks the remote, with HEAD, about each local action-cache
// entry and counts those it lacks. It stops at limits.MaxEntries entries and
// when ctx ends, and then reports the count as "at least".
//
// Parameters: ctx bounds the time; local is the local cache; store is the
// remote; limits bound the work.
// Returns: the pending count; an error when the local cache cannot be read
// or the remote refuses the credentials (remotecache.Refused).
func CountUnpushed(ctx context.Context, local taskcache.Storage, store remotecache.ObjectStore, limits CountLimits) (Pending, error) {
	keys, more, err := localEntries(local, limits.MaxEntries)
	if err != nil {
		return Pending{}, err
	}
	parallel := limits.Parallel
	if parallel <= 0 {
		parallel = 8
	}
	var (
		mu      sync.Mutex
		out     = Pending{AtLeast: more}
		refused error
		wg      sync.WaitGroup
		work    = make(chan string)
	)
	for range parallel {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for key := range work {
				has, err := store.Has(remotecache.ObjectName(taskcache.AC, key))
				mu.Lock()
				switch {
				case remotecache.Refused(err):
					refused = err
				case err != nil:
					// One failed question: the count is a floor from here on.
					out.AtLeast = true
				default:
					out.Checked++
					if !has {
						out.Count++
					}
				}
				mu.Unlock()
			}
		}()
	}
feed:
	for _, key := range keys {
		select {
		case <-ctx.Done():
			mu.Lock()
			out.AtLeast = true
			mu.Unlock()
			break feed
		case work <- key:
		}
		mu.Lock()
		stop := refused != nil
		mu.Unlock()
		if stop {
			break
		}
	}
	close(work)
	finished := make(chan struct{})
	go func() { wg.Wait(); close(finished) }()
	select {
	case <-finished:
	case <-ctx.Done():
		// A request that hangs is left to its own timeout; the count so far is a floor.
	}
	mu.Lock()
	defer mu.Unlock()
	snapshot := out
	select {
	case <-finished:
	default:
		snapshot.AtLeast = true
	}
	return snapshot, refused
}

// localEntries lists the keys of the local action-cache entries, at most max
// (0: all), the most recently used first; more is true when there were more.
func localEntries(local taskcache.Storage, max int) (keys []string, more bool, err error) {
	type entry struct {
		key  string
		used time.Time
	}
	var all []entry
	err = local.Walk(taskcache.AC, func(key string, _ int64, used time.Time) error {
		all = append(all, entry{key, used})
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].used.After(all[j].used) })
	if max > 0 && len(all) > max {
		all, more = all[:max], true
	}
	for _, e := range all {
		keys = append(keys, e.key)
	}
	return keys, more, nil
}

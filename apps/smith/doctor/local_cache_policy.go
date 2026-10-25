package doctor

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// localCacheFindings judge the local task cache from a cheap look at it (a stat
// of each file and a read of each entry; no blob is hashed, which is what
// `smith cache verify` does). The cache over its limit means automatic pruning
// has not caught up (it runs after runs that stored results, about hourly); an
// entry that cannot be used is a miss that never heals by itself. Both only
// warn: a run is slower, never wrong.
func localCacheFindings(s cachemaintenance.Status) []protocol.Finding {
	if !s.Exists {
		return nil
	}
	var out []protocol.Finding
	out = append(out, protocol.Finding{
		Check: "the local cache is within its limit", OK: !s.OverLimit(), Warning: true,
		Detail: fmt.Sprintf("the cache at %s is %s, over [cache] max_size_mb (%s)", s.Dir, size(s.Bytes), size(s.LimitBytes)),
		Remedy: "run `smith cache prune` to trim it now; smith also trims it after runs that stored results",
	})
	out = append(out, protocol.Finding{
		Check: "the local cache has no unusable entries", OK: s.BrokenEntries == 0, Warning: true,
		Detail: fmt.Sprintf("%d cache entries are unreadable or name a blob that is gone, so they can never be a hit", s.BrokenEntries),
		Remedy: "run `smith cache verify --fix` to remove them",
	})
	return out
}

func size(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	}
	return fmt.Sprintf("%d KB", n>>10)
}

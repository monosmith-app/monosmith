package cachemaintenance

import "time"

const (
	// CheckInterval is how often, at most, a run walks the store to measure or
	// prune it.
	CheckInterval = time.Hour
	// BigRunBytes is how much one run must store to be checked sooner than
	// CheckInterval, when the store is likely over its limit.
	BigRunBytes = 4 << 20
	// RetryAfter is how soon the store is looked at again when a prune left it
	// over the limit because what it would remove was in use or just written.
	RetryAfter = InUseGrace + time.Minute
)

// shouldCheck decides whether a run that stored `added` bytes walks the store.
// It does not when there is no limit; it does when the store was never
// measured or was measured CheckInterval ago, which also corrects the running
// total for what other processes stored or removed; and between those, only
// when this run stored BigRunBytes or more and the running total says the
// store is over its limit. Most runs therefore read and write one small file.
func shouldCheck(st pruneState, now time.Time, added, limit int64) bool {
	if limit <= 0 {
		return false
	}
	if st.LastCheck.IsZero() || now.Sub(st.LastCheck) >= CheckInterval {
		return true
	}
	return added >= BigRunBytes && st.SizeBytes+st.AddedBytes+added > limit
}

package taskrun

import (
	"errors"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/features/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/features/taskcache"
)

// missOf says why a cached task found no entry to restore: it compares the
// inputs of its key now with those of its last cached run, as smith why-miss
// does, without computing anything again.
//
// Parameters: s is the run's storage; task is the task id; now are the inputs
// its key was computed from in this run.
// Returns: the reason, or nil when the storage cannot say (it failed to read).
func missOf(s taskcache.Storage, task string, now taskcache.KeyInputs) *runevents.Miss {
	last, _, err := taskcache.LastRun(s, task)
	if errors.Is(err, taskcache.ErrNoLastRun) {
		return &runevents.Miss{Reason: runevents.MissNoEntry}
	}
	if err != nil {
		return nil
	}
	if last.Inputs == nil {
		// The last run predates the records of inputs: only its key is known.
		return &runevents.Miss{Reason: runevents.MissKey}
	}
	return missFrom(taskcache.Diff(*last.Inputs, now))
}

// missFrom is the reason a list of differences gives: the files that changed
// counted, and each other kind of change named once, in the order of diffs.
// No difference at all, with a key that still differs, is a changed key.
func missFrom(diffs []taskcache.Difference) *runevents.Miss {
	if len(diffs) == 0 {
		return &runevents.Miss{Reason: runevents.MissKey}
	}
	m := &runevents.Miss{Reason: runevents.MissInputs}
	for _, d := range diffs {
		if d.What == "file" {
			m.Files++
			continue
		}
		if !slices.Contains(m.Changed, d.What) {
			m.Changed = append(m.Changed, d.What)
		}
	}
	return m
}

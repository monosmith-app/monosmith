package rundurations

import (
	"slices"
	"time"
)

// Keep is how many durations a task keeps: the newest ones.
const Keep = 5

// History is how long each task took in its last runs, by task id
// ("web:build"), in milliseconds, the oldest first, at most Keep each.
type History map[string][]int64

// Record adds a duration to a task and drops the oldest past Keep. A duration
// of zero or less says nothing and is left out.
//
// Parameters: task is the task id; d is how long it took.
// Returns: the history with the duration; h itself is not changed.
func (h History) Record(task string, d time.Duration) History {
	if d <= 0 || task == "" {
		return h
	}
	out := make(History, len(h)+1)
	for k, v := range h {
		out[k] = v
	}
	list := append(slices.Clone(h[task]), d.Milliseconds())
	if len(list) > Keep {
		list = list[len(list)-Keep:]
	}
	out[task] = list
	return out
}

// Typical is the median of a task's durations.
//
// Parameters: task is the task id.
// Returns: the median and true, or false when the task has no history.
func (h History) Typical(task string) (time.Duration, bool) {
	list := slices.Clone(h[task])
	if len(list) == 0 {
		return 0, false
	}
	slices.Sort(list)
	mid := list[len(list)/2]
	if len(list)%2 == 0 {
		mid = (list[len(list)/2-1] + mid) / 2
	}
	return time.Duration(mid) * time.Millisecond, true
}

package rundurations

import "time"

// Pending is a task of a run that has not ended: queued, or running for Elapsed.
type Pending struct {
	ID      string
	Running bool
	Elapsed time.Duration
}

// Estimate is about how long the pending tasks of a run still take: the typical
// time of each, less what a running one already spent, shared by the tasks that
// run at once. It is an estimate only when every pending task has a history.
//
// Parameters: h is the history; pending are the tasks that have not ended.
// Returns: the time left and true, or false when nothing is pending or a task
// has no history.
func Estimate(h History, pending []Pending) (time.Duration, bool) {
	if len(pending) == 0 {
		return 0, false
	}
	var work time.Duration
	running := 0
	for _, p := range pending {
		typical, ok := h.Typical(p.ID)
		if !ok {
			return 0, false
		}
		if p.Running {
			running++
			typical = max(typical-p.Elapsed, 0)
		}
		work += typical
	}
	return work / time.Duration(max(running, 1)), true
}

// Round makes an estimate read as one: whole seconds up to a minute, five
// seconds at least, and whole minutes past a minute.
//
// Parameters: d is the estimate.
// Returns: the rounded time.
func Round(d time.Duration) time.Duration {
	switch {
	case d < 5*time.Second:
		return 5 * time.Second
	case d < time.Minute:
		return d.Round(time.Second)
	}
	return d.Round(time.Minute)
}

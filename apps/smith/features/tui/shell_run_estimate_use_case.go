package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/rundurations"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runview"
)

// The time estimate of a run: how long its tasks took before, kept in the
// workspace's local cache, gives "about 20s left" while it runs.

// keepDurations adds the time of each task the run just ran and passed to the
// history, and keeps the history. A task the cache answered, one that failed
// and one that was stopped say nothing about how long the task takes.
func (s *Shell) keepDurations() {
	if s.run == nil || s.run.Cancelled {
		return
	}
	h, added := s.durations, false
	for _, t := range s.run.Tasks {
		if t.Status == runview.Passed && t.Duration > 0 {
			h, added = h.Record(t.ID, t.Duration), true
		}
	}
	if !added {
		return
	}
	s.durations = h
	if s.saveDurations != nil {
		s.saveDurations(h)
	}
}

// timeLeft is about how long the run shown still takes, rounded ("20s",
// "2m0s"), while it runs and every task left has a history.
func (s Shell) timeLeft() (string, bool) {
	if !s.running() || s.stopping {
		return "", false
	}
	now := s.now()
	var pending []rundurations.Pending
	for _, t := range s.run.Tasks {
		switch t.Status {
		case runview.Queued:
			pending = append(pending, rundurations.Pending{ID: t.ID})
		case runview.Running:
			pending = append(pending, rundurations.Pending{ID: t.ID, Running: true, Elapsed: s.run.Elapsed(t, now)})
		}
	}
	left, ok := rundurations.Estimate(s.durations, pending)
	if !ok {
		return "", false
	}
	return rundurations.Round(left).String(), true
}

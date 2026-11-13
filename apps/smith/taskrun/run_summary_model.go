package taskrun

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/lintfix"
	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

// Status is a task's outcome.
type Status string

// Outcomes.
const (
	Succeeded Status = "success"
	Cached    Status = "cached"
	Failed    Status = "failure"
	Skipped   Status = "skipped"
)

// Result is one task's outcome.
type Result struct {
	Task     string        `json:"task"`
	Status   Status        `json:"status"`
	Key      string        `json:"key,omitempty"`
	Duration time.Duration `json:"durationNs"`
	// Saved is the original duration of a task replayed from cache.
	Saved  time.Duration `json:"savedNs,omitempty"`
	Reason string        `json:"reason,omitempty"`
	// Source is the tier that served a cached task: "local" or "remote".
	// It is empty for a task that was not replayed.
	Source taskcache.Source `json:"source,omitempty"`
}

// Summary is the outcome of a run.
type Summary struct {
	Results  []Result      `json:"results"`
	Duration time.Duration `json:"durationNs"`
	// Fixes are the lint outcomes of a fixing run (smith lint --fix), one per
	// project; empty otherwise.
	Fixes []lintfix.Outcome `json:"fixes,omitempty"`
}

// OK reports whether every task succeeded or was cached.
func (s Summary) OK() bool {
	for _, r := range s.Results {
		if r.Status == Failed || r.Status == Skipped {
			return false
		}
	}
	return true
}

// Count returns how many results have a status.
func (s Summary) Count(st Status) int {
	n := 0
	for _, r := range s.Results {
		if r.Status == st {
			n++
		}
	}
	return n
}

// CountRemote returns how many cached results the remote cache served.
func (s Summary) CountRemote() int {
	n := 0
	for _, r := range s.Results {
		if r.Status == Cached && r.Source == taskcache.SourceRemote {
			n++
		}
	}
	return n
}

// String is the one-paragraph report printed after a run.
func (s Summary) String() string {
	var b strings.Builder
	results := append([]Result(nil), s.Results...)
	sort.Slice(results, func(i, j int) bool { return results[i].Task < results[j].Task })
	// The fix report closes the run unless another task failed: then the totals follow it.
	onlyLint := len(s.Fixes) > 0
	for _, r := range results {
		fixed := len(s.Fixes) > 0 && strings.HasSuffix(r.Task, ":"+taskgraph.LintTask)
		onlyLint = onlyLint && (fixed || (r.Status != Failed && r.Status != Skipped))
		if (r.Status == Failed || r.Status == Skipped) && !fixed {
			fmt.Fprintf(&b, "  %s %s %s\n", mark(r.Status), r.Task, r.Reason)
		}
	}
	if len(s.Fixes) > 0 {
		// The fix report has its own closing line with the totals and the time.
		text, _ := lintfix.Report(s.Fixes, s.Duration)
		b.WriteString(text)
		if onlyLint {
			return strings.TrimRight(b.String(), "\n")
		}
	}
	var saved time.Duration
	for _, r := range s.Results {
		saved += r.Saved
	}
	done := s.Count(Succeeded) + s.Count(Cached)
	fmt.Fprintf(&b, "%s %d/%d tasks succeeded (%d from cache", mark(map[bool]Status{true: Succeeded, false: Failed}[s.OK()]), done, len(s.Results), s.Count(Cached))
	if remote := s.CountRemote(); remote > 0 {
		fmt.Fprintf(&b, ", %d of them remote", remote)
	}
	if saved > 0 {
		fmt.Fprintf(&b, ", %s saved", saved.Round(time.Millisecond))
	}
	fmt.Fprintf(&b, ") in %s", s.Duration.Round(time.Millisecond))
	return b.String()
}

func mark(s Status) string {
	switch s {
	case Failed:
		return marks.Failed
	case Skipped:
		return marks.Skipped
	default:
		return marks.Passed
	}
}

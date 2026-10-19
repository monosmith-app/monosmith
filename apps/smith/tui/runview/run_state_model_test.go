package runview

import (
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

func feed(r *Run, at time.Duration, lines ...string) {
	for _, l := range lines {
		r.Feed(l, t0.Add(at))
	}
}

func TestPlannedTasksStartQueuedAndGiveTheTotal(t *testing.T) {
	r := NewRun([]string{"test"}, []string{"kit:test", "web:test", "sdk:test"}, t0)
	c := r.Counts()
	if c.Queued != 3 || c.Total != 3 || c.Done != 0 {
		t.Fatalf("%+v", c)
	}
}

func TestARunWalksTheTasksThroughTheirStates(t *testing.T) {
	r := NewRun([]string{"test"}, []string{"kit:test", "web:test", "sdk:test"}, t0)
	feed(r, 0, "kit:test ▶ go test ./...", "web:test ▶ npm run test")
	c := r.Counts()
	if c.Running != 2 || c.Queued != 1 {
		t.Fatalf("%+v", c)
	}
	feed(r, time.Second, "kit:test │ ok example.com/kit 0.3s", "kit:test ✔ 934ms", "sdk:test ✔ from cache")
	c = r.Counts()
	if c.Passed != 1 || c.Cached != 1 || c.Running != 1 || c.Done != 2 {
		t.Fatalf("%+v", c)
	}
	if got := r.Task("kit:test").Log; len(got) != 1 || got[0] != "ok example.com/kit 0.3s" {
		t.Fatalf("kit log %v", got)
	}
	feed(r, 2*time.Second, "web:test ✔ 1.5s", "✔ 3/3 tasks succeeded (1 from cache) in 2s")
	r.End(0, false, t0.Add(2*time.Second))
	if !r.OK() || r.Counts().Done != 3 {
		t.Fatalf("ok %v %+v", r.OK(), r.Counts())
	}
	if r.RunTime(t0.Add(time.Hour)) != 2*time.Second {
		t.Fatalf("a finished run stops its clock: %v", r.RunTime(t0.Add(time.Hour)))
	}
}

func TestFailuresComeFirstThenRunningThenQueued(t *testing.T) {
	r := NewRun([]string{"build"}, []string{"a:build", "b:build", "c:build", "d:build"}, t0)
	feed(r, 0, "a:build ▶ x", "b:build ▶ x", "a:build ✔ 1s", "b:build ✖ exited with code 1")
	var ids []string
	for _, tk := range r.Ordered() {
		ids = append(ids, tk.ID+":"+tk.Status.String())
	}
	want := []string{"b:build:failed", "c:build:queued", "d:build:queued", "a:build:passed"}
	if len(ids) != len(want) {
		t.Fatalf("%v", ids)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("got %v want %v", ids, want)
		}
	}
	if r.Task("b:build").Reason != "exited with code 1" {
		t.Fatalf("reason %q", r.Task("b:build").Reason)
	}
}

func TestTheClosingSummaryMarksSkippedTasks(t *testing.T) {
	r := NewRun([]string{"test"}, []string{"kit:test", "web:test"}, t0)
	feed(r, 0, "kit:test ✖ exited with code 2", "  ✖ kit:test exited with code 2", "  ↷ web:test because kit:test did not succeed", "✖ 0/2 tasks succeeded (0 from cache) in 1s")
	r.End(1, false, t0.Add(time.Second))
	if r.Task("web:test").Status != Skipped || r.Task("web:test").Reason != "because kit:test did not succeed" {
		t.Fatalf("%+v", r.Task("web:test"))
	}
	if r.OK() || r.Counts().Failed != 1 || r.Counts().Skipped != 1 {
		t.Fatalf("%+v", r.Counts())
	}
}

func TestCancellingLeavesNothingSpinning(t *testing.T) {
	r := NewRun([]string{"build"}, []string{"a:build", "b:build"}, t0)
	feed(r, 0, "a:build ▶ x")
	r.End(1, true, t0.Add(3*time.Second))
	if r.Task("a:build").Status != Failed || r.Task("a:build").Reason != "interrupted" {
		t.Fatalf("%+v", r.Task("a:build"))
	}
	if r.Task("b:build").Status != Skipped || r.Task("b:build").Reason != "cancelled" {
		t.Fatalf("%+v", r.Task("b:build"))
	}
	if r.OK() {
		t.Fatal("a cancelled run is not ok")
	}
}

func TestTasksNotInThePlanAreAddedAsTheyAppear(t *testing.T) {
	r := NewRun([]string{"build"}, nil, t0)
	feed(r, 0, "affected: 1 file → 1 project(s)", "kit:build ▶ go build", "kit:build ✔ 1s")
	if c := r.Counts(); c.Total != 1 || c.Passed != 1 {
		t.Fatalf("%+v", c)
	}
	if len(r.Lines) != 3 {
		t.Fatalf("every line is kept: %d", len(r.Lines))
	}
}

func TestElapsedCountsUpWhileRunning(t *testing.T) {
	r := NewRun([]string{"build"}, nil, t0)
	feed(r, time.Second, "a:build ▶ x")
	if got := r.Elapsed(r.Task("a:build"), t0.Add(4*time.Second)); got != 3*time.Second {
		t.Fatalf("elapsed %v", got)
	}
}

func TestTheLogsAreCapped(t *testing.T) {
	r := NewRun([]string{"build"}, nil, t0)
	for i := 0; i < maxTaskLines+50; i++ {
		r.Feed("a:build │ line", t0)
	}
	if len(r.Task("a:build").Log) != maxTaskLines {
		t.Fatalf("%d lines", len(r.Task("a:build").Log))
	}
}

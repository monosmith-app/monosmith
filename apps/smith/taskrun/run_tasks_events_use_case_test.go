package taskrun

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

type collected struct {
	mu     sync.Mutex
	events []runevents.Event
}

func (c *collected) Emit(e runevents.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, e)
}

func (c *collected) kinds() []string {
	var out []string
	for _, e := range c.events {
		out = append(out, fmt.Sprintf("%s %s %s", e.Event, e.Task, e.Status))
	}
	return out
}

func (f *fixture) runWithEvents(t *testing.T, tasks ...string) (Summary, *collected) {
	t.Helper()
	g, err := taskgraph.Build(f.pg, tasks, []string{"build"})
	if err != nil {
		t.Fatal(err)
	}
	got := &collected{}
	s := Run(context.Background(), g, Options{Root: f.root, Files: f.files, Projects: f.pg, Storage: f.store, Parallel: 2,
		Out: &bytes.Buffer{}, ToolVersion: "test", Command: "smith run build", Events: runevents.NewRunStream("r1", got)})
	return s, got
}

func TestARunEmitsItsEventsInOrder(t *testing.T) {
	f := newFixture(t, "core")
	s, got := f.runWithEvents(t, "web")
	if !s.OK() {
		t.Fatalf("%s", s)
	}
	want := []string{
		"run.started  ", "task.queued core:build ", "task.queued web:build ",
		"task.started core:build ", "task.output core:build ", "task.finished core:build passed",
		"task.started web:build ", "task.output web:build ", "task.finished web:build passed",
		"run.finished  ",
	}
	if fmt.Sprint(got.kinds()) != fmt.Sprint(want) {
		t.Fatalf("events\n got %q\nwant %q", got.kinds(), want)
	}
	for i, e := range got.events {
		if e.Run != "r1" || e.Seq != i+1 || e.Time.IsZero() {
			t.Fatalf("event %d is not stamped: %+v", i, e)
		}
	}
	if got.events[0].Command != "smith run build" || len(got.events[0].Tasks) != 2 {
		t.Fatalf("run.started %+v", got.events[0])
	}
	if out := got.events[4]; out.Line != "helper ran with core" || out.Stream != runevents.Combined {
		t.Fatalf("task.output %+v", out)
	}
	last := got.events[len(got.events)-1]
	if last.OK == nil || !*last.OK || last.Summary.Passed != 2 {
		t.Fatalf("run.finished %+v", last)
	}
	if got.events[5].CacheKey == "" || got.events[5].CacheKey != s.Results[0].Key {
		t.Fatalf("task.finished carries the cache key: %+v", got.events[5])
	}
}

func TestACachedTaskAndASkippedTaskAreReportedAsSuch(t *testing.T) {
	f := newFixture(t, "core")
	f.runWithEvents(t, "web")
	_ = os.RemoveAll(filepath.Join(f.root, "apps", "web", "out"))
	_, got := f.runWithEvents(t, "web")
	statuses := map[string]runevents.Status{}
	for _, e := range got.events {
		if e.Event == runevents.TaskFinished {
			statuses[e.Task] = e.Status
		}
	}
	if statuses["core:build"] != runevents.Cached || statuses["web:build"] != runevents.Cached {
		t.Fatalf("second run %v", statuses)
	}

	failing := newFixture(t, "fail")
	s, got := failing.runWithEvents(t, "web")
	if s.OK() {
		t.Fatal("the run must fail")
	}
	statuses = map[string]runevents.Status{}
	for _, e := range got.events {
		if e.Event == runevents.TaskFinished {
			statuses[e.Task] = e.Status
		}
	}
	if statuses["core:build"] != runevents.Failed || statuses["web:build"] != runevents.Skipped {
		t.Fatalf("failing run %v", statuses)
	}
	if last := got.events[len(got.events)-1]; last.OK == nil || *last.OK || last.Summary.Failed != 1 || last.Summary.Skipped != 1 {
		t.Fatalf("run.finished %+v", last)
	}
}

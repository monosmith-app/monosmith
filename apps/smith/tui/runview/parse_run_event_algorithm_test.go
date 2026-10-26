package runview

import (
	"reflect"
	"testing"
	"time"
)

func TestParseLineReadsTheEventStream(t *testing.T) {
	cases := []struct {
		line string
		want Event
	}{
		{`{"event":"run.started","run":"r","seq":1,"time":"2026-10-25T10:00:00Z","command":"run build","tasks":["a:build","b:build"]}`, Event{Kind: Plan, Tasks: []string{"a:build", "b:build"}, Hidden: true}},
		{`{"event":"task.queued","run":"r","seq":2,"time":"2026-10-25T10:00:00Z","task":"a:build"}`, Event{Kind: Queue, Task: "a:build", Hidden: true}},
		{`{"event":"task.started","run":"r","seq":3,"time":"2026-10-25T10:00:00Z","task":"a:build"}`, Event{Kind: Started, Task: "a:build", Display: "a:build ▶ started"}},
		{`{"event":"task.output","run":"r","seq":4,"time":"2026-10-25T10:00:00Z","task":"a:build","stream":"combined","line":"a:build ✖ not a status"}`, Event{Kind: Output, Task: "a:build", Text: "a:build ✖ not a status", Display: "a:build │ a:build ✖ not a status"}},
		{`{"event":"task.finished","run":"r","seq":5,"time":"2026-10-25T10:00:00Z","task":"a:build","status":"passed","durationMs":106}`, Event{Kind: Finish, Task: "a:build", Status: Passed, Duration: 106 * time.Millisecond, Text: "106ms", Display: "a:build ✔ 106ms"}},
		{`{"event":"task.finished","run":"r","seq":5,"time":"2026-10-25T10:00:00Z","task":"a:build","status":"cached"}`, Event{Kind: Finish, Task: "a:build", Status: Cached, Text: "from cache", Display: "a:build ✔ from cache"}},
		{`{"event":"task.finished","run":"r","seq":5,"time":"2026-10-25T10:00:00Z","task":"a:build","status":"failed","durationMs":20,"reason":"exited with code 2"}`, Event{Kind: Finish, Task: "a:build", Status: Failed, Duration: 20 * time.Millisecond, Text: "exited with code 2", Display: "a:build ✖ exited with code 2"}},
		{`{"event":"task.finished","run":"r","seq":5,"time":"2026-10-25T10:00:00Z","task":"b:build","status":"skipped","reason":"because a:build did not succeed"}`, Event{Kind: Skip, Task: "b:build", Status: Skipped, Text: "because a:build did not succeed", Display: "  ↷ b:build because a:build did not succeed"}},
		{`{"event":"run.finished","run":"r","seq":9,"time":"2026-10-25T10:00:01Z","ok":false,"summary":{"passed":1,"failed":1,"cached":1,"skipped":1,"durationMs":1200}}`, Event{Kind: Summary, Done: 2, Total: 4, Text: "✖ 2/4 tasks succeeded (1 from cache) in 1.2s", Display: "✖ 2/4 tasks succeeded (1 from cache) in 1.2s"}},
	}
	for _, c := range cases {
		got := ParseLine(c.line)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s\n got %+v\nwant %+v", c.line, got, c.want)
		}
	}
	// A line that only looks like an event is text.
	for _, line := range []string{`{"event":"nope"`, `{"event":"mystery"}`, `{"other":1}`} {
		if got := ParseLine(line); got.Kind != Plain {
			t.Errorf("%s parsed as %+v", line, got)
		}
	}
}

func TestAnEventStreamBuildsTheStateOfARun(t *testing.T) {
	now := time.Now()
	r := NewRun([]string{"run", "build"}, nil, now)
	for _, l := range []string{
		`{"event":"run.started","run":"r","seq":1,"time":"2026-10-25T10:00:00Z","tasks":["a:build","b:build","c:build"]}`,
		`{"event":"task.queued","run":"r","seq":2,"time":"2026-10-25T10:00:00Z","task":"a:build"}`,
		`{"event":"task.started","run":"r","seq":3,"time":"2026-10-25T10:00:00Z","task":"a:build"}`,
		`{"event":"task.output","run":"r","seq":4,"time":"2026-10-25T10:00:00Z","task":"a:build","stream":"combined","line":"boom"}`,
		`{"event":"task.finished","run":"r","seq":5,"time":"2026-10-25T10:00:00Z","task":"a:build","status":"failed","reason":"exited with code 2"}`,
		`{"event":"task.finished","run":"r","seq":6,"time":"2026-10-25T10:00:00Z","task":"b:build","status":"skipped","reason":"because a:build did not succeed"}`,
		"smith: something printed outside the stream",
	} {
		r.Feed(l, now)
	}
	c := r.Counts()
	if c.Total != 3 || c.Failed != 1 || c.Skipped != 1 || c.Queued != 1 {
		t.Fatalf("counts %+v", c)
	}
	if got := r.Task("a:build").Log; len(got) != 1 || got[0] != "boom" {
		t.Fatalf("log %v", got)
	}
	// The log reads as the text format does, and the plan is not a line of it.
	want := []string{"a:build ▶ started", "a:build │ boom", "a:build ✖ exited with code 2", "  ↷ b:build because a:build did not succeed", "smith: something printed outside the stream"}
	if !reflect.DeepEqual(r.Lines, want) {
		t.Fatalf("lines %q", r.Lines)
	}
}

package runview

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
)

// parseEventLine reads a line of the event stream `smith run --json-events`
// prints (docs/reference/serve.md): one JSON object per line. ok is false for
// any other line, which is text (smith's own messages, or a failure before the
// run started, both go to the same pipe). The events are the structured form
// of what the text format says, so a task's output can never be mistaken for
// its status; each carries the line the text format would show, for the log.
func parseEventLine(line string) (Event, bool) {
	if !strings.HasPrefix(line, `{"event":"`) {
		return Event{}, false
	}
	var e runevents.Event
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		return Event{}, false
	}
	dur := time.Duration(e.DurationMs) * time.Millisecond
	switch e.Event {
	case runevents.RunStarted:
		return Event{Kind: Plan, Tasks: e.Tasks, Hidden: true}, true
	case runevents.TaskQueued:
		return Event{Kind: Queue, Task: e.Task, Hidden: true}, true
	case runevents.TaskStarted:
		return Event{Kind: Started, Task: e.Task, Display: e.Task + " ▶ started"}, true
	case runevents.TaskOutput:
		return Event{Kind: Output, Task: e.Task, Text: e.Line, Display: e.Task + " │ " + e.Line}, true
	case runevents.TaskFinished:
		switch e.Status {
		case runevents.Failed:
			return Event{Kind: Finish, Task: e.Task, Status: Failed, Text: e.Reason, Duration: dur, Display: strings.TrimSpace(e.Task + " ✖ " + e.Reason)}, true
		case runevents.Cached:
			return Event{Kind: Finish, Task: e.Task, Status: Cached, Text: fromCache, Display: e.Task + " ✔ " + fromCache}, true
		case runevents.Skipped:
			return Event{Kind: Skip, Task: e.Task, Status: Skipped, Text: e.Reason, Display: skipLine(e.Task, e.Reason)}, true
		}
		return Event{Kind: Finish, Task: e.Task, Status: Passed, Duration: dur, Text: dur.String(), Display: e.Task + " ✔ " + dur.String()}, true
	case runevents.RunFinished:
		var s runevents.Summary
		if e.Summary != nil {
			s = *e.Summary
		}
		done, total := s.Passed+s.Cached, s.Passed+s.Cached+s.Failed+s.Skipped
		ok := e.OK != nil && *e.OK
		mark := "✖"
		if ok {
			mark = "✔"
		}
		took := (time.Duration(s.DurationMs) * time.Millisecond).String()
		text := fmt.Sprintf("%s %d/%d tasks succeeded (%d from cache) in %s", mark, done, total, s.Cached, took)
		return Event{Kind: Summary, Done: done, Total: total, OK: ok, Text: text, Display: text}, true
	}
	return Event{}, false
}

// skipLine is the line of the closing summary that tells a task was skipped.
func skipLine(task, reason string) string {
	if reason == "" {
		return "  ↷ " + task
	}
	return "  ↷ " + task + " " + reason
}

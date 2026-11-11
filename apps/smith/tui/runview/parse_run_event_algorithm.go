package runview

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/monosmith-app/monosmith/libs/words/words"
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
		return Event{Kind: Started, Task: e.Task, Display: e.Task + words.Say("runview.started")}, true
	case runevents.TaskOutput:
		// styled is the line with its colour, sent when the TUI handed tool
		// colour on; the pane shows it, and measures it without the codes.
		text := e.Line
		if e.Styled != "" {
			text = e.Styled
		}
		return Event{Kind: Output, Task: e.Task, Text: text, Display: e.Task + " │ " + text}, true
	case runevents.TaskFinished:
		switch e.Status {
		case runevents.Failed:
			return Event{Kind: Finish, Task: e.Task, Status: Failed, Text: e.Reason, Duration: dur, Display: strings.TrimSpace(e.Task + " × " + e.Reason)}, true
		case runevents.Cached:
			label := fromCache
			if e.Source == "remote" {
				label = fromRemoteCache
			}
			return Event{Kind: Finish, Task: e.Task, Status: Cached, Text: label, Remote: label == fromRemoteCache, Display: e.Task + " ✓ " + label}, true
		case runevents.Skipped:
			return Event{Kind: Skip, Task: e.Task, Status: Skipped, Text: e.Reason, Display: skipLine(e.Task, e.Reason)}, true
		}
		return Event{Kind: Finish, Task: e.Task, Status: Passed, Duration: dur, Text: dur.String(), Display: e.Task + " ✓ " + dur.String()}, true
	case runevents.RunFinished:
		var s runevents.Summary
		if e.Summary != nil {
			s = *e.Summary
		}
		done, total := s.Passed+s.Cached, s.Passed+s.Cached+s.Failed+s.Skipped
		ok := e.OK != nil && *e.OK
		mark := marks.Failed
		if ok {
			mark = marks.Passed
		}
		took := (time.Duration(s.DurationMs) * time.Millisecond).String()
		text := words.Say("runview.summary", "mark", mark, "done", done, "total", total, "cached", s.Cached, "took", took)
		if s.CachedRemote > 0 {
			text = words.Say("runview.summary_remote", "mark", mark, "done", done, "total", total, "cached", s.Cached, "remote", s.CachedRemote, "took", took)
		}
		return Event{Kind: Summary, Done: done, Total: total, OK: ok, Text: text, Display: text}, true
	}
	return Event{}, false
}

// skipLine is the line of the closing summary that tells a task was skipped.
func skipLine(task, reason string) string {
	if reason == "" {
		return "  » " + task
	}
	return "  » " + task + " " + reason
}

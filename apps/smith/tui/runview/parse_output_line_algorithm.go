package runview

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// taskLine is "<project>:<task> <marker> <rest>" as taskrun prints it.
	taskLine = regexp.MustCompile(`^([^\s│]+:[^\s│]+) (│|▶|✔|✖|■|!)(?: (.*))?$`)
	// summaryTask is a line of the closing summary: "  ↷ web:test because ..."
	// or "  ✖ web:build exited with code 1".
	summaryTask = regexp.MustCompile(`^\s+(↷|✖) (\S+:\S+)(?: (.*))?$`)
	// summaryLine is "✔ 3/4 tasks succeeded (1 from cache) in 1.2s".
	summaryLine = regexp.MustCompile(`^(✔|✖) (\d+)/(\d+) tasks succeeded`)
)

// fromCache is how smith says a task was replayed.
const fromCache = "from cache"

// ParseLine reads one line of `smith run` output (without its newline): a line
// of the event stream (`--json-events`, see parseEventLine) or of the text.
func ParseLine(line string) Event {
	line = strings.TrimRight(line, "\r\n")
	if ev, ok := parseEventLine(line); ok {
		return ev
	}
	if m := taskLine.FindStringSubmatch(line); m != nil {
		task, marker, rest := m[1], m[2], m[3]
		switch marker {
		case "│":
			return Event{Kind: Output, Task: task, Text: rest}
		case "▶":
			return Event{Kind: Started, Task: task, Text: rest}
		case "■":
			return Event{Kind: Stopped, Task: task, Text: rest}
		case "!":
			return Event{Kind: Warned, Task: task, Text: rest}
		case "✖":
			return Event{Kind: Finish, Task: task, Status: Failed, Text: rest}
		case "✔":
			if rest == fromCache {
				return Event{Kind: Finish, Task: task, Status: Cached, Text: rest}
			}
			d, _ := time.ParseDuration(rest)
			return Event{Kind: Finish, Task: task, Status: Passed, Duration: d, Text: rest}
		}
	}
	if m := summaryLine.FindStringSubmatch(line); m != nil {
		done, _ := strconv.Atoi(m[2])
		total, _ := strconv.Atoi(m[3])
		return Event{Kind: Summary, Done: done, Total: total, OK: m[1] == "✔", Text: line}
	}
	if m := summaryTask.FindStringSubmatch(line); m != nil {
		if m[1] == "↷" {
			return Event{Kind: Skip, Task: m[2], Status: Skipped, Text: m[3]}
		}
		return Event{Kind: Failure, Task: m[2], Status: Failed, Text: m[3]}
	}
	return Event{Kind: Plain, Text: line}
}

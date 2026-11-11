package runview

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

var (
	// taskLine is "<project>:<task> <marker> <rest>" as taskrun prints it.
	taskLine = regexp.MustCompile(`^([^\s│]+:[^\s│]+) (│|▶|✓|×|■|!)(?: (.*))?$`) // words:literal
	// summaryTask is a line of the closing summary: "  » web:test because ..."
	// or "  × web:build exited with code 1".
	summaryTask = regexp.MustCompile(`^\s+(»|×) (\S+:\S+)(?: (.*))?$`) // words:literal
	// summaryLine is "✓ 3/4 tasks succeeded (1 from cache) in 1.2s".
	summaryLine = regexp.MustCompile(`^(✓|×) (\d+)/(\d+) tasks succeeded`) // words:literal
)

// fromCache is how smith says a task was replayed.
const fromCache = "from cache" // words:literal the text of `smith run` that the parser matches

// ParseLine reads one line of `smith run` output (without its newline): a line
// of the event stream (`--json-events`, see parseEventLine) or of the text.
func ParseLine(line string) Event {
	line = strings.TrimRight(line, "\r\n")
	if ev, ok := parseEventLine(line); ok {
		return ev
	}
	// A tool's line may carry colour: the markers are matched on the plain
	// text, and a task's output keeps its colour for the pane.
	plain := termcolour.Strip(line)
	if m := taskLine.FindStringSubmatch(plain); m != nil {
		task, marker, rest := m[1], m[2], m[3]
		switch marker {
		case "│":
			if _, coloured, ok := strings.Cut(line, " │ "); ok {
				rest = coloured
			}
			return Event{Kind: Output, Task: task, Text: rest}
		case "▶":
			return Event{Kind: Started, Task: task, Text: rest}
		case "■":
			return Event{Kind: Stopped, Task: task, Text: rest}
		case "!":
			return Event{Kind: Warned, Task: task, Text: rest}
		case marks.Failed:
			return Event{Kind: Finish, Task: task, Status: Failed, Text: rest}
		case marks.Passed:
			if rest == fromCache {
				return Event{Kind: Finish, Task: task, Status: Cached, Text: rest}
			}
			d, _ := time.ParseDuration(rest)
			return Event{Kind: Finish, Task: task, Status: Passed, Duration: d, Text: rest}
		}
	}
	if m := summaryLine.FindStringSubmatch(plain); m != nil {
		done, _ := strconv.Atoi(m[2])
		total, _ := strconv.Atoi(m[3])
		return Event{Kind: Summary, Done: done, Total: total, OK: m[1] == marks.Passed, Text: plain}
	}
	if m := summaryTask.FindStringSubmatch(plain); m != nil {
		if m[1] == marks.Skipped {
			return Event{Kind: Skip, Task: m[2], Status: Skipped, Text: m[3]}
		}
		return Event{Kind: Failure, Task: m[2], Status: Failed, Text: m[3]}
	}
	return Event{Kind: Plain, Text: line}
}

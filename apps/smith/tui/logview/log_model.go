package logview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Log is the scroll state of a log. The zero value follows the end.
type Log struct {
	top int
	// paused is true once the user scrolled away from the end.
	paused bool
}

// Following reports whether the view sticks to the newest line.
func (l Log) Following() bool { return !l.paused }

// Top is the index of the first visible line for total lines in h rows.
func (l Log) Top(total, h int) int {
	last := max(total-h, 0)
	if !l.paused {
		return last
	}
	return min(max(l.top, 0), last)
}

// Scroll moves by delta lines (negative is up). Scrolling up pauses the log;
// scrolling down to the end follows it again.
func (l Log) Scroll(delta, total, h int) Log {
	top := min(max(l.Top(total, h)+delta, 0), max(total-h, 0))
	l.top = top
	l.paused = top < max(total-h, 0)
	return l
}

// Page scrolls by a page in direction dir (-1 up, 1 down).
func (l Log) Page(dir, total, h int) Log { return l.Scroll(dir*max(h-1, 1), total, h) }

// Home goes to the first line, pausing; End goes to the last and follows.
func (l Log) Home() Log { return Log{top: 0, paused: true} }
func (l Log) End() Log  { return Log{} }

// Toggle pauses a following log where it is, or resumes following.
func (l Log) Toggle(total, h int) Log {
	if l.paused {
		return Log{}
	}
	return Log{top: l.Top(total, h), paused: true}
}

// Render draws h rows of w cells from lines. A short log is padded with blank rows.
func (l Log) Render(th theme.Theme, lines []string, w, h int) []string {
	top := l.Top(len(lines), h)
	out := make([]string, 0, h)
	for i := top; i < len(lines) && len(out) < h; i++ {
		out = append(out, pane.Fit(Style(th, lines[i]), w))
	}
	return pane.Block(out, w, h)
}

// Style colours a line of smith's output by what it says: a task that passed,
// failed, was cached or started; task output keeps its own text, with the
// "<task> │" prefix dimmed.
func Style(th theme.Theme, line string) string {
	ev := runview.ParseLine(line)
	switch ev.Kind {
	case runview.Output:
		if id, text, ok := strings.Cut(line, " │ "); ok {
			return th.Muted(id+" │ ") + text
		}
	case runview.Started:
		return th.Info(line)
	case runview.Finish:
		switch ev.Status {
		case runview.Failed:
			return th.Failure(line)
		default:
			return th.Success(line)
		}
	case runview.Failure:
		return th.Failure(line)
	case runview.Skip, runview.Warned:
		return th.Warning(line)
	case runview.Summary:
		if ev.OK {
			return th.Success(line)
		}
		return th.Failure(line)
	}
	return line
}

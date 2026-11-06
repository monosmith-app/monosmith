package resultdialog

import (
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// outputLines is how many lines of a running command the dialog shows.
const outputLines = 8

// RunState says where a command that runs in the dialog is.
type RunState int

// The states of a command that runs in the dialog.
const (
	// Running: the command has not ended.
	Running RunState = iota
	// Succeeded: the command ended with code 0.
	Succeeded
	// Failed: the command ended with another code, or was stopped.
	Failed
)

// NewRunOutput builds the dialog that shows a command that runs where there is
// no Run tab (smith new and smith convert, before a workspace exists). It shows
// the end of the output; when the command is done it offers to open the
// workspace, or to go back to the form.
//
// Parameters: argv is the smith command line that runs; lines is its output so
// far; state says whether it runs, worked or failed.
// Returns: the dialog. Run while it runs, it offers only to stop the command.
func NewRunOutput(argv, lines []string, state RunState) Model {
	op := OperationText(argv)
	m := Model{}
	switch state {
	case Running:
		m.title, m.working = runningWords(argv, op)
	case Succeeded:
		m.title = words.Say("resultdialog.done", "operation", op)
	default:
		m.title = words.Say("resultdialog.run.failed", "operation", op)
	}
	if len(lines) > outputLines {
		lines = lines[len(lines)-outputLines:]
	}
	for _, l := range lines {
		if l = strings.TrimRight(l, " \t\r"); l != "" {
			m.facts = append(m.facts, Fact{Info, l})
		}
	}
	switch state {
	case Running:
		m.choices = []choice{{words.Say("resultdialog.run.stop.label"), words.Say("resultdialog.run.stop.hint"), Outcome{Stop: true}}}
		m.foot = words.Say("resultdialog.foot.running")
	case Succeeded:
		m.facts = append(m.facts, Fact{Good, words.Say("resultdialog.run.finished")})
		m.choices = []choice{
			{words.Say("resultdialog.run.enter.label"), words.Say("resultdialog.run.enter.hint"), Outcome{Enter: OutsideDir(argv)}},
			{words.Say("resultdialog.run.stay.label"), words.Say("resultdialog.run.stay.hint"), Outcome{Done: true}},
		}
		m.foot = words.Say("resultdialog.foot.finished")
	default:
		m.facts = append(m.facts, Fact{Bad, words.Say("resultdialog.run.stopped")})
		m.choices = []choice{{words.Say("resultdialog.run.back.label"), words.Say("resultdialog.run.back.hint"), Outcome{Done: true}}}
		m.foot = words.Say("resultdialog.foot.failed")
	}
	return m
}

// runningWords names what a command that runs in the dialog is doing: the
// title, and the sentence next to the spinner.
func runningWords(argv []string, op string) (title, working string) {
	switch {
	case len(argv) > 0 && argv[0] == "convert":
		return words.Say("resultdialog.convert.title"), words.Say("resultdialog.convert.working")
	case len(argv) > 0 && argv[0] == "new":
		return words.Say("resultdialog.new.title"), words.Say("resultdialog.new.working")
	}
	return words.Say("resultdialog.running.title", "operation", op), words.Say("resultdialog.running.working", "operation", op)
}

// Pulse shows that a running command is alive: a spinner frame and the time it
// has run, in the line above its output. It changes nothing once the command
// has ended.
//
// Parameters: tick is the animation frame; elapsed is how long the command ran.
// Returns: the dialog with the moving line.
func (m Model) Pulse(tick int, elapsed time.Duration) Model {
	m.tick, m.elapsed = tick, elapsed
	return m
}

// workingLine is the spinner line of a running command, or "" when the command
// is not running.
func (m Model) workingLine(th theme.Theme) string {
	if m.working == "" {
		return ""
	}
	return th.Accent(th.Spin(m.tick)) + " " + progress.Waiting(m.working+"…", m.elapsed)
}

// OutsideDir is the folder the shell opens after smith new or smith convert
// made a workspace: the folder of the name, the folder named by --dir, or the
// current folder (--here, or a convert, which works in place).
//
// Parameters: argv is the smith command line, without the leading "smith".
// Returns: a path relative to the current folder.
func OutsideDir(argv []string) string {
	if len(argv) == 0 || argv[0] != "new" {
		return "."
	}
	if slices.Contains(argv, "--here") {
		return "."
	}
	for i, a := range argv {
		if a == "--dir" && i+1 < len(argv) {
			return argv[i+1]
		}
	}
	for _, a := range argv[1:] {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return "."
}

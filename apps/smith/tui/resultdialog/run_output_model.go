package resultdialog

import (
	"slices"
	"strings"
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
		m.title = "Running: " + op
	case Succeeded:
		m.title = "Done: " + op
	default:
		m.title = "It did not finish: " + op
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
		m.choices = []choice{{"Stop it", "the command ends; files it already wrote stay", Outcome{Stop: true}}}
		m.foot = "smith shows its output here until it ends."
	case Succeeded:
		m.facts = append(m.facts, Fact{Good, "It finished."})
		m.choices = []choice{
			{"Open the workspace here", "smith starts again inside the folder (default)", Outcome{Enter: OutsideDir(argv)}},
			{"Stay here", "nothing more happens", Outcome{Done: true}},
		}
		m.foot = "Nothing is committed unless the command said so above."
	default:
		m.facts = append(m.facts, Fact{Bad, "It stopped before the end. The last lines above say why."})
		m.choices = []choice{{"Back", "closes this box; the form keeps your answers", Outcome{Done: true}}}
		m.foot = "Fix what the lines say, then try again."
	}
	return m
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

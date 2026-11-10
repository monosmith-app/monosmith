package tui

import (
	"encoding/json"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toolsdialog"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// capTools is the purpose of the capture that reads the install choices.
const capTools = "tools-plan"

// toolsStartedMsg says an install the dialog asked for started, or why not.
type toolsStartedMsg struct {
	proc runner.Handle
	err  error
}

// toolsOutputMsg is output of the install that streams in the dialog.
type toolsOutputMsg struct {
	proc  runner.Handle
	lines []string
	exit  *runner.Exit
}

// isToolsInstall reports whether a command line is a plain `tools install`
// (tool names only): the dialog asks the same questions the command asks in a
// terminal. A command with a flag that answers for the person (--yes,
// --manager, --dry-run) runs as it is.
func isToolsInstall(argv []string) bool {
	if len(argv) < 2 || argv[0] != "tools" || argv[1] != "install" {
		return false
	}
	return !slices.ContainsFunc(argv[2:], func(a string) bool { return strings.HasPrefix(a, "-") })
}

// openTools reads the install choices of the missing tools. The dialog opens
// when they arrive; without a way to read them the command hands over to the
// terminal, as before.
func (s *Shell) openTools(argv []string) tea.Cmd {
	if s.sources.Capture == nil || s.sources.Start == nil || !s.inWorkspace {
		s.Argv, s.quit = argv, true
		return nil
	}
	s.toolsArgv = argv
	plan := append(slices.Clone(argv), "--dry-run", "--json")
	return s.captureWorking(capTools, words.Say("tui.tools.install_the_missing_tools"), words.Say("tui.tools.smith_looks_for_the_tools_this"), plan, plan, words.Say("tui.tools.looking_for_the_tools_that_are"), words.Say("tui.tools.cancelled_no_tool_was_installed"))
}

// toolsRead opens the dialog on the plans the command printed. When they
// cannot be read, the terminal hand-over stays the way.
func (s *Shell) toolsRead(msg capturedMsg) tea.Cmd {
	var plans []toolinstall.ToolPlan
	if !msg.res.OK || json.Unmarshal([]byte(msg.res.Text), &plans) != nil {
		s.notify(words.Say("tui.tools.could_not_read_the_tools_here"), toast.Warning)
		s.Argv, s.quit = s.toolsArgv, true
		return nil
	}
	m := toolsdialog.New(plans).Place(s.w, s.h)
	s.tools = &m
	return nil
}

// toolsKey sends a key to the dialog and does what it asks.
func (s *Shell) toolsKey(name string) tea.Cmd {
	m, out := s.tools.Key(name)
	s.tools = &m
	return s.afterTools(out)
}

// afterTools does what the tools dialog asked for.
func (s *Shell) afterTools(out toolsdialog.Outcome) tea.Cmd {
	switch {
	case out.Handover != nil:
		// A password needs a terminal of its own: the shell leaves and the
		// command runs there, with the answer already given.
		s.tools, s.Argv, s.quit = nil, out.Handover, true
	case out.Run != nil:
		argv := out.Run
		begin := s.sources.Start
		return func() tea.Msg {
			p, err := begin(argv)
			return toolsStartedMsg{proc: p, err: err}
		}
	case out.Stop:
		if s.toolsProc != nil {
			s.toolsProc.Stop()
		}
	case out.Done:
		s.tools = nil
		s.reopenResult()
		return s.refreshAfter([]string{"tools", "install"})
	}
	return nil
}

// toolsStarted follows an install that started.
func (s *Shell) toolsStarted(msg toolsStartedMsg) tea.Cmd {
	if s.tools == nil {
		if msg.proc != nil {
			msg.proc.Stop()
		}
		return nil
	}
	if msg.err != nil {
		m := s.tools.Output([]string{words.Say("tui.tools.could_not_start", "err", msg.err.Error())}).Finished(1, false)
		s.tools = &m
		return nil
	}
	s.toolsProc = msg.proc
	return waitTools(msg.proc)
}

// waitTools blocks, off the Update loop, until the install has output or ends.
func waitTools(p runner.Handle) tea.Cmd {
	return func() tea.Msg {
		lines, exit := p.Next()
		return toolsOutputMsg{proc: p, lines: lines, exit: exit}
	}
}

// toolsOutput shows the output of the install in the dialog and, at the end,
// goes on to the next tool.
func (s *Shell) toolsOutput(msg toolsOutputMsg) tea.Cmd {
	if s.tools == nil || s.toolsProc != msg.proc {
		return nil
	}
	m := s.tools.Output(msg.lines)
	if msg.exit == nil {
		s.tools = &m
		return waitTools(msg.proc)
	}
	m = m.Finished(msg.exit.Code, msg.exit.Cancelled)
	s.tools, s.toolsProc = &m, nil
	return nil
}

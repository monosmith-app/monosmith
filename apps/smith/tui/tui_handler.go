package tui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/screenreader"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingssource"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/userprefs"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
	"github.com/monosmith-app/monosmith/apps/smith/updatenotice"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"golang.org/x/term"
)

// Interactive reports whether a bare `smith` should ask what to do: someone
// is typing at a terminal and this is not CI. When standard output is not a
// terminal the questions are plain lines (see screenreader.Plain).
func Interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && os.Getenv(words.Say("tui.tui.ci")) == ""
}

// Choose runs the shell full screen, with the mouse, and returns the smith
// arguments of a command that needs a terminal of its own (it asks questions),
// or none when the user quit; every other command already ran inside the shell.
// catalog feeds the palette. resume, when set, opens that form first. A panic restores the terminal and comes back as
// an error.
func Choose(version string, out io.Writer, catalog []clicontract.CommandDescription, resume *Resume) (Chosen, error) {
	var start []string
	note := ""
	for {
		chosen, reopen, err := chooseOnce(version, out, catalog, start, note, resume)
		if err != nil || reopen == nil {
			return chosen, err
		}
		if reopen.Dir != "" {
			if err := os.Chdir(reopen.Dir); err != nil {
				return Chosen{}, err
			}
		}
		start, note, resume = reopen.Start, reopen.Note, nil
	}
}

// Chosen is what the shell hands back: the command to run in the terminal and
// the form and answers that made it ("" and nil when no form did).
type Chosen struct {
	Argv    []string
	Form    string
	Answers map[string][]string
}

// Resume is a guided form to open again because its command failed: the form
// keeps the answers and shows Problem, so one answer can be corrected.
type Resume struct {
	Form    string
	Answers map[string][]string
	Problem string
}

// reopening asks for the shell to start again, now inside the workspace the
// last one made; Start is a command to run as soon as it is open.
type reopening struct {
	Start []string
	// Note is a sentence for the first toast of the reopened shell.
	Note string
	// Dir is the folder to open it in, when a command made a workspace there.
	Dir string
}

// chooseOnce is one run of the shell. reopen is set when the folder became a
// workspace inside it (adoption), so the shell starts again to open it.
func chooseOnce(version string, out io.Writer, catalog []clicontract.CommandDescription, start []string, note string, resume *Resume) (chosen Chosen, reopen *reopening, err error) {
	defer func() {
		if r := recover(); r != nil {
			chosen, reopen, err = Chosen{}, nil, errors.New(words.Say("tui.handler.stopped_unexpectedly", "reason", r))
		}
	}()
	name := words.Say("tui.tui.no_workspace")
	inside := false
	enabled := pluginhost.BuiltinNames()
	problem, commitOps := "", false
	var environments []string
	if root, err := workspace.FindRoot("."); err == nil {
		cfg, err := workspace.Load(root)
		if err != nil {
			// smith.toml exists but cannot be read: say why (the error names the
			// line and column) and do not offer to adopt a folder that is one.
			problem = err.Error()
		} else {
			inside = true
			commitOps = cfg.Workspace.CommitOperations
			environments = cfg.Workspace.Environments
			name = cfg.Workspace.Name
			if len(cfg.Plugins.Enabled) > 0 {
				enabled = cfg.Plugins.Enabled
			}
		}
	}
	empty := !inside && sources.FolderIsEmpty(".")
	var actions []guide.Action
	for _, a := range append(guide.Actions(enabled, pluginhost.BuiltinNames(), enabled), guide.GeneratedForms(catalog)...) {
		// An unreadable smith.toml means the folder already is a workspace:
		// neither creating nor adopting one is offered over it.
		if (a.NeedsWorkspace && !inside) || (a.OutsideOnly && (inside || problem != "")) || (a.NeedsContent && empty) {
			continue
		}
		actions = append(actions, a)
	}
	if !inside && !empty && problem == "" {
		// A folder with files that is not a workspace yet: adoption is the
		// first entry, so Enter starts it.
		actions = guide.AdoptionFirst(actions)
	}
	stdoutTTY := term.IsTerminal(int(os.Stdout.Fd()))
	if screenreader.Plain(os.Getenv, stdoutTTY) {
		// No full-screen drawing, no mouse and no colour: plain numbered
		// questions, then the command, which smith runs in this terminal.
		// With output that is not a terminal the questions go to stderr, where
		// the person can still read them.
		if !stdoutTTY {
			out = os.Stderr
		}
		argv, err := screenreader.Ask(os.Stdin, out, actions)
		if err != nil {
			return Chosen{}, nil, err
		}
		if len(argv) > 0 {
			fmt.Fprint(out, words.Say("tui.handler.running", "command", commands.Line(argv)))
		}
		return Chosen{Argv: argv}, nil, nil
	}
	var repair *sources.ConfigRepair
	if problem != "" {
		// A renamed key is the one problem smith can fix: offer the change as
		// the first entry. Any other problem keeps its message.
		if r, ok := sources.ConfigRepairOf(".", version); ok {
			repair = &r
			actions = guide.FixConfigFirst(actions)
		}
	}
	// What the person saved in the Settings tab; the environment still wins over it.
	prefs, _ := userprefs.Load(pluginhost.Home())
	cfg := Config{
		Clipboard: out,
		Version:   version, Workspace: name, Actions: actions, Commands: catalog, EmptyFolder: empty,
		Problem: problem, Repair: repair, CommitOperations: commitOps, Environments: environments, Resume: resume,
		Theme: theme.New(theme.ChooseWith(os.Getenv, theme.DarkBackground(os.Getenv), prefs.Theme, prefs.Icons)),
		// The sizes of the panels are kept in the same file as the other choices.
		Splits: prefs.Splits, SaveSplits: saveSplits,
	}
	cfg.Notice = strings.Join(updatenotice.Start(updatenotice.ForThisProcess(version, 0)).Lines(), "  ")
	if note != "" {
		cfg.Notice = note
	}
	if inside {
		s := sources.WorkspaceSources(".", version)
		s.Settings = settingssource.New(".", pluginhost.Home(), version, os.Getenv)
		cfg.Sources = &s
		cfg.Start = start
	} else if problem == "" {
		cfg.Adopt = sources.AdoptionOpener(".", version)
		outside := sources.OutsideSources(".")
		cfg.Outside = &outside
	}
	// The alternate screen gives the shell the whole terminal and puts the
	// user's screen back on exit. Cell-motion mouse reports (clicks, drags,
	// the wheel; no hover) keep Shift+drag free for the terminal's own text
	// selection, and `m` turns the mouse off altogether.
	// The background is decided here, from the environment, so lipgloss and
	// termenv never ask the terminal: a terminal that ignores the query made
	// the screen stay blank for five seconds.
	lipgloss.SetHasDarkBackground(theme.DarkBackground(os.Getenv))
	final, err := tea.NewProgram(NewShell(cfg), tea.WithOutput(out), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	if err != nil {
		return Chosen{}, nil, err
	}
	sh := final.(Shell)
	if sh.reopen {
		return Chosen{}, &reopening{Start: sh.reopenStart, Note: sh.reopenNote, Dir: sh.reopenDir}, nil
	}
	if line := sh.ExitSummary(); line != "" {
		fmt.Fprintf(out, "%s\n", line)
	}
	argv := sh.Argv
	if len(argv) > 0 {
		fmt.Fprint(out, words.Say("tui.handler.chosen", "command", commands.Line(argv)))
	}
	return Chosen{Argv: argv, Form: sh.Form, Answers: sh.Answers}, nil, nil
}

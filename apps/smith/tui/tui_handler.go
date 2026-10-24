package tui

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/screenreader"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"golang.org/x/term"
)

// Interactive reports whether a bare `smith` should ask what to do: someone
// is typing at a terminal and this is not CI. When standard output is not a
// terminal the questions are plain lines (see screenreader.Plain).
func Interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && os.Getenv("CI") == ""
}

// Choose runs the shell full screen, with the mouse, and returns the smith
// arguments of a command that needs a terminal of its own (it asks questions),
// or nil when the user quit; every other command already ran inside the shell.
// catalog feeds the palette. A panic restores the terminal and comes back as
// an error.
func Choose(version string, out io.Writer, catalog []clicontract.CommandDescription) (argv []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			argv, err = nil, fmt.Errorf("the interface stopped unexpectedly: %v", r)
		}
	}()
	name := "no workspace"
	inside := false
	enabled := pluginhost.BuiltinNames()
	if root, err := workspace.FindRoot("."); err == nil {
		if cfg, err := workspace.Load(root); err == nil {
			inside = true
			name = cfg.Workspace.Name
			if len(cfg.Plugins.Enabled) > 0 {
				enabled = cfg.Plugins.Enabled
			}
		}
	}
	empty := !inside && sources.FolderIsEmpty(".")
	var actions []guide.Action
	for _, a := range append(guide.Actions(enabled, pluginhost.BuiltinNames(), enabled), guide.GeneratedForms(catalog)...) {
		if (a.NeedsWorkspace && !inside) || (a.OutsideOnly && inside) || (a.NeedsContent && empty) {
			continue
		}
		actions = append(actions, a)
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
			return nil, err
		}
		if len(argv) > 0 {
			fmt.Fprintf(out, "Running: %s\n\n", commands.Line(argv))
		}
		return argv, nil
	}
	cfg := Config{
		Clipboard: out,
		Version:   version, Workspace: name, Actions: actions, Commands: catalog, EmptyFolder: empty,
		Theme: theme.New(theme.Choose(os.Getenv, theme.DarkBackground(os.Getenv))),
	}
	if inside {
		s := sources.WorkspaceSources(".", version)
		cfg.Sources = &s
	}
	// The alternate screen gives the shell the whole terminal and puts the
	// user's screen back on exit. Cell-motion mouse reports (clicks, drags,
	// the wheel; no hover) keep Shift+drag free for the terminal's own text
	// selection, and `m` turns the mouse off altogether.
	final, err := tea.NewProgram(NewShell(cfg), tea.WithOutput(out), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	if err != nil {
		return nil, err
	}
	sh := final.(Shell)
	if line := sh.ExitSummary(); line != "" {
		fmt.Fprintf(out, "%s\n", line)
	}
	argv = sh.Argv
	if len(argv) > 0 {
		fmt.Fprintf(out, "→ %s\n\n", commands.Line(argv))
	}
	return argv, nil
}

package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/clipboard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// focusedText is what `y` copies: the remedy (else the message) of the
// selected Doctor check, the last command line of the Run view, the root of
// the selected project, the package of the selected dependency. ok is false
// when the view has nothing to copy.
func (s Shell) focusedText() (text string, ok bool) {
	switch s.tab {
	case commands.TabProjects:
		if p, found := s.list.Selected(); found && p.Root != "" {
			return p.Root, true
		}
	case commands.TabRun:
		if s.run != nil {
			return s.run.CommandLine(), true
		}
	case commands.TabDoctor:
		if f, found := s.doctor.Selected(); found {
			switch {
			case f.Remedy != "":
				return f.Remedy, true
			case f.Detail != "":
				return f.Detail, true
			}
			return f.Check, true
		}
	case commands.TabDeps:
		if s.deps.Cursor >= 0 && s.deps.Cursor < len(s.deps.Rows) {
			return s.deps.Rows[s.deps.Cursor].Name, true
		}
	}
	return "", false
}

// copyFocused puts the focused item on the clipboard and says so in a toast.
func (s *Shell) copyFocused() {
	text, ok := s.focusedText()
	if !ok || text == "" {
		s.say("Nothing to copy here")
		return
	}
	s.copyText(text)
}

// openRegistryPage opens the registry page of the selected dependency in the
// browser, and copies the link, as the report does. It runs only on this key:
// smith never opens a browser by itself.
func (s *Shell) openRegistryPage() tea.Cmd {
	r, ok := s.deps.Selected()
	switch {
	case !ok:
		s.say("No dependency is selected")
	case r.URL == "":
		s.say("No registry page is known for " + r.Name + " (a private registry whose web address smith cannot work out)")
	default:
		s.copyText(r.URL)
		if open := s.sources.Open; open != nil {
			link := r.URL
			return func() tea.Msg { return openedMsg{path: link, link: true, err: open(link)} }
		}
	}
	return nil
}

// copyText puts text on the clipboard and says so in a toast.
func (s *Shell) copyText(text string) {
	if err := clipboard.Copy(s.clip, text); err != nil {
		s.notify("Could not copy: "+err.Error(), toast.Warning)
		return
	}
	first, _, more := strings.Cut(text, "\n")
	if more {
		first += " …"
	}
	s.notify("copied: "+first, toast.Success)
}

// toggleMouse turns mouse reporting off, so the terminal selects text the
// ordinary way, or on again.
func (s *Shell) toggleMouse() tea.Cmd {
	s.mouseOff = !s.mouseOff
	if s.mouseOff {
		s.notify("Mouse off: drag selects text. Press m to turn it on", toast.Info)
		return tea.DisableMouse
	}
	s.notify("Mouse on: shift+drag selects text", toast.Info)
	return tea.EnableMouseCellMotion
}

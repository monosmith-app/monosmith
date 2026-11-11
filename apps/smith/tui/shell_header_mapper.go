package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/tabs"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// inlineHeaderMax is the most columns the header may take at the right end of
// the tab row: the tabs keep the rest.
const inlineHeaderMax = 40

// inlineHeader is true when the header shares the row of the tabs: the screen
// is wide enough (layout.InlineHeaderFrom) and the tabs keep their full names
// beside it. Otherwise the header has a row of its own, as on a narrow screen.
func (s Shell) inlineHeader() bool {
	if s.w < layout.InlineHeaderFrom || s.h <= 4 {
		return false
	}
	used := 0
	for _, sp := range s.stripAt(s.w, "").Spans {
		used = max(used, sp.X1)
	}
	return used+pane.Width(s.headerText())+2 <= s.w
}

// headerText is the header cut to inlineHeaderMax columns, for the tab row.
func (s Shell) headerText() string {
	return pane.Fit(strings.TrimLeft(strings.Join(s.headerParts(), "  "), " "), inlineHeaderMax)
}

// headerRight is the header as it sits at the right end of the tab row; ""
// when the header has its own row.
func (s Shell) headerRight() string {
	if !s.inlineHeader() {
		return ""
	}
	return s.headerText()
}

// tabRow is the first row on a wide screen: the tabs at the left and the header
// at the right.
func (s Shell) tabRow(strip tabs.Strip) string {
	right := s.headerRight()
	room := s.w - pane.Width(right) - 2
	left := pane.Fit(strip.Lines[0], room)
	return left + strings.Repeat(" ", max(s.w-pane.Width(left)-pane.Width(right)-1, 1)) + right + " "
}

// headerRow is the top row: the workspace, its branch and its health, at the
// right end of the row, where the tab row has them on a wide screen.
func (s Shell) headerRow() string {
	text := pane.Cut(strings.TrimLeft(strings.Join(s.headerParts(), "  "), " "), max(s.w-1, 0))
	return strings.Repeat(" ", max(s.w-pane.Width(text)-1, 0)) + text + " "
}

// headerParts are the pieces of the header: the workspace, what is wrong with
// it, its branch and its health.
func (s Shell) headerParts() []string {
	name := s.header
	if name == "" {
		name = "smith"
	}
	parts := []string{" " + s.th.Bold(name)}
	if !s.inWorkspace && s.problem != "" {
		// The full message, with its line, is wrapped in the guide below.
		parts = append(parts, s.th.Warning(words.Say("tui.header.smith_toml_cannot_be_read_see")))
	} else if !s.inWorkspace {
		parts = append(parts, s.th.Muted(strings.TrimSpace(s.noWorkspaceHint())))
	}
	if s.branch != "" {
		parts = append(parts, s.th.Muted("⎇ ")+s.branch)
	}
	if h := s.healthSummary(); h != "" {
		parts = append(parts, h)
	}
	return parts
}

// healthSummary is what the last doctor run found, as counts: "×2 ▲2" for two
// failures and two warnings, a check mark when everything passed, and "checking…"
// until the first doctor run (it starts with the shell) has finished.
func (s Shell) healthSummary() string {
	r := s.doctor.Report
	if r == nil {
		if s.doctor.Loading {
			return s.th.Muted("checking…")
		}
		return ""
	}
	var parts []string
	if r.Failed > 0 {
		parts = append(parts, s.th.Failure(fmt.Sprintf("%s%d", s.th.Icons.Failed, r.Failed)))
	}
	if r.Warnings > 0 {
		parts = append(parts, s.th.Warning(fmt.Sprintf("%s%d", s.th.Icons.Warning, r.Warnings)))
	}
	if len(parts) == 0 {
		return s.th.Success(s.th.Icons.Passed)
	}
	return strings.Join(parts, " ")
}

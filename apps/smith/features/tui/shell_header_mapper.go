package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// headerRow is the title bar, the first row at every width, on the bar
// background: the smith tag, the workspace and its branch at the left; the
// palette hint and the smith version at the right. The long hint gives way to the short one, then the
// hint, then the version go when the row is short.
func (s Shell) headerRow() string {
	left := " " + strings.Join(s.headerParts(), "  ")
	for _, right := range s.headerRightChoices() {
		if gap := s.w - pane.Width(left) - pane.Width(right) - 1; gap >= 2 {
			return s.th.Bar(left + strings.Repeat(" ", gap) + right + " ")
		}
	}
	// A header too long for the row ends with an ellipsis, as every cut line does.
	return s.th.Bar(pane.Fit(pane.CutEllipsis(left, max(s.w-1, 0)), s.w))
}

// headerRightChoices are the right ends of the title bar, widest first: the
// long palette hint and the version, the short hint and the version, the
// version alone.
func (s Shell) headerRightChoices() []string {
	key := s.th.Accent(paletteKey())
	long := s.th.Muted(words.Say("tui.header.palette_hint")) + "  " + key
	short := key + " " + s.th.Muted(words.Say("tui.header.palette_short"))
	version := ""
	if s.version != "" {
		version = s.th.Muted(words.Say("tui.header.version", "version", s.version))
	}
	var out []string
	for _, hint := range []string{long, short} {
		if version == "" {
			out = append(out, hint)
			continue
		}
		out = append(out, hint+"   "+version)
	}
	if version != "" {
		out = append(out, version)
	}
	return out
}

// paletteKey is the key that opens the command palette, as the footer names
// keys ("Ctrl+P"): the first chord of the palette command, else its first key.
func paletteKey() string {
	c, ok := commands.ByID("file.palette")
	if !ok || len(c.Keys) == 0 {
		return ""
	}
	for _, k := range c.Keys {
		if strings.Contains(k, "+") {
			return commands.KeyLabel(k)
		}
	}
	return commands.KeyLabel(c.Keys[0])
}

// headerParts are the pieces of the header: the smith tag, the workspace in
// bold, what is wrong with it, its branch (muted) and the state of its remote
// cache. The doctor state is in the status bar.
func (s Shell) headerParts() []string {
	parts := []string{s.th.ButtonActive(words.Say("tui.header.tag"))}
	if s.header != "" {
		parts = append(parts, s.th.Bold(s.header))
	}
	if !s.inWorkspace && s.problem != "" {
		// The full message, with its line, is wrapped in the guide below.
		parts = append(parts, s.th.Warning(words.Say("tui.header.smith_toml_cannot_be_read_see")))
	} else if !s.inWorkspace {
		parts = append(parts, s.th.Muted(strings.TrimSpace(s.noWorkspaceHint())))
	}
	if s.branch != "" {
		parts = append(parts, s.th.Muted(s.branch))
	}
	if r := s.remoteSummary(); r != "" {
		parts = append(parts, r)
	}
	return parts
}

// healthSummary is what the last doctor run found, for the status bar, as counts: "×2 ▲2" for two
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

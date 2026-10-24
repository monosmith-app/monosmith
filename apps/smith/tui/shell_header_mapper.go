package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// headerRow is the top row: the workspace, its branch and its health.
func (s Shell) headerRow() string {
	name := s.header
	if name == "" {
		name = "smith"
	}
	parts := []string{" " + s.th.Accent(name)}
	if !s.inWorkspace {
		parts = append(parts, s.th.Muted(strings.TrimSpace(s.noWorkspaceHint())))
	}
	if s.branch != "" {
		parts = append(parts, s.th.Muted("⎇ ")+s.branch)
	}
	if h := s.healthSummary(); h != "" {
		parts = append(parts, h)
	}
	return pane.Fit(strings.Join(parts, "  "), s.w)
}

// healthSummary is what the last doctor run found, as counts: "✖2 ▲2" for two
// failures and two warnings, a check mark when everything passed, and nothing
// until doctor has run (the Doctor view runs it).
func (s Shell) healthSummary() string {
	r := s.doctor.Report
	if r == nil {
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

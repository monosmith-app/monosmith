package releaseview

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Model shows what the next release would do; releasing itself runs
// `smith release` in the run view, after a confirmation that shows the command.
type Model struct {
	Loaded   bool
	Plan     release.Plan
	Blockers []string
	Warnings []string
	Err      error
}

// Result is what planning the release found: the shell receives it as a message.
type Result struct {
	Plan     release.Plan
	Blockers []string
	Warnings []string
	Err      error
}

// FromResult is the model for a planned release.
func FromResult(r Result) Model {
	return Model{Loaded: true, Plan: r.Plan, Blockers: r.Blockers, Warnings: r.Warnings, Err: r.Err}
}

// View draws the plan in w cells.
func (v Model) View(th theme.Theme, w int) []string {
	var lines []string
	switch {
	case v.Err != nil:
		lines = append(lines, th.Failure(v.Err.Error()))
	case !v.Loaded:
		lines = append(lines, th.Muted("Planning the release…"))
	case len(v.Blockers) > 0:
		for _, m := range v.Warnings {
			lines = append(lines, th.Warning("warning: "+m))
		}
		lines = append(lines, th.Failure("The release is blocked: its tags would give wrong versions."), "")
		for _, m := range v.Blockers {
			lines = append(lines, "  • "+m)
		}
		lines = append(lines, "", th.Muted("F2 shows the same with smith release --dry-run"))
	default:
		for _, m := range v.Warnings {
			lines = append(lines, th.Warning("warning: "+m))
		}
		if len(v.Plan.Items) == 0 {
			lines = append(lines, "Nothing to release.")
		}
		for _, item := range v.Plan.Items {
			from := item.Current
			if from == "" {
				from = "new"
			}
			lines = append(lines, fmt.Sprintf("  %-24s %s → %s  %s", item.Project, from, th.Accent(item.Next), th.Muted(item.Reason)))
		}
		if len(v.Plan.Unchanged) > 0 {
			lines = append(lines, "", th.Muted("unchanged: "+strings.Join(v.Plan.Unchanged, ", ")))
		}
		help := "F2 preview (smith release --dry-run)"
		if len(v.Plan.Items) > 0 {
			help += " · F3 release (smith release: confirms first)"
		}
		lines = append(lines, "", th.Muted(help))
	}
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines
}

// CanRelease reports whether a release can be started from the plan.
func (v Model) CanRelease() bool {
	return v.Loaded && v.Err == nil && len(v.Blockers) == 0 && len(v.Plan.Items) > 0
}

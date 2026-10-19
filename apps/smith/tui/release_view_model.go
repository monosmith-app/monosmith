package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// releaseView shows what the next release would do; releasing itself runs
// `smith release` in the run view, after a confirmation that shows the command.
type releaseView struct {
	loaded   bool
	plan     release.Plan
	blockers []string
	warnings []string
	err      error
}

type releaseLoaded struct {
	plan     release.Plan
	blockers []string
	warnings []string
	err      error
}

func (v releaseView) view(th theme.Theme, w int) []string {
	var lines []string
	switch {
	case v.err != nil:
		lines = append(lines, th.Failure(v.err.Error()))
	case !v.loaded:
		lines = append(lines, th.Muted("Planning the release…"))
	case len(v.blockers) > 0:
		for _, m := range v.warnings {
			lines = append(lines, th.Warning("warning: "+m))
		}
		lines = append(lines, th.Failure("The release is blocked: its tags would give wrong versions."), "")
		for _, m := range v.blockers {
			lines = append(lines, "  • "+m)
		}
		lines = append(lines, "", th.Muted("F2 shows the same with smith release --dry-run"))
	default:
		for _, m := range v.warnings {
			lines = append(lines, th.Warning("warning: "+m))
		}
		if len(v.plan.Items) == 0 {
			lines = append(lines, "Nothing to release.")
		}
		for _, item := range v.plan.Items {
			from := item.Current
			if from == "" {
				from = "new"
			}
			lines = append(lines, fmt.Sprintf("  %-24s %s → %s  %s", item.Project, from, th.Accent(item.Next), th.Muted(item.Reason)))
		}
		if len(v.plan.Unchanged) > 0 {
			lines = append(lines, "", th.Muted("unchanged: "+strings.Join(v.plan.Unchanged, ", ")))
		}
		help := "F2 preview (smith release --dry-run)"
		if len(v.plan.Items) > 0 {
			help += " · F3 release (smith release: confirms first)"
		}
		lines = append(lines, "", th.Muted(help))
	}
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines
}

// canRelease reports whether a release can be started from the plan.
func (v releaseView) canRelease() bool {
	return v.loaded && v.err == nil && len(v.blockers) == 0 && len(v.plan.Items) > 0
}

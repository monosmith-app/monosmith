package releaseconfirm

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// maxRows bounds the packages listed in the confirmation; the rest is a count.
const maxRows = 8

// Details are the lines of the release confirmation, between its sentence and
// the command: the checks, the packages with their versions and tags, and what
// the release does to the repository.
//
// Parameters: th draws the marks and the colours; plan is the plan the command
// line will run (holds and bumps already applied); held lists the packages held
// back; checks come from Preflight; f says where the tags go.
// Returns: the lines, without indentation (the dialog wraps them by words).
func Details(th theme.Theme, plan release.Plan, held []string, checks []Check, f Facts) []string {
	var lines []string
	if plan.Channel != "" {
		lines = append(lines, release.ChannelNotice(plan.Channel), "")
	}
	lines = append(lines, th.Bold(words.Say("tui.release_confirm.preflight")))
	for _, c := range checks {
		if c.Warn {
			lines = append(lines, th.Warning(th.Icons.Warning+" "+c.Text))
			continue
		}
		lines = append(lines, th.Success(th.Icons.Passed)+" "+c.Text)
	}
	if Warns(checks) {
		lines = append(lines, th.Muted(words.Say("tui.release_confirm.warning_hint")))
	}
	lines = append(lines, "", th.Bold(words.Say("tui.confirm_details.packages_and_versions_to_publish")))
	for i, item := range plan.Items {
		if i == maxRows {
			lines = append(lines, th.Muted(words.Say("tui.confirm_details.and_more", "value", len(plan.Items)-maxRows)))
			break
		}
		from := item.Current
		if from == "" {
			from = "new"
		}
		lines = append(lines, words.Say("tui.release_confirm.row", "project", item.Project, "from", from, "next", item.Next, "tag", item.Tag))
	}
	if len(held) > 0 {
		lines = append(lines, th.Muted(words.Say("tui.confirm_details.held_back_not_released", "list", strings.Join(held, ", "))))
	}
	lines = append(lines, "", th.Bold(words.Say("tui.release_confirm.effects")), words.Say("tui.release_confirm.commit"), words.Count("tui.release_confirm.tag", len(plan.Items)), pushLine(f), words.Say("tui.release_confirm.publish"))
	return lines
}

// pushLine says where the tags go.
func pushLine(f Facts) string {
	switch {
	case f.NoPush:
		return words.Say("tui.release_confirm.push_off")
	case !f.HasRemote:
		return words.Say("tui.release_confirm.push_no_remote")
	}
	return words.Say("tui.release_confirm.push", "remote", f.Remote)
}

// RunLabel is the label of the button that confirms: it says "anyway" when a
// check warns, so going on is a choice that was seen.
func RunLabel(checks []Check) string {
	if Warns(checks) {
		return words.Say("tui.release_confirm.run_anyway")
	}
	return words.Say("tui.release_confirm.run")
}

package releaseview

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// View draws the plan in a w by h area; rowOf maps a line to the plan item drawn
// on it, or to RowRelease, RowFirst or RowWhy for the line of a button or fold.
func (v Model) View(th theme.Theme, w int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Err != nil:
		lines = append(lines, th.Failure(v.Err.Error()))
	case !v.Loaded:
		lines = append(lines, pane.Wrap(th.Muted(progress.Waiting(words.Say("releaseview.waiting", "projects", progress.OfProjects(v.Projects)), v.Waited)+words.Say("releaseview.waiting.more")), w, 0)...)
	case len(v.Blockers) > 0:
		lines = append(lines, th.Failure(th.Icons.Failed+" "+words.Say("releaseview.blocked")), "")
		for _, m := range v.Warnings {
			lines = append(lines, th.Warning(words.Say("releaseview.warning", "message", m)))
		}
		for _, m := range v.Blockers {
			lines = append(lines, "  • "+m)
		}
		lines = append(lines, "", th.Muted(words.Say("releaseview.blocked.hint")))
	default:
		if notice := release.PreviewNotice(v.Git.Branch, v.Plan); notice != "" {
			lines = append(lines, th.Muted(notice))
		}
		lines = append(lines, v.headline(th))
		if v.CanRelease() && v.Button != "" {
			rowOf[len(lines)] = RowRelease
			lines = append(lines, v.Button)
		}
		lines = append(lines, "")
		if v.Plan.Channel != "" {
			// A prerelease line is not the stable release: say so before the versions.
			lines = append(lines, th.Warning(th.Icons.Warning+" "+release.ChannelNotice(v.Plan.Channel)), "")
		}
		for _, m := range v.Warnings {
			lines = append(lines, th.Warning(words.Say("releaseview.warning", "message", m)))
		}
		if len(v.Plan.Items) == 0 {
			lines = v.emptyLines(th, lines, rowOf)
		}
		nameW := 8
		for _, item := range v.Plan.Items {
			nameW = max(nameW, pane.Width(item.Project))
		}
		nameW = min(nameW, max(w/3, 8))
		for i, item := range v.Plan.Items {
			from := item.Current
			if from == "" {
				from = words.Say("releaseview.new")
			}
			selected := i == v.Cursor
			cursor := choicerow.Cursor(th, false)
			if selected {
				cursor = th.Accent(choicerow.Cursor(th, true))
			}
			rowOf[len(lines)] = i
			// The mark says the row opens: ▸ closed, ▾ while its commits show.
			mark := th.Icons.Closed
			if selected && v.Open {
				mark = th.Icons.Open
			}
			lines = append(lines, cursor+th.Muted(mark)+" "+pane.MiddleCell(item.Project, nameW+1)+from+" → "+th.Bold(item.Next)+"  "+levelMark(th, item.Level)+"  "+th.Muted(item.Reason))
			if selected && v.Open {
				lines = append(lines, v.commits(th, item)...)
			}
		}
		if len(v.Plan.Unchanged) > 0 {
			lines = append(lines, "", th.Muted(words.Say("releaseview.unchanged", "projects", strings.Join(v.Plan.Unchanged, ", "))))
		}
	}
	// A long line wraps (the unchanged list, the help) instead of being cut.
	lines, rowOf = pane.WrapRows(lines, rowOf, w, 2)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// headline says what the release is in a few words, before the detail.
func (v Model) headline(th theme.Theme) string {
	n := len(v.Plan.Items)
	if n == 0 {
		return th.Bold(v.emptyHeadline())
	}
	levels := map[string]int{}
	for _, item := range v.Plan.Items {
		levels[item.Level]++
	}
	var parts []string
	for _, l := range []string{"major", "minor", "patch"} {
		if levels[l] > 0 {
			parts = append(parts, strconv.Itoa(levels[l])+" "+l)
		}
	}
	return th.Bold(words.Count("releaseview.headline", n)) + th.Muted(" ("+strings.Join(parts, " · ")+")")
}

// levelMark names a bump's level with a symbol, so it reads without colour.
func levelMark(th theme.Theme, level string) string {
	switch level {
	case "major":
		return th.Failure("major")
	case "minor":
		return th.Warning("minor")
	case "patch":
		return th.Success("patch")
	}
	return th.Muted(level)
}

// commits is the drill-down: the tag and the commits that make the project release.
func (v Model) commits(th theme.Theme, item release.Item) []string {
	out := []string{"      " + th.Muted(words.Say("releaseview.tag", "tag", item.Tag, "level", item.Level, "plugin", item.Plugin))}
	if len(item.Changes) == 0 {
		return append(out, "      "+th.Muted(words.Say("releaseview.nocommit", "reason", item.Reason)))
	}
	for i, c := range item.Changes {
		if i == maxCommits {
			out = append(out, "      "+th.Muted(words.Say("releaseview.more", "count", len(item.Changes)-maxCommits)))
			break
		}
		kind := c.Type
		if len(c.Scopes) > 0 {
			kind += "(" + strings.Join(c.Scopes, ",") + ")"
		}
		if c.Breaking {
			kind += "!"
		}
		line := "      " + th.Muted(short(c.Hash)) + " " + th.Bold(kind) + ": " + c.Description
		if c.Author != "" {
			line += th.Muted("  " + c.Author)
		}
		out = append(out, line)
		if c.BreakingNote != "" {
			out = append(out, "        "+th.Failure(words.Say("releaseview.breaking", "note", c.BreakingNote)))
		}
	}
	return out
}

// short is a commit hash as git abbreviates it.
func short(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

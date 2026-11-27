package releaseview

import (
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The widths of the version and bump columns of the plan.
const verW, bumpW = 10, 11

// View draws the plan in w cells: the headline with the count of each bump,
// then the table Package, Now, Next, Bump and Why with a rule under its
// heading, and the packages that do not change. rowOf maps a line to the plan
// item drawn on it, or to RowRelease, RowFirst or RowWhy for the line of a
// button or fold.
//
// Parameters: th draws, w is the width in cells.
// Returns: the lines (each w cells wide) and what each line stands for.
func (v Model) View(th theme.Theme, w int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Err != nil:
		lines = append(lines, th.Failure(v.Err.Error()))
	case !v.Loaded:
		// A spinner says at once that the plan is on its way.
		wait := th.Info(th.Spin(v.Tick)) + " " + th.Muted(progress.Waiting(words.Say("releaseview.waiting", "projects", progress.OfProjects(v.Projects)), v.Waited)+words.Say("releaseview.waiting.more"))
		lines = append(lines, pane.Wrap(wait, w, 0)...)
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
		} else {
			lines = v.table(th, w, lines, rowOf)
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

// table adds the plan's table to lines and records the row of each item.
func (v Model) table(th theme.Theme, w int, lines []string, rowOf map[int]int) []string {
	nameW := 8
	for _, item := range v.Plan.Items {
		nameW = max(nameW, pane.Width(item.Project))
	}
	nameW = min(nameW+2, max(w/3, 8))
	whyW := max(w-2-nameW-2*verW-bumpW, 6)
	cols := func(name, now, next, bump, why string) string {
		return pane.MiddleCell(name, nameW) + pane.Cell(now, verW) + pane.Cell(next, verW) + pane.Cell(bump, bumpW) + pane.Cell(why, whyW)
	}
	lines = append(lines, th.Muted("  "+cols(words.Say("releasepage.column.package"), words.Say("releasepage.column.now"), words.Say("releasepage.column.next"), words.Say("releasepage.column.bump"), words.Say("releasepage.column.why"))))
	lines = append(lines, " "+th.Border(strings.Repeat("─", max(w-2, 1)), false))
	for i, item := range v.Plan.Items {
		from := item.Current
		if from == "" {
			from = words.Say("releaseview.new")
		}
		rowOf[len(lines)] = i
		if i == v.Cursor {
			// The selected look of every list: the cursor mark and the accent on the Highlight background.
			lines = append(lines, th.Selected(th.Accent(pane.Fit(choicerow.Cursor(th, true)+cols(item.Project, from, item.Next, bumpText(th, item.Level), item.Reason), w))))
			continue
		}
		lines = append(lines, choicerow.Cursor(th, false)+cols(item.Project, from, th.Bold(item.Next), BumpMark(th, item.Level), th.Muted(item.Reason)))
	}
	return lines
}

// headline says what the release is in a few words: the mark, how many
// packages release, how many of each bump and the branch it is made from.
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
	if v.Git.Branch != "" {
		parts = append(parts, words.Say("releasepage.from", "branch", v.Git.Branch))
	}
	return th.Success(th.Icons.Passed) + " " + th.Bold(words.Count("releaseview.headline", n)) + th.Muted(" · "+strings.Join(parts, " · "))
}

// bumpText is the mark and the name of a bump, without colour.
func bumpText(th theme.Theme, level string) string {
	if level == "major" {
		return th.Icons.Warning + " " + level
	}
	return th.Icons.Dot + " " + level
}

// BumpMark names a bump with its mark in its colour, so it reads without
// colour too: a dot for a patch (success) or a minor bump (info), the warning
// mark for a major one.
//
// Parameters: th draws, level is "patch", "minor" or "major".
// Returns: the coloured mark and the level.
func BumpMark(th theme.Theme, level string) string {
	switch level {
	case "major":
		return th.Warning(bumpText(th, level))
	case "patch":
		return th.Success(bumpText(th, level))
	case "minor":
		return th.Info(bumpText(th, level))
	}
	return th.Muted(level)
}

// commits are the commits that make the project release, newest first, at
// most maxCommits of them.
func (v Model) commits(th theme.Theme, item release.Item) []string {
	if len(item.Changes) == 0 {
		return []string{th.Muted(words.Say("releaseview.nocommit", "reason", item.Reason))}
	}
	var out []string
	for i, c := range item.Changes {
		if i == maxCommits {
			out = append(out, th.Muted(words.Say("releaseview.more", "count", len(item.Changes)-maxCommits)))
			break
		}
		kind := c.Type
		if len(c.Scopes) > 0 {
			kind += "(" + strings.Join(c.Scopes, ",") + ")"
		}
		if c.Breaking {
			kind += "!"
		}
		line := th.Muted(short(c.Hash)) + " " + th.Bold(kind) + ": " + c.Description
		if c.Author != "" {
			line += th.Muted("  " + c.Author)
		}
		out = append(out, line)
		if c.BreakingNote != "" {
			out = append(out, "  "+th.Failure(words.Say("releaseview.breaking", "note", c.BreakingNote)))
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

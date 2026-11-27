package depsview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The widths of the version and level columns of the table.
const verW, levelW = 10, 10

// levelCounts counts the rows of each level.
func (v Model) levelCounts() (patch, minor, major int) {
	for _, r := range v.All {
		switch r.Level {
		case "patch":
			patch++
		case "minor":
			minor++
		default:
			major++
		}
	}
	return
}

// inProjects says whose dependencies a load asks about, once the count is known.
func inProjects(n int) string {
	if n <= 0 {
		return ""
	}
	return words.Say("depsview.waiting.scope", "projects", progress.OfProjects(n))
}

// View draws the list in a w by h area: the headline with the count of each
// level, the chips, then the safe updates and the major ones as two groups,
// each with its count, its column heading and a rule. rowOf maps a line to the
// row drawn on it, or to HeadingRow for a column heading (its box marks every row).
//
// Parameters: th draws, w and h are the cells of the area.
// Returns: the lines (each w cells wide) and what each line stands for.
func (v Model) View(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Err != nil:
		return []string{th.Failure(v.Err.Error())}, rowOf
	case !v.Loaded:
		// A spinner and the seconds say the wait is alive and how long it has taken.
		wait := th.Info(th.Spin(v.Tick)) + " " + th.Muted(progress.Waiting(words.Say("depsview.waiting", "scope", inProjects(v.Projects)), v.Waited))
		return pane.Wrap(wait+th.Muted(words.Say("depsview.waiting.more")), w, 0), rowOf
	}
	if len(v.All) == 0 {
		lines = append(lines, th.Success(th.Icons.Passed+" "+words.Say("depsview.latest")))
		lines = append(lines, v.policyTail(th, w)...)
		return fitAll(lines, w), rowOf
	}
	lines = append(lines, v.headline(th), "")
	head, _ := v.headlineRows(th, w)
	lines = append(lines, head...)
	// What an update of the chosen packages risks is said above the list, where no scrolling hides it.
	if n := v.chosenMajors(); n > 0 {
		lines = append(lines, pane.Wrap(th.Warning(th.Icons.Warning+" "+words.Say("depsview.chosen.majors", "count", n)), w, 2)...)
	}
	lines = append(lines, "")
	fixed := len(lines)
	body, bodyRow, cursorLine := v.groups(th, w)
	if len(v.Rows) == 0 {
		body = append([]string{" " + th.Muted(v.noneText())}, v.policyTail(th, w)...)
	}
	avail := max(h-fixed, 3)
	top := v.scroll.Show(cursorLine, avail, len(body))
	shown := body[top:min(top+avail, len(body))]
	for i := range shown {
		if r, ok := bodyRow[top+i]; ok {
			rowOf[fixed+i] = r
		}
	}
	if len(body) > avail {
		shown = pane.Scrollbar(th, shown, w, top, len(body))
	}
	lines = append(lines, shown...)
	return fitAll(lines, w), rowOf
}

// fitAll cuts or pads every line to w cells.
func fitAll(lines []string, w int) []string {
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines
}

// headline is the first line: how many packages have newer releases, how many
// of each level and how many are marked.
func (v Model) headline(th theme.Theme) string {
	patch, minor, major := v.levelCounts()
	parts := []string{words.Say("depspage.level.patch", "count", patch), words.Say("depspage.level.minor", "count", minor), words.Say("depspage.level.major", "count", major)}
	return th.Bold(words.Say("depspage.headline", "count", len(v.All))) + th.Muted(" · "+strings.Join(parts, " · ")+words.Say("depsview.marked", "count", len(v.Marked)))
}

// policyTail is the one-version policy under the list, for a narrow screen with
// no package to open (the detail of a package tells it otherwise).
func (v Model) policyTail(th theme.Theme, w int) []string {
	if v.Beside {
		return nil
	}
	out := []string{"", th.Title(PolicyTitle())}
	for _, it := range v.PolicyItems(th) {
		out = append(out, pane.Wrap("  "+it, w, 4)...)
	}
	return out
}

// groups draws the rows the chip lets through as two groups: the patch and
// minor releases, safe to take together, then the major ones. Each group has
// its heading with its count, the column heading and a rule. rowOf maps a line
// of body to its row; cursorLine is the line of the row under the cursor.
func (v Model) groups(th theme.Theme, w int) (body []string, rowOf map[int]int, cursorLine int) {
	rowOf = map[int]int{}
	nameW := 8
	for _, r := range v.Rows {
		nameW = max(nameW, pane.Width(r.Name))
	}
	nameW = min(nameW+2, max(w/3, 8))
	usedW := max(w-6-nameW-2*verW-levelW, 6)
	cols := func(name, now, latest, level, used string) string {
		return pane.MiddleCell(name, nameW) + pane.Cell(now, verW) + pane.Cell(latest, verW) + pane.Cell(level, levelW) + pane.Cell(used, usedW)
	}
	headBox := choicerow.Box(v.MarkState() == 2, v.MarkState() == 1)
	for _, major := range []bool{false, true} {
		var idx []int
		for i, r := range v.Rows {
			if (r.Level == "major") == major {
				idx = append(idx, i)
			}
		}
		if len(idx) == 0 {
			continue
		}
		if len(body) > 0 {
			body = append(body, "")
		}
		if major {
			body = append(body, " "+th.Warning(words.Say("depspage.group.major", "count", len(idx))))
		} else {
			body = append(body, " "+th.Success(words.Say("depspage.group.safe", "count", len(idx))))
		}
		rowOf[len(body)] = HeadingRow
		body = append(body, th.Muted("  "+headBox+" "+cols(words.Say("depsview.column.package"), words.Say("depsview.column.now"), words.Say("depsview.column.latest"), words.Say("depsview.column.level"), words.Say("depspage.column.used"))))
		body = append(body, " "+th.Border(strings.Repeat("─", max(w-2, 1)), false))
		for _, i := range idx {
			r := v.Rows[i]
			box := choicerow.Box(v.Marked[r.Name], false) + " "
			used := usedText(r.Projects, usedW)
			rowOf[len(body)] = i
			if i == v.Cursor {
				// The selected look of every list: the cursor mark and the accent on the Highlight background.
				cursorLine = len(body)
				body = append(body, th.Selected(th.Accent(pane.Fit(choicerow.Cursor(th, true)+box+cols(r.Name, r.Current, r.Latest, levelText(th, r.Level), used), w))))
				continue
			}
			body = append(body, choicerow.Cursor(th, false)+box+cols(r.Name, r.Current, th.Bold(r.Latest), LevelMark(th, r.Level), used))
		}
	}
	return body, rowOf, cursorLine
}

// usedText names the projects that use a package in w cells: all of them when
// they fit, else the first and how many more.
func usedText(projects []string, w int) string {
	all := strings.Join(projects, ", ")
	if pane.Width(all) <= w || len(projects) < 2 {
		return all
	}
	return words.Say("depspage.used.more", "first", projects[0], "count", len(projects)-1)
}

// levelText is the mark and the name of a level, without colour.
func levelText(th theme.Theme, level string) string {
	if level == "major" {
		return th.Icons.Warning + " " + level
	}
	return th.Icons.Dot + " " + level
}

// LevelMark names a release level with its mark in its colour: a dot for a
// patch (success) or a minor release (info), the warning mark for a major one.
//
// Parameters: th draws, level is "patch", "minor" or "major".
// Returns: the coloured mark and the level.
func LevelMark(th theme.Theme, level string) string {
	switch level {
	case "major":
		return th.Warning(levelText(th, level))
	case "patch":
		return th.Success(levelText(th, level))
	}
	return th.Info(levelText(th, level))
}

// noneText says that the chip that is on holds no release.
func (v Model) noneText() string {
	switch v.Filter {
	case FilterPatch:
		return words.Say("depsview.none.patch")
	case FilterMinor:
		return words.Say("depsview.none.minor")
	}
	return words.Say("depsview.none.major")
}

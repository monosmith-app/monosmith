package depsview

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// levelCounts counts the rows of each level.
func (v Model) levelCounts() (patch, minor, major int) {
	for _, r := range v.Rows {
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

// View draws the view in a w by h area; rowOf maps a line to the row drawn on it.
func (v Model) View(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Err != nil:
		return []string{th.Failure(v.Err.Error())}, rowOf
	case !v.Loaded:
		return pane.Wrap(th.Muted(progress.Waiting(words.Say("depsview.waiting", "scope", inProjects(v.Projects)), v.Waited)+words.Say("depsview.waiting.more")), w, 0), rowOf
	}
	var tail []string
	tail = append(tail, "", th.Title(words.Say("depsview.policy.title")))
	if len(v.Policy) == 0 {
		tail = append(tail, "  "+th.Muted(words.Say("depsview.policy.empty")))
	}
	for _, p := range v.Policy {
		mark := th.Success(th.Icons.Passed)
		switch {
		case !p.OK && p.Warning:
			mark = th.Warning(th.Icons.Warning)
		case !p.OK:
			mark = th.Failure(th.Icons.Failed)
		}
		tail = append(tail, "  "+mark+" "+p.Check)
		if !p.OK && p.Detail != "" {
			for _, d := range strings.Split(p.Detail, "; ") {
				tail = append(tail, "      "+th.Muted(d))
			}
		}
	}
	if n := v.chosenMajors(); n > 0 {
		tail = append(tail, "", th.Warning(th.Icons.Warning+" "+words.Say("depsview.chosen.majors", "count", n)))
	}
	var wrapped []string
	for _, l := range tail {
		wrapped = append(wrapped, pane.Wrap(l, w, 4)...)
	}
	tail = wrapped

	if len(v.Rows) == 0 {
		lines = append(lines, th.Success(th.Icons.Passed+" "+words.Say("depsview.latest")))
	} else {
		patch, minor, major := v.levelCounts()
		lines = append(lines, th.Bold(words.Say("depsview.summary", "count", len(v.Rows), "patch", patch, "minor", minor, "major", major))+th.Muted(words.Say("depsview.marked", "count", len(v.Marked))))
		if safe := patch + minor; safe > 0 {
			lines = append(lines, pane.Wrap(th.Success(words.Say("depsview.safe", "count", safe))+th.Muted(majorWarning(major)), w, 0)...)
		}
		lines = append(lines, "")
		nameW := 8
		for _, r := range v.Rows {
			nameW = max(nameW, pane.Width(r.Name))
		}
		nameW = min(nameW+1, max(w/3, 8))
		const verW, levelW, kindW = 11, 8, 12
		usedW := max(w-6-nameW-2*verW-levelW-kindW, 8)
		cols := func(name, now, latest, level, kind, projects string) string {
			return pane.MiddleCell(name, nameW) + pane.Cell(now, verW) + pane.Cell(latest, verW) + pane.Cell(level, levelW) + pane.Cell(kind, kindW) + pane.Cell(projects, usedW)
		}
		headBox := choicerow.Box(v.MarkState() == 2, v.MarkState() == 1)
		rowOf[len(lines)] = HeadingRow
		lines = append(lines, th.Muted(pane.Fit("  "+headBox+" "+cols(words.Say("depsview.column.package"), words.Say("depsview.column.now"), words.Say("depsview.column.latest"), words.Say("depsview.column.level"), words.Say("depsview.column.kind"), words.Say("depsview.column.projects")), w)))
		// Two group headings may sit among the rows: they take room too.
		avail := max(h-len(lines)-len(tail)-2, 3)
		top := v.scroll.Show(v.Cursor, avail, len(v.Rows))
		first := len(lines)
		prevMajor := top > 0 && v.Rows[top-1].Level == "major"
		for i := top; i < len(v.Rows) && i < top+avail; i++ {
			r := v.Rows[i]
			if isMajor := r.Level == "major"; i == top || isMajor != prevMajor {
				if isMajor {
					lines = append(lines, "  "+th.Warning(words.Say("depsview.group.major")))
				} else {
					lines = append(lines, "  "+th.Success(words.Say("depsview.group.safe")))
				}
			}
			prevMajor = r.Level == "major"
			box := choicerow.Box(v.Marked[r.Name], false) + " "
			level := th.Muted(r.Level)
			switch r.Level {
			case "major":
				level = th.Warning(r.Level)
			case "patch":
				level = th.Success(r.Level)
			}
			kind := strings.Join(r.Kinds, ", ")
			projects := strings.Join(r.Projects, ", ")
			plain := "  " + box + cols(r.Name, r.Current, r.Latest, r.Level, kind, projects)
			line := "  " + box + cols(r.Name, r.Current, r.Latest, level, kind, projects)
			if i == v.Cursor {
				line = th.Selected(pane.Fit(plain, w))
			}
			rowOf[len(lines)] = i
			lines = append(lines, line)
		}
		if len(v.Rows) > avail {
			// The rows and the group headings between them scroll as one block.
			block := pane.Scrollbar(th, lines[first:], w, top, len(v.Rows)+len(lines[first:])-min(avail, len(v.Rows)-top))
			copy(lines[first:], block)
		}
	}
	lines = append(lines, tail...)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// majorWarning is the sentence about major releases after the count of the safe
// ones: nothing when there is none, singular and plural otherwise.
func majorWarning(major int) string {
	switch major {
	case 0:
		return ""
	case 1:
		return words.Say("depsview.major.one")
	}
	return words.Say("depsview.major.many", "count", major)
}

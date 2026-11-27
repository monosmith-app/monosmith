package doctorview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// View draws the page in w cells: the headline with its three filter buttons
// first, then the checks the filters let through, one line each (the state and
// the title), the passed ones behind one line unless they are on. With expand the open check shows its problem, remedy and fix under
// its line; without it the detail beside the list says all that, so the page
// does not repeat it. rowOf maps a line to the finding drawn on it.
func (v Model) View(th theme.Theme, w int, expand bool) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if v.Err != nil {
		return []string{th.Failure(v.Err.Error())}, rowOf
	}
	if v.Report == nil {
		return []string{th.Muted(progress.Waiting(words.Say("doctor.waiting", "projects", progress.OfProjects(v.Covered)), v.Waited))}, rowOf
	}
	head, _ := v.headlineRows(th, w)
	lines = append(lines, head...)
	lines = append(lines, "")
	headEnd := len(lines)
	position := 0
	for gi, g := range v.groups() {
		open := v.groupOpen(g)
		if v.Grouped {
			rowOf[len(lines)] = HeaderCode(gi)
			lines = append(lines, v.header(th, g, open, position == v.Cursor))
			position++
		}
		if !open {
			continue
		}
		indent := ""
		if v.Grouped {
			indent = "  "
		}
		var projectOf map[int]string
		if v.ByState {
			projectOf = v.projectsOf()
		}
		for _, i := range g.Indexes {
			f := v.Report.Findings[i]
			selected := position == v.Cursor
			rowOf[len(lines)] = i
			line := choicerow.Line(th, choicerow.Row{Lead: indent + Mark(th, f) + " ", Label: f.Title(), Selected: selected}, 0)
			if _, ok := v.FixFor(f); ok && v.ByState {
				line = withTag(line, th.Muted(words.Say("doctor.row.fix")), w)
			}
			lines = append(lines, line)
			if v.ByState {
				if where := whereLine(projectOf[i], FileOf(f)); where != "" {
					rowOf[len(lines)] = i
					lines = append(lines, "  "+indent+"  "+th.Muted(where))
				}
			}
			if selected && v.Open && expand && !f.OK {
				lines = append(lines, Text(th, f, 5+len(indent))...)
			}
			position++
		}
	}
	if v.NothingMatches() {
		lines = append(lines, " "+th.Muted(words.Say("doctor.filter.none", "failed", FilterFailed.Key(), "warnings", FilterWarnings.Key(), "passed", FilterPassed.Key())))
	}
	if v.collapsed() {
		if len(lines) > headEnd {
			lines = append(lines, "")
		}
		rowOf[len(lines)] = PassedRow
		lines = append(lines, choicerow.Line(th, choicerow.Row{Lead: th.Icons.Closed + " ", Label: words.Say("doctor.passed.line", "count", v.Report.Passed), Hint: words.Say("doctor.passed.hint"), Selected: v.OnPassedLine()}, 0))
	}
	if len(v.Extras) > 0 {
		lines = append(lines, "", " "+th.Muted(words.Say("doctor.extra.heading")))
		first := len(v.layout()) - len(v.Extras)
		for i, e := range v.Extras {
			rowOf[len(lines)] = ExtraCode(i)
			lines = append(lines, choicerow.Line(th, choicerow.Row{Lead: th.Icons.Closed + " ", Label: e.Title, Hint: e.Hint, Selected: v.Cursor == first+i}, 0))
		}
	}
	lines, rowOf = pane.WrapRows(lines, rowOf, w, 5)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// Mark is the symbol of a finding's state.
func Mark(th theme.Theme, f doctor.Finding) string {
	switch {
	case !f.OK && f.Warning:
		return th.Warning(th.Icons.Warning)
	case !f.OK:
		return th.Failure(th.Icons.Failed)
	}
	return th.Success(th.Icons.Passed)
}

// Text is a check told in full, each part once: the problem, the remedy, the
// fix and the files it touches, indented by indent cells.
func Text(th theme.Theme, f doctor.Finding, indent int) []string {
	pad := strings.Repeat(" ", indent)
	var out []string
	if f.Detail != "" {
		out = append(out, pad+th.Muted(f.Detail))
	}
	if f.Remedy != "" {
		out = append(out, pad+th.Bold(words.Say("doctor.remedy.prefix"))+f.Remedy)
	}
	if f.OK {
		return out
	}
	switch {
	case f.Fix != nil && f.Fix.Safe && len(f.Fix.Choices) == 0:
		out = append(out, pad+th.Bold(words.Say("doctor.fix.prefix"))+f.Fix.Description+th.Muted(words.Say("doctor.fix.safe.hint")))
	case f.Fix != nil:
		out = append(out, pad+th.Bold(words.Say("doctor.fix.prefix"))+f.Fix.Description+th.Muted(words.Say("doctor.fix.asks.hint")))
		// The first answer is the suggested one: say it, so f then Enter is known to do it.
		if s := suggestion(*f.Fix); s != "" {
			out = append(out, pad+"  "+th.Muted(words.Say("doctor.fix.suggestion", "suggestion", s)))
		}
	}
	if f.Fix != nil {
		for _, e := range f.Fix.Files {
			key := "doctor.file.writes"
			if e.Delete {
				key = "doctor.file.removes"
			}
			out = append(out, pad+"  "+th.Muted(words.Say(key, "path", e.Path)))
		}
	}
	return out
}

// suggestion is the first option of a fix's only question: the answer the
// wizard has under its cursor. A fix with several questions has none, since
// one key cannot answer them all.
func suggestion(fix protocol.Fix) string {
	if len(fix.Choices) != 1 || len(fix.Choices[0].Options) == 0 {
		return ""
	}
	return fix.Choices[0].Options[0].Label
}

// withTag puts tag at the right end of a row of w cells, when it fits.
func withTag(line, tag string, w int) string {
	gap := w - pane.Width(line) - pane.Width(tag) - 2
	if gap < 1 {
		return line
	}
	return line + strings.Repeat(" ", gap) + tag
}

// whereLine is the second line of a check grouped by state: its project (the
// workspace for none) and the file it is about.
func whereLine(project, file string) string {
	if project == "" {
		project = words.Say("doctor.group.workspace")
	}
	if file == "" {
		return project
	}
	return words.Say("doctor.row.where", "project", project, "file", file)
}

// projectsOf is the project each finding is filed under, by its index ("" for
// the workspace), as GroupByProject files them.
func (v Model) projectsOf() map[int]string {
	out := map[int]string{}
	for _, g := range GroupByProject(v.Report.Findings, v.Projects) {
		for _, i := range g.Indexes {
			out[i] = g.Project
		}
	}
	return out
}

// stateHeader is the heading of a group of one state: its fold mark, the
// state in its colour and bold, and the count of its checks.
func (v Model) stateHeader(th theme.Theme, g Group, open, selected bool) string {
	counted := g.Counted
	if counted == nil {
		counted = g.Indexes
	}
	label, paint := words.Say("doctor.state.failed"), th.Failure
	switch g.State {
	case StateWarning:
		label, paint = words.Say("doctor.state.warnings"), th.Warning
	case StatePassed:
		label, paint = words.Say("doctor.state.passed"), th.Success
	}
	mark := th.Icons.Closed
	if open {
		mark = th.Icons.Open
	}
	return choicerow.Line(th, choicerow.Row{Lead: mark + " ", Label: th.Bold(paint(label)) + "  " + th.Muted(words.Say("doctor.state.count", "count", len(counted))), Selected: selected}, 0)
}

// header is a group's heading: its fold mark, its name and what is wrong in it.
func (v Model) header(th theme.Theme, g Group, open, selected bool) string {
	if g.State != "" {
		return v.stateHeader(th, g, open, selected)
	}
	name := g.Project
	if name == "" {
		name = words.Say("doctor.group.workspace")
	}
	failed, warned := 0, 0
	counted := g.Counted
	if counted == nil {
		counted = g.Indexes
	}
	for _, i := range counted {
		switch f := v.Report.Findings[i]; {
		case f.OK:
		case f.Warning:
			warned++
		default:
			failed++
		}
	}
	var facts []string
	if failed > 0 {
		facts = append(facts, th.Failure(words.Say("doctor.group.failed", "count", failed)))
	}
	if warned > 0 {
		facts = append(facts, th.Warning(words.Count("doctor.group.warnings", warned)))
	}
	if len(facts) == 0 {
		facts = append(facts, th.Muted(words.Say("doctor.group.passed")))
	}
	mark := th.Icons.Closed
	if open {
		mark = th.Icons.Open
	}
	return choicerow.Line(th, choicerow.Row{Lead: mark + " ", Label: th.Title(name) + "  " + strings.Join(facts, " · "), Selected: selected}, 0)
}

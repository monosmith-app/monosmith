package doctorview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Headline is the state of the workspace in a few words: "2 failed · 2 warnings".
func (v Model) Headline(th theme.Theme) string {
	r := v.Report
	var parts []string
	if r.Failed > 0 {
		parts = append(parts, th.Failure(words.Say("doctor.headline.failed", "icon", th.Icons.Failed, "count", r.Failed)))
	}
	if r.Warnings > 0 {
		parts = append(parts, th.Warning(words.Say("doctor.headline.warnings", "icon", th.Icons.Warning, "count", r.Warnings)))
	}
	if len(parts) == 0 {
		parts = append(parts, th.Success(words.Say("doctor.headline.passed", "icon", th.Icons.Passed)))
	}
	return strings.Join(parts, th.Muted(" · "))
}

// View draws the page in w cells: the headline first, then the failing
// checks one line each (the state and the title), the passed ones behind one
// line. With expand the open check shows its problem, remedy and fix under
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
	if v.Grouped {
		rowOf[0] = ToggleAllRow
	}
	lines = append(lines, v.headline(th, w), "")
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
		for _, i := range g.Indexes {
			f := v.Report.Findings[i]
			selected := position == v.Cursor
			rowOf[len(lines)] = i
			lines = append(lines, guide.Row(th, selected, indent+Mark(th, f)+" ", f.Title(), ""))
			if selected && v.Open && expand && !f.OK {
				lines = append(lines, Text(th, f, 5+len(indent))...)
			}
			position++
		}
	}
	if v.collapsed() {
		if len(lines) > 2 {
			lines = append(lines, "")
		}
		rowOf[len(lines)] = PassedRow
		lines = append(lines, guide.Row(th, v.OnPassedLine(), th.Icons.Closed+" ", words.Say("doctor.passed.line", "count", v.Report.Passed), words.Say("doctor.passed.hint")))
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

// headline is the state of the workspace, and, with groups, the button at the
// right end that folds or unfolds all of them.
func (v Model) headline(th theme.Theme, w int) string {
	left := " " + v.Headline(th)
	if !v.Grouped {
		return left
	}
	label := words.Say("doctor.toggle.collapse")
	if !v.allOpen() {
		label = words.Say("doctor.toggle.expand")
	}
	toggle := button.Render(th, label, commands.ShortcutOf("doctor.foldall"), button.Rest)
	gap := w - pane.Width(left) - pane.Width(toggle) - 1
	return left + strings.Repeat(" ", max(gap, 2)) + toggle
}

// header is a group's heading: its fold mark, its name and what is wrong in it.
func (v Model) header(th theme.Theme, g Group, open, selected bool) string {
	name := g.Project
	if name == "" {
		name = words.Say("doctor.group.workspace")
	}
	failed, warned := 0, 0
	for _, i := range g.Indexes {
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
		facts = append(facts, th.Warning(words.Say("doctor.group.warnings", "count", warned)))
	}
	if len(facts) == 0 {
		facts = append(facts, th.Muted(words.Say("doctor.group.passed")))
	}
	mark := th.Icons.Closed
	if open {
		mark = th.Icons.Open
	}
	return guide.Row(th, selected, mark+" ", th.Title(name)+"  "+strings.Join(facts, " · "), "")
}

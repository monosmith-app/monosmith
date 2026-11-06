package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// derivedSection lists the derived files of a project with their state in
// plain words, one row per file; the heading counts those that need action.
// It returns false for a project with none (derived files are off, or the
// project has none).
func (s Shell) derivedSection(p projectlist.Project) (detail.Section, bool) {
	if len(p.Derived) == 0 {
		return detail.Section{}, false
	}
	heading := words.Say("tui.derived.derived_files")
	if n := p.DerivedNeedingAction(); n > 0 {
		heading += words.Say("tui.derived.need_action", "n", n)
	}
	rows := make([]detail.Row, 0, len(p.Derived))
	for _, f := range p.Derived {
		mark := s.th.Icons.Passed
		switch {
		case f.NeedsDecision():
			mark = s.th.Icons.Warning
		case f.Status != projectlist.StatusCurrent && f.Status != projectlist.StatusEjected && f.Status != projectlist.StatusDiffers:
			mark = s.th.Icons.Pointer
		}
		rows = append(rows, detail.Row{Label: f.Path, Value: mark + " " + projectlist.StatusWords(f.Status)})
	}
	return detail.Section{Heading: heading, Rows: rows}, true
}

// derivedNeedingAction counts the derived files of all projects that derive
// would write or that wait for a choice.
func derivedNeedingAction(projects []projectlist.Project) int {
	n := 0
	for _, p := range projects {
		n += p.DerivedNeedingAction()
	}
	return n
}

// derivedActions are the things to do about the derived files of a project:
// Explain, Apply, Adopt, Eject and Replace. One that no file takes is shown
// off, with the reason under it.
func (s Shell) derivedActions(p projectlist.Project) []detail.Action {
	if len(p.Derived) == 0 {
		return nil
	}
	action := func(id, label, hint, verb string) detail.Action {
		a := detail.Action{ID: id, Label: label, Hint: hint, Disabled: s.running()}
		if p.DeriveArgv(verb) == nil {
			a.Disabled, a.Hint = true, words.Say("tui.derived.no_file_of_this_project_needs", "hint", hint)
		}
		return a
	}
	return []detail.Action{
		action("x:derive-explain", words.Say("tui.derived.explain_the_derived_files"), words.Say("tui.derived.who_owns_each_file_and_why"), projectlist.DeriveExplain),
		{ID: "x:derive-apply", Label: words.Say("tui.derived.update_the_derived_files"), Hint: words.Say("tui.derived.writes_the_files_that_are_out"), Disabled: s.running()},
		action("x:derive-adopt", words.Say("tui.derived.keep_my_edit_of_a_derived"), words.Say("tui.derived.takes_your_edit_of_a_keys"), projectlist.DeriveAdopt),
		action("x:derive-eject", words.Say("tui.derived.own_the_files_i_edited"), words.Say("tui.derived.the_files_become_yours_smith_stops"), projectlist.DeriveEject),
		action("x:derive-replace", words.Say("tui.derived.replace_the_files_i_edited"), words.Say("tui.derived.overwrites_them_with_what_smith_derives"), projectlist.DeriveReplace),
	}
}

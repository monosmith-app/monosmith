package tui

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
)

// derivedSection lists the derived files of a project with their state in
// plain words, one row per file; the heading counts those that need action.
// It returns false for a project with none (derived files are off, or the
// project has none).
func (s Shell) derivedSection(p projectlist.Project) (detail.Section, bool) {
	if len(p.Derived) == 0 {
		return detail.Section{}, false
	}
	heading := "Derived files"
	if n := p.DerivedNeedingAction(); n > 0 {
		heading += fmt.Sprintf(" (%d need action)", n)
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
			a.Disabled, a.Hint = true, hint+" (no file of this project needs it)"
		}
		return a
	}
	return []detail.Action{
		action("x:derive-explain", "Explain the derived files", "who owns each file and why it is in its state (smith derive --explain)", projectlist.DeriveExplain),
		{ID: "x:derive-apply", Label: "Update the derived files", Hint: "writes the files that are out of date or missing, for the whole workspace; never overwrites a file you edited (smith derive)", Disabled: s.running()},
		action("x:derive-adopt", "Keep my edit of a derived file", "takes your edit of a keys file as the new baseline (smith derive --adopt)", projectlist.DeriveAdopt),
		action("x:derive-eject", "Own the files I edited", "the files become yours, smith stops touching them (smith derive --eject)", projectlist.DeriveEject),
		action("x:derive-replace", "Replace the files I edited", "overwrites them with what smith derives; your edit is lost (smith derive --replace)", projectlist.DeriveReplace),
	}
}

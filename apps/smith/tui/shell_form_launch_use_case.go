package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// slicesCheck is the catalogue action of `smith slices check`, the check of the
// vertical-slice layout that doctor does not run itself.
const slicesCheck = "slices check" // words:literal

// openFormFor opens the guided form of an action of the catalogue ("why"), the
// same one the palette and the Start tab open, with answers already filled in
// for the person to confirm or change; without answers it opens as it is. It
// says so when the workspace has no such form.
func (s *Shell) openFormFor(id string, answers map[string][]string) tea.Cmd {
	a, ok := s.guide.ActionPresenting(id)
	if !ok {
		s.say(words.Say("tui.keys.no_form_for", "action", id))
		return nil
	}
	if len(answers) == 0 {
		return s.perform(commands.Command{Kind: commands.Guided, Guide: a.Title})
	}
	if g, ok := s.guide.Resume(a.Title, answers, ""); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// doctorExtras are the checks the Doctor tab lists after the findings because
// doctor does not run them: the vertical-slice layout, when the workspace has
// its form.
func (s Shell) doctorExtras() []doctorview.Extra {
	if _, ok := s.guide.ActionPresenting(slicesCheck); !ok {
		return nil
	}
	return []doctorview.Extra{{Title: words.Say("doctor.extra.slices"), Hint: words.Say("doctor.extra.slices.hint"), Presents: slicesCheck}}
}

// extraAnswers fill in the form of a check doctor does not run: the slice
// check starts on the folders of the projects in the workspace's main
// language, checked as that language.
func (s Shell) extraAnswers(id string) map[string][]string {
	if id != slicesCheck {
		return nil
	}
	lang := s.mainLanguage()
	var roots []string
	for _, p := range s.list.Projects() {
		if p.Language == lang {
			roots = append(roots, p.Root)
		}
	}
	if len(roots) == 0 {
		return nil
	}
	return map[string][]string{"arg:project directory": {strings.Join(roots, " ")}, "flag:lang": {lang}}
}

// onDoctorExtra is true when the Doctor list has a check doctor does not run
// under its cursor.
func (s Shell) onDoctorExtra() bool {
	_, ok := s.doctor.SelectedExtra()
	return ok
}

// runSelectedExtra opens the form of the check doctor does not run that the
// Doctor list has selected; ok is false when none is.
func (s *Shell) runSelectedExtra() (tea.Cmd, bool) {
	e, ok := s.doctor.SelectedExtra()
	if !ok {
		return nil, false
	}
	return s.openFormFor(e.Presents, s.extraAnswers(e.Presents)), true
}

// whyFile opens the why form for the file the selected finding is about.
func (s *Shell) whyFile() tea.Cmd {
	f, ok := s.doctor.Selected()
	if !ok {
		return nil
	}
	file := doctorview.FileOf(f)
	if file == "" {
		return nil
	}
	return s.openFormFor("why", map[string][]string{"path": {file}})
}

// whyMissReason says why the run panel's task cannot be asked why it ran
// again: none is selected, it came from the cache, or it did not run; "" when
// it ran.
func (s Shell) whyMissReason() string {
	if s.run == nil || s.runSel == "" {
		return words.Say("tui.keys.why_miss_select_task")
	}
	t := s.run.Task(s.runSel)
	switch {
	case t == nil:
		return words.Say("tui.keys.why_miss_select_task")
	case t.Status == runview.Cached:
		return words.Say("tui.keys.why_miss_cached", "task", t.ID)
	case t.Status != runview.Passed && t.Status != runview.Failed:
		return words.Say("tui.keys.why_miss_not_run", "task", t.ID)
	}
	return ""
}

// whyMiss opens the why-miss form for the task selected in the run panel,
// one that ran instead of coming from the cache.
func (s *Shell) whyMiss() tea.Cmd {
	if reason := s.whyMissReason(); reason != "" {
		s.say(reason)
		return nil
	}
	return s.openFormFor("why-miss", map[string][]string{"task": {s.runSel}})
}

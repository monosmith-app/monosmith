package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// sliceTask is the cached task the vertical-slices plugin gives each project
// of a language its checker knows: `smith slices check` on the project.
const sliceTask = "slice-check" // words:literal

// openFormFor opens the guided form of an action of the catalogue ("why"), the
// same one the palette and the Start tab open, with answers already filled in
// for the person to confirm or change; without answers it opens as it is. It
// says so when the workspace has no such form.
func (s *Shell) openFormFor(id string, answers map[string][]string) tea.Cmd {
	a, ok := s.guide.ActionPresenting(id)
	if !ok || len(answers) == 0 {
		return s.openFormPresenting(id)
	}
	if g, ok := s.guide.Resume(a.Title, answers, ""); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// doctorExtras are the checks the Doctor tab lists after the findings because
// doctor does not run them: the vertical-slice layout, when a project can run it.
func (s Shell) doctorExtras() []doctorview.Extra {
	var tasks []string
	if s.anyProjectHas(sliceTask, "") {
		tasks = append(tasks, sliceTask)
	}
	if s.anyProjectHas("lint", "node") {
		// The checker does not know Node: ESLint's vertical-slices rules check
		// the layout of a Node project as part of lint.
		tasks = append(tasks, "lint")
	}
	if len(tasks) == 0 {
		return nil
	}
	return []doctorview.Extra{{
		Title: words.Say("doctor.extra.slices"),
		Hint:  words.Say("doctor.extra.slices.hint", "tasks", strings.Join(tasks, " and ")),
		Run:   append([]string{"run"}, tasks...),
	}}
}

// anyProjectHas is true when a project of the workspace has the task; a
// language that is not "" also asks that the project is in it.
func (s Shell) anyProjectHas(task, language string) bool {
	for _, p := range s.list.Projects() {
		if _, ok := p.Task(task); ok && (language == "" || p.Language == language) {
			return true
		}
	}
	return false
}

// onDoctorExtra is true when the Doctor list has a check doctor does not run
// under its cursor.
func (s Shell) onDoctorExtra() bool {
	_, ok := s.doctor.SelectedExtra()
	return ok
}

// runSelectedExtra runs the check doctor does not run that the Doctor list has
// selected, at once and with no question, in the run panel; ok is false when
// none is selected.
func (s *Shell) runSelectedExtra() (tea.Cmd, bool) {
	e, ok := s.doctor.SelectedExtra()
	if !ok {
		return nil, false
	}
	return s.execute(e.Run, false), true
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

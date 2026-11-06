package guide

import (
	"maps"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pathfinder"
)

// Model is the guided menu and its question steps: pick an action, answer
// its questions one at a time, and it ends as a smith command line.
type Model struct {
	actions []Action
	action  *Action
	step    int
	cursor  int
	marked  map[int]bool
	input   textinput.Model
	answers map[string][]string
	err     string
	// notice is a muted line that says what the last tick did to the others.
	notice string
	// filter is what was typed to narrow a Filter step's options.
	filter string
	// searched is what was typed in the last Filter step that was answered; a
	// later step can start on the option it names (Step.Preselect).
	searched string
	// defaults are the options a step starts on, by step key.
	defaults map[string]string
	// defaultWhy is the line that says why a step starts on its default.
	defaultWhy map[string]string
	// Argv is the chosen command, empty until the questions are answered.
	Argv []string
	// Wizard names the wizard an action handed itself to (Action.Wizard).
	Wizard string
	// prefill are answers to start the steps on, by step key (see Resume).
	prefill map[string][]string
	// done is set when a command is chosen, or the user backed out of the top.
	done bool
	// view is where the list's window starts; a pointer, so it survives the
	// copies of the model between frames (see pane.Scroll).
	view *listView
	// words opens the glossary of a step on a short screen.
	words bool
	// trail are the steps already answered, in order: Esc goes back to the last
	// one. Steps that were skipped or answered by themselves are not in it.
	trail []int
	// replace is true while a text field holds an earlier value that the next
	// letter replaces.
	replace bool
	// path is the list of folders and files under a Path step's field; pathRoot
	// is the folder it starts in ("" is the folder smith started in).
	path     pathfinder.Model
	pathRoot string
}

// WithPathRoot sets the folder the path finder of a Path step starts in; the
// default is the folder smith started in.
func (m Model) WithPathRoot(dir string) Model {
	m.pathRoot = dir
	return m
}

// rederive makes the later questions follow an answer that was changed after
// going back: the ticks the old answer had given (a framework picked before) are
// replaced by the ticks the new answer gives, and the other answers stay.
func (m Model) rederive(old []string) Model {
	if m.prefill == nil {
		return m
	}
	prefill := maps.Clone(m.prefill)
	for i := range m.action.Steps {
		s := &m.action.Steps[i]
		if s.Ticked == nil {
			continue
		}
		before, ok := prefill[s.Key]
		if !ok {
			continue
		}
		kept := slices.DeleteFunc(slices.Clone(before), func(v string) bool { return slices.Contains(old, v) })
		for _, v := range s.Ticked(m.answers) {
			if !slices.Contains(kept, v) {
				kept = append(kept, v)
			}
		}
		prefill[s.Key] = kept
	}
	m.prefill = prefill
	return m
}

// Reopen takes a form that ended back to its last question, with every answer
// filled in: Esc on the plan of the command it made.
func (m Model) Reopen() Model {
	m.done, m.Argv = false, nil
	return m.back()
}

// back goes to the question before this one, with every answer still filled in.
// On the first question it leaves the form for the list of actions.
func (m Model) back() Model {
	if len(m.trail) == 0 {
		m.action, m.step, m.cursor, m.err, m.trail = nil, 0, 0, "", nil
		return m.newView()
	}
	if s := m.current(); s != nil && s.Options == nil {
		// What was typed and not entered yet is kept for the way forward.
		if v := strings.TrimSpace(m.input.Value()); v != "" {
			m.answers[s.Key] = []string{v}
		}
	}
	m.step = m.trail[len(m.trail)-1]
	m.trail = slices.Clone(m.trail[:len(m.trail)-1])
	m.prefill = maps.Clone(m.answers)
	m.cursor, m.marked, m.err, m.filter, m.notice = 0, map[int]bool{}, "", "", ""
	m.input.SetValue("")
	return m.newView().advance()
}

// listView is the window of the list on screen: where it starts, and how many
// rows a page key moves.
type listView struct {
	scroll pane.Scroll
	page   int
}

// NewModel builds the guide.
func NewModel(actions []Action) Model {
	in := textinput.New()
	in.CharLimit = 50
	return Model{actions: actions, input: in, marked: map[int]bool{}, view: &listView{}}
}

// Wheel scrolls the list by delta lines; the selection stays where it is.
func (m Model) Wheel(delta int) Model {
	if m.view != nil {
		m.view.scroll.By(delta)
	}
	return m
}

// HasGlossary reports whether the question on screen has words to explain
// that the "?" key opens or folds.
func (m Model) HasGlossary() bool {
	s := m.current()
	return s != nil && s.Glossary
}

// newView starts the list of the next question at its top.
func (m Model) newView() Model {
	m.view, m.words = &listView{}, false
	return m
}

// Done reports whether the guide finished: with Argv set, a command was
// chosen; with Argv empty, the user left the top level.
func (m Model) Done() bool { return m.done }

// Title is the heading of the chosen action ("Add a project"), or "" while
// the list of actions is showing.
func (m Model) Title() string {
	if m.action == nil {
		return ""
	}
	return m.action.Title
}

// WithAction puts an action in the menu, in the place of the one with the same
// title (or at the end): the shell uses it for a form whose questions come from
// what it just read, such as the choices of the plugins.
func (m Model) WithAction(a Action) Model {
	actions := slices.Clone(m.actions)
	for i := range actions {
		if actions[i].Title == a.Title {
			actions[i] = a
			m.actions = actions
			return m
		}
	}
	m.actions = append(actions, a)
	return m
}

// ActionTitled returns the action with a title; ok is false when there is none.
func (m Model) ActionTitled(title string) (Action, bool) {
	for _, a := range m.actions {
		if a.Title == title {
			return a, true
		}
	}
	return Action{}, false
}

// Reset returns the guide to its first screen, to be used again.
func (m Model) Reset() Model {
	r := NewModel(m.actions)
	r.defaults = m.defaults
	r.defaultWhy = m.defaultWhy
	r.pathRoot = m.pathRoot
	return r
}

// WithDefault makes the step with a key start on the option with a value, for
// the questions whose best answer depends on the workspace (the language the
// workspace mostly uses).
func (m Model) WithDefault(key, value string) Model {
	d := make(map[string]string, len(m.defaults)+1)
	for k, v := range m.defaults {
		d[k] = v
	}
	d[key] = value
	m.defaults = d
	return m
}

// WithDefaultWhy sets the line a step shows under its question to say why it
// starts on its default ("Preselected: ..."); the line is empty for no reason.
func (m Model) WithDefaultWhy(key, why string) Model {
	d := make(map[string]string, len(m.defaultWhy)+1)
	for k, v := range m.defaultWhy {
		d[k] = v
	}
	d[key] = why
	m.defaultWhy = d
	return m
}

// Editing reports whether a text field has the keyboard, so letters that are
// shortcuts elsewhere must go to it.
func (m Model) Editing() bool {
	s := m.current()
	return s != nil && (s.Options == nil || s.Filter)
}

// Answers are the answers given so far, by step key.
func (m Model) Answers() map[string][]string { return m.answers }

// Resume starts the action with a title again with the answers of an earlier
// try already filled in, and shows problem under the first question. It is
// how a failed command lets the person correct one answer instead of
// answering everything again; ok is false when there is no such action.
func (m Model) Resume(title string, answers map[string][]string, problem string) (Model, bool) {
	for i, a := range m.actions {
		if a.Title == title {
			r := m.Reset()
			r.prefill, r.cursor = answers, i
			r = r.confirm()
			r.err = problem
			return r, true
		}
	}
	return m, false
}

// Pick starts the action with a title, as if chosen from the list; ok is
// false when there is no such action.
func (m Model) Pick(title string) (Model, bool) {
	for i, a := range m.actions {
		if a.Title == title {
			m = m.Reset()
			m.cursor = i
			return m.confirm(), true
		}
	}
	return m, false
}

func (m Model) current() *Step {
	if m.action == nil || m.step >= len(m.action.Steps) {
		return nil
	}
	s := &m.action.Steps[m.step]
	if s.OptionsFor != nil {
		c := *s
		c.Options = s.OptionsFor(m.answers)
		return &c
	}
	return s
}

// options are the answers on offer: all of them, or for a Filter step those
// that match what was typed.
func (m Model) options() []Option {
	s := m.current()
	if s == nil {
		return nil
	}
	if !s.Filter || m.filter == "" {
		return s.Options
	}
	var out []Option
	for _, o := range s.Options {
		if MatchesFilter(o, m.filter) {
			out = append(out, o)
		}
	}
	return out
}

func (m Model) optionCount() int {
	if s := m.current(); s != nil {
		if s.Options == nil && s.Path {
			return m.path.Count()
		}
		return len(m.options())
	}
	if m.action != nil {
		return 0
	}
	return len(m.actions)
}

// problem is what is wrong with a typed answer, or "" when it is fine.
func (m Model) problem(step *Step, value string) string {
	if step.Check == nil {
		return ""
	}
	return step.Check(m.answers, value)
}

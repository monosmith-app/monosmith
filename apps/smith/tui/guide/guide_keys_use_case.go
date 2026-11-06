package guide

import (
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pathfinder"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model; only key messages matter.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m.Handle(msg) }

// Handle is Update returning the concrete model.
func (m Model) Handle(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	step := m.current()
	text := step != nil && step.Options == nil
	if key.Type == tea.KeyTab && step != nil {
		if text && step.Suggest != nil {
			if v := strings.TrimSpace(m.input.Value()); v != "" && m.problem(step, v) != "" {
				if s := step.Suggest(m.answers, v); s != "" {
					m.input.SetValue(s)
					m.input.CursorEnd()
				}
			}
			return m, nil
		}
		if step.Filter && step.Options != nil && m.filter != "" && len(m.options()) == 0 {
			if near := Suggestions(step.Options, m.filter); len(near) > 0 {
				m.filter, m.cursor = near[0], 0
			}
			return m, nil
		}
	}
	if text && step.Path {
		var out pathfinder.Outcome
		if m.path, out = m.path.Handle(key); out.Handled {
			m.cursor = m.path.Cursor()
			if out.Set {
				m.input.SetValue(out.Value)
				m.input.CursorEnd()
				m.replace, m.err = false, ""
			}
			if out.Confirm {
				return m.confirm(), nil
			}
			return m, nil
		}
	}
	if text && key.Type != tea.KeyEnter && key.Type != tea.KeyEsc && key.Type != tea.KeyCtrlC {
		if m.replace && (key.Type == tea.KeyRunes || key.Type == tea.KeySpace) {
			m.input.SetValue("")
		}
		m.replace = false
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if step.Path {
			m.path, m.cursor = m.path.Sync(m.input.Value()), 0
		}
		return m, cmd
	}
	if step != nil && step.Filter && step.Options != nil {
		switch key.Type {
		case tea.KeyRunes, tea.KeySpace:
			m.filter += string(key.Runes)
			m.cursor = 0
			return m, nil
		case tea.KeyBackspace:
			if r := []rune(m.filter); len(r) > 0 {
				m.filter, m.cursor = string(r[:len(r)-1]), 0
			}
			return m, nil
		case tea.KeyEsc:
			if m.filter != "" {
				m.filter, m.cursor = "", 0
				return m, nil
			}
		}
	}
	switch nav := commands.Nav(key.String()); nav {
	case commands.NavUp:
		m.cursor = max(m.cursor-1, 0)
	case commands.NavDown:
		m.cursor = min(m.cursor+1, max(m.optionCount()-1, 0))
	case commands.NavPageUp, commands.NavPageDown, commands.NavHome, commands.NavEnd:
		page := 8
		if m.view != nil && m.view.page > 1 {
			page = m.view.page - 1
		}
		m.cursor = pane.MoveCursor(m.cursor, m.optionCount(), nav, page)
	default:
		switch key.String() {
		case "ctrl+c":
			m.done = true
		case "esc", "q":
			switch {
			case m.action != nil && key.String() == "esc":
				m = m.back()
			case m.action != nil:
				m.action, m.step, m.cursor, m.err, m.trail = nil, 0, 0, "", nil
				m = m.newView()
			case key.String() == "esc":
				m.done = true
			}
		case "?":
			if m.HasGlossary() {
				m.words = !m.words
			}
		case " ":
			if step != nil && step.Multi {
				m.marked, m.notice = ToggleOption(m.marked, step.Options, m.cursor)
			}
		case "enter":
			return m.confirm(), nil
		}
	}
	return m, nil
}

// SelectRow moves the cursor to option i (a mouse click); in a multiple
// choice it also toggles the option.
func (m Model) SelectRow(i int) Model {
	if s := m.current(); s != nil && s.Options == nil && s.Path {
		m.path = m.path.Select(i)
		m.cursor = m.path.Cursor()
		return m
	}
	if i >= 0 && i < m.optionCount() {
		m.cursor = i
		if s := m.current(); s != nil && s.Multi {
			m.marked, m.notice = ToggleOption(m.marked, s.Options, i)
		}
	}
	return m
}

func (m Model) confirm() Model {
	if m.action == nil {
		if m.cursor < 0 || m.cursor >= len(m.actions) {
			return m
		}
		a := m.actions[m.cursor]
		if a.Wizard != "" {
			m.Wizard, m.done = a.Wizard, true
			return m
		}
		m.action, m.step, m.answers, m.cursor, m.marked, m.trail = &a, 0, map[string][]string{}, 0, map[int]bool{}, nil
		return m.newView().advance()
	}
	step := m.current()
	if step == nil {
		return m.advance()
	}
	switch {
	case step.Options == nil:
		value := strings.TrimSpace(m.input.Value())
		if value == "" && !step.Optional {
			m.err = words.Say("guide.guide.type_a_value_or_esc_to")
			return m
		}
		if problem := m.problem(step, value); problem != "" {
			m.err = problem
			return m
		}
		m.answers[step.Key] = []string{value}
		m.input.SetValue("")
	case step.Multi:
		var picked []string
		for i, o := range step.Options {
			if m.marked[i] {
				picked = append(picked, o.Value)
			}
		}
		if len(picked) == 0 && !step.Optional {
			m.err = words.Say("guide.guide.toggle_at_least_one_with_space")
			return m
		}
		if problem := m.problem(step, strings.Join(picked, ",")); problem != "" {
			m.err = problem
			return m
		}
		m.answers[step.Key] = picked
	default:
		opts := m.options()
		if len(opts) == 0 {
			m.err = words.Say("guide.guide.nothing_matches", "filter", m.filter)
			return m
		}
		if step.Filter {
			m.searched = m.filter
		}
		picked := opts[min(m.cursor, len(opts)-1)].Value
		if problem := m.problem(step, picked); problem != "" {
			m.err = problem
			return m
		}
		old := m.answers[step.Key]
		m.answers[step.Key] = []string{picked}
		if len(old) > 0 && !slices.Equal(old, m.answers[step.Key]) {
			m = m.rederive(old)
		}
	}
	m.trail = append(slices.Clone(m.trail), m.step)
	m.step++
	m.cursor, m.marked, m.err, m.filter, m.notice = 0, map[int]bool{}, "", "", ""
	return m.newView().advance()
}

func (m Model) advance() Model {
	m.replace = false
	for m.action != nil {
		s := m.current()
		if s == nil {
			break
		}
		if s.Skip != nil && s.Skip(m.answers) {
			m.step++
			continue
		}
		if s.Take != nil {
			if v, ok := s.Take(m.answers); ok {
				m.answers[s.Key] = v
				m.step++
				continue
			}
		}
		if s.Auto != nil {
			if v := s.Auto(m.answers); v != "" {
				m.answers[s.Key] = []string{v}
				m.step++
				continue
			}
		}
		break
	}
	step := m.current()
	if step == nil {
		if m.action != nil && m.action.Argv != nil {
			m.Argv = m.action.Argv(m.answers)
		}
		m.done = true
		return m
	}
	if step.Multi {
		for i, o := range step.Options {
			m.marked[i] = o.Hint == "enabled" || step.Key == "tasks" ||
				(step.Ticked != nil && slices.Contains(step.Ticked(m.answers), o.Value))
		}
	}
	if step.Multi && step.Ticked != nil {
		// The cursor starts on the first pick of an earlier question, so the list
		// opens where the tick is, not above it on a short terminal.
		for i, o := range step.Options {
			if slices.Contains(step.Ticked(m.answers), o.Value) {
				m.cursor = i
				break
			}
		}
	}
	if step.Options == nil {
		m.input.Placeholder = step.Placeholder
		if step.PlaceholderFor != nil {
			m.input.Placeholder = step.PlaceholderFor(m.answers)
		}
		if step.Initial != nil {
			m.input.SetValue(step.Initial(m.answers))
		}
		m.input.Focus()
	} else if want := m.defaults[step.Key]; want != "" {
		for i, o := range step.Options {
			if o.Value == want {
				m.cursor = i
			}
		}
	}
	if step.Preselect != nil && !step.Multi {
		if want := step.Preselect(m.searched, step.Options); want != "" {
			for i, o := range step.Options {
				if o.Value == want {
					m.cursor = i
				}
			}
		}
	}
	if before, ok := m.prefill[step.Key]; ok && len(before) > 0 {
		if step.Options == nil {
			m.input.SetValue(before[0])
			// The earlier value is shown selected: the next letter replaces it,
			// Backspace edits it. Typing never appends to it silently.
			m.replace = true
		}
		for i, o := range step.Options {
			if step.Multi {
				m.marked[i] = slices.Contains(before, o.Value)
			} else if o.Value == before[0] {
				m.cursor = i
			}
		}
	}
	if step.Options == nil && step.Path {
		var settings []pathfinder.Setting
		if step.PathSettings {
			settings = settingEntries()
		}
		m.path, m.cursor = pathfinder.New(m.pathRoot, settings).Sync(m.input.Value()), 0
	}
	return m
}

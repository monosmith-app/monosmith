package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pathfinder"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// View implements tea.Model. It can be called in any state, including the
// last frame Bubble Tea draws after the final answer.
func (m Model) View() string {
	lines, _ := m.Body(theme.Default(), 80)
	return strings.Join(lines, "\n")
}

// Body draws the current screen in w cells. A long entry wraps onto further
// lines; optionAt maps every line an option takes to the option, for a click.
func (m Model) Body(th theme.Theme, w int) (lines []string, optionAt map[int]int) {
	return m.body(th, w, false)
}

// body is Body with the glossary of the step folded to one line when fold is true.
func (m Model) body(th theme.Theme, w int, fold bool) (lines []string, optionAt map[int]int) {
	optionAt = map[int]int{}
	add := func(i int, row string) {
		for _, l := range pane.Wrap(row, w, 4) {
			optionAt[len(lines)] = i
			lines = append(lines, l)
		}
	}
	step := m.current()
	switch {
	case m.action == nil:
		for i, a := range m.actions {
			add(i, Row(th, i == m.cursor, "", a.Title, a.Hint))
		}
		lines = append(lines, "", th.Muted(words.Say("guide.guide.up_down_choose_enter_select_click")))
	case step == nil:
		lines = append(lines, th.Muted(words.Say("guide.guide.done", "title", m.action.Title)))
	case step.Options == nil:
		lines = append(lines, th.Title(m.action.Title)+" › "+step.Prompt)
		for _, note := range strings.Split(step.Note, "\n") {
			if note != "" {
				lines = append(lines, th.Muted(note))
			}
		}
		lines = append(lines, "", m.input.View(), "")
		if step.Path {
			for _, l := range m.path.Rows(th, w, pathfinder.DefaultRows) {
				if l.Index >= 0 {
					optionAt[len(lines)] = l.Index
				}
				lines = append(lines, l.Text)
			}
			lines = append(lines, "")
		}
		if m.replace {
			lines = append(lines, th.Muted(words.Say("guide.guide.this_is_what_you_typed_before")))
		}
		if v := strings.TrimSpace(m.input.Value()); v == "" && m.input.Placeholder != "" && !step.Optional {
			// The grey text is an example, not a default Enter would accept.
			lines = append(lines, th.Muted(words.Say("guide.guide.the_grey_text_is_an_example")))
		} else if v != "" {
			if problem := m.problem(step, v); problem != "" {
				lines = append(lines, th.Failure(problem), th.Muted(nameRule()))
				if step.Suggest != nil {
					if s := step.Suggest(m.answers, v); s != "" {
						lines = append(lines, "", words.Say("guide.guide.you_can"), words.Say("guide.guide.tab_use", "name", th.Bold(s)), words.Say("guide.guide.type_another_name_instead"))
					}
				}
			} else if step.Check != nil {
				lines = append(lines, th.Success(th.Icons.Passed+words.Say("guide.guide.name_works")))
			}
		}
		if step.Path {
			lines = append(lines, th.Muted(pathfinder.Hint))
		} else {
			lines = append(lines, th.Muted(words.Say("guide.guide.enter_confirm_esc_back")))
		}
	default:
		lines = append(lines, th.Title(m.action.Title)+" › "+step.Prompt)
		if why := m.defaultWhy[step.Key]; why != "" && hasOption(step.Options, m.defaults[step.Key]) {
			lines = append(lines, th.Muted(why))
		}
		if step.Glossary && fold {
			lines = append(lines, th.Muted(wordsHint))
		} else {
			for _, note := range strings.Split(step.Note, "\n") {
				if note != "" {
					lines = append(lines, th.Muted(note))
				}
			}
		}
		lines = append(lines, "")
		if step.Filter {
			if m.filter == "" {
				narrow := words.Say("guide.guide.type_to_narrow_the_list")
				if step.Try != "" {
					narrow += words.Say("guide.guide.try", "examples", step.Try)
				}
				lines = append(lines, th.Muted(narrow), "")
			} else {
				lines = append(lines, th.Muted(words.Say("guide.guide.showing_matches_for"))+th.Bold(m.filter)+th.Muted(words.Say("guide.guide.esc_clears_it")), "")
			}
		}
		for i, o := range m.options() {
			box := ""
			if step.Multi {
				box = choicerow.Box(m.marked[i], false) + " "
			}
			hint := o.Hint
			if via := MatchedVia(o, m.filter); len(via) > 0 {
				hint = words.Say("guide.guide.has_via", "via", strings.Join(via, ", "), "hint", hint)
			}
			add(i, Row(th, i == m.cursor, box, o.Label, hint))
		}
		if step.Filter && len(m.options()) == 0 {
			lines = append(lines, th.Muted("  "+NothingMatches(step.Options, m.filter)))
			if near := Suggestions(step.Options, m.filter); len(near) > 0 {
				lines = append(lines, "", words.Say("guide.guide.you_can_2"), words.Say("guide.guide.tab_show", "name", th.Bold(near[0])), words.Say("guide.guide.backspace_edit_what_you_typed"), words.Say("guide.guide.esc_clear_the_filter_and_see"))
			}
		}
		help := words.Say("guide.guide.up_down_choose_enter_select_esc")
		if step.Filter {
			help = words.Say("guide.guide.type_to_filter_up_down_choose")
		}
		if step.Multi {
			help = words.Say("guide.guide.up_down_move_space_or_click")
			if step.Optional {
				help = words.Say("guide.guide.up_down_move_space_or_click_2")
			}
		}
		if m.notice != "" {
			lines = append(lines, "", th.Accent(m.notice))
		}
		lines = append(lines, "", th.Muted(help))
	}
	if m.err != "" && !shown(lines, m.err) {
		lines = append(lines, th.Failure(m.err))
	}
	// A long question or hint wraps onto further lines; it is not cut at the edge.
	// The option rows are wrapped already, so the row numbers of a click stay.
	var wrapped []string
	moved := map[int]int{}
	for i, l := range lines {
		if v, ok := optionAt[i]; ok {
			moved[len(wrapped)] = v
			wrapped = append(wrapped, pane.Fit(l, w))
			continue
		}
		for _, part := range pane.Wrap(l, w, 2) {
			wrapped = append(wrapped, pane.Fit(part, w))
		}
	}
	return wrapped, moved
}

// hasOption reports whether an option has the value; "" is never an option.
func hasOption(options []Option, value string) bool {
	return value != "" && slices.ContainsFunc(options, func(o Option) bool { return o.Value == value })
}

// shown reports whether a line of the screen already says text, so the same
// hint is not printed twice.
func shown(lines []string, text string) bool {
	for _, l := range lines {
		if strings.Contains(pane.Strip(l), text) {
			return true
		}
	}
	return false
}

// Row draws one selectable line: a pointer, an optional checkbox, the label
// and a dimmed hint.
func Row(th theme.Theme, selected bool, box, label, hint string) string {
	cursor := "  "
	if selected {
		cursor = th.Accent(th.Icons.Pointer + " ")
		label = th.Accent(label)
	}
	if hint != "" {
		hint = "  " + th.Muted(hint)
	}
	return cursor + box + label + hint
}

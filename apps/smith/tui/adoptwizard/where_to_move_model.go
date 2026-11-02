package adoptwizard

import (
	"errors"

	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// hasRootProject is true for a plain repository whose one project is at the
// root, where smith cannot manage it: the wizard has to ask where it goes.
func (m Model) hasRootProject() bool {
	return !m.found.Kind.Convertible() && m.found.RootProject != nil
}

// asksWhereToMove is true while that question is still to be answered.
func (m Model) asksWhereToMove() bool { return m.hasRootProject() && m.moveTo == "" }

// moveOptions are the question's options, the suggested one first (the dialog
// marks the first option "suggested"): apps/{name}, packages/{name}, libs/{name},
// another folder, and no. {name} is the text the person edits.
func moveOptions(c adoptmove.Candidate) []adoptmove.Choice {
	var first, rest []adoptmove.Choice
	for _, ch := range adoptmove.Choices(c) {
		if ch.Default {
			first = append(first, ch)
		} else {
			rest = append(rest, ch)
		}
	}
	return append(first, rest...)
}

// moveLabel is an option as the dialog prints it.
func moveLabel(ch adoptmove.Choice) string {
	switch ch.Key {
	case adoptmove.KeyOther:
		return "another folder: {name} (edit the text; it is the whole folder, such as tools/web)"
	case adoptmove.KeyNone:
		return ch.Label + ": smith adopts an empty workspace and the project stays at the root"
	}
	return ch.Key + "/{name}"
}

// askMove asks where the project at the root goes: five options and the name,
// which e edits. The suggested option is the first.
func (m Model) askMove() Model {
	c := *m.found.RootProject
	m.stage = StageMove
	m.retry = nil
	m.full = false
	var opts []protocol.FixChoiceOption
	for _, ch := range moveOptions(c) {
		opts = append(opts, protocol.FixChoiceOption{Label: moveLabel(ch)})
	}
	desc := adoptmove.Explanation(c)
	if m.moveNote != "" {
		desc = m.moveNote + "\n" + desc
	}
	fix := protocol.Fix{Description: desc, Choices: []protocol.FixChoice{{Question: adoptmove.Question, Options: opts}}}
	style := adoptStyle
	style.Back, style.Ask, style.Input = true, "where it goes", "Name"
	text := m.moveName
	if text == "" {
		text = c.Name
	}
	m.moveTo = ""
	m.fw = fixwizard.New(doctorfix.Planned{Fix: fix}, nil).Styled(style).WithText(text).Place(m.sw, m.sh)
	return m
}

// answeredMove reads the answer of the question: the option and the edited
// name. A name or folder smith cannot use asks the question again, with the
// reason above it.
func (m Model) answeredMove(answers []int) (Model, Outcome) {
	c := *m.found.RootProject
	opts := moveOptions(c)
	if len(answers) != 1 || answers[0] < 1 || answers[0] > len(opts) {
		return m.askMove(), Outcome{}
	}
	choice := opts[answers[0]-1]
	m.moveName = m.fw.Text()
	if choice.Key == adoptmove.KeyNone {
		m.moveTo, m.moveNote = adoptmove.Leave, ""
		return m.afterMove(), Outcome{}
	}
	dir, err := adoptmove.Destination(choice.Key, m.moveName)
	if err != nil {
		m.moveNote = "! " + err.Error()
		var p *problem.Error
		if errors.As(err, &p) {
			m.moveNote = "! " + p.Problem
		}
		return m.askMove(), Outcome{}
	}
	m.moveTo, m.moveNote = dir, ""
	return m.afterMove(), Outcome{}
}

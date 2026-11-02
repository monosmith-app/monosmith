package adoptwizard

import (
	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// asksInit is true for a plain repository while the decisions of adopt --init
// (line endings, CI provider, a pipeline of the person's own) are still to be
// asked. The agent files have a step of their own after the adoption.
func (m Model) asksInit() bool { return !m.found.Kind.Convertible() && !m.initAsked }

// afterMove goes on from the question where the project goes (or from the
// found screen): to the decisions of adopt --init, then to the review.
func (m Model) afterMove() Model {
	if m.asksInit() {
		return m.askInit()
	}
	return m.showConservative()
}

// askInit asks the decisions of adopt --init, each with what its answers do.
// The first answer of each is the safe one: nothing of the person's changes.
func (m Model) askInit() Model {
	m.stage = StageInit
	m.retry = nil
	m.full = false
	m.initQuestions = adoptdecision.Questions(m.flow.InitFacts())
	var qs []protocol.FixChoice
	for _, q := range m.initQuestions {
		c := protocol.FixChoice{Question: q.Prompt}
		for _, a := range q.Answers {
			c.Options = append(c.Options, protocol.FixChoiceOption{Label: a.Label + ": " + a.Hint})
		}
		qs = append(qs, c)
	}
	desc := "smith will write smith.toml and the pipelines. These choices change files of yours, so smith asks. The first answer of each changes the least; the next screen shows every file with a diff before anything is written."
	style := adoptStyle
	style.Back, style.Ask = true, "how smith sets up the repository"
	m.fw = fixwizard.New(doctorfix.Planned{Fix: protocol.Fix{Description: desc, Choices: qs}}, nil).Styled(style).Place(m.sw, m.sh)
	return m
}

// answeredInit reads the answers of the decisions and goes to the review.
func (m Model) answeredInit(answers []int) (Model, Outcome) {
	values := map[string]string{}
	for i, q := range m.initQuestions {
		if i < len(answers) && answers[i] >= 1 && answers[i] <= len(q.Answers) {
			values[q.Key] = q.Answers[answers[i]-1].Value
		}
	}
	m.initChoices = adoptdecision.FromAnswers(m.flow.InitFacts(), values)
	m.initAsked = true
	return m.showConservative(), Outcome{}
}

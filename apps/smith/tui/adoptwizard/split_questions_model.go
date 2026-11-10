package adoptwizard

import (
	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// splitQuestion is one question of the split: where a development dependency
// goes (name is the package), or what happens to the project's ESLint setup
// (name is empty).
type splitQuestion struct {
	name string
}

func (q splitQuestion) isLint() bool { return q.name == "" }

// asksSplit is true when the project moves to a folder and the questions of the
// split have not been asked yet.
func (m Model) asksSplit() bool {
	return m.hasRootProject() && m.moveTo != "" && m.moveTo != adoptmove.Leave && !m.splitAsked
}

// askSplit asks what smith cannot place when the project moves: each
// development dependency it does not know as a tool, and the project's own
// ESLint setup. The first answer of each changes the least. When there is
// nothing to ask, or the plan cannot be read (the step then reports it), the
// walk goes on.
func (m Model) askSplit() Model {
	plan, err := m.flow.MovePlan(m.moveTo)
	m.splitQuestions = nil
	if err == nil {
		for _, name := range plan.Unsure {
			m.splitQuestions = append(m.splitQuestions, splitQuestion{name: name})
		}
		if plan.Lint {
			m.splitQuestions = append(m.splitQuestions, splitQuestion{})
		}
	}
	if len(m.splitQuestions) == 0 {
		m.splitAsked = true
		return m.afterMove()
	}
	m.stage = StageSplit
	m.retry = nil
	m.full = false
	var qs []protocol.FixChoice
	for _, q := range m.splitQuestions {
		if q.isLint() {
			qs = append(qs, protocol.FixChoice{Question: words.Say("adopt.split.lint.question"), Options: []protocol.FixChoiceOption{
				{Label: words.Say("adopt.split.lint.keep")}, {Label: words.Say("adopt.split.lint.shared")},
			}})
			continue
		}
		qs = append(qs, protocol.FixChoice{Question: words.Say("adopt.split.dep.question", "name", q.name), Options: []protocol.FixChoiceOption{
			{Label: words.Say("adopt.split.dep.project")}, {Label: words.Say("adopt.split.dep.workspace")},
		}})
	}
	style := adoptStyle
	style.Back, style.Ask = true, words.Say("adopt.split.ask")
	m.fw = fixwizard.New(doctorfix.Planned{Fix: protocol.Fix{Description: words.Say("adopt.split.desc"), Choices: qs}}, nil).Styled(style).Place(m.sw, m.sh)
	return m
}

// answeredSplit reads the answers of the split and goes on to the decisions of
// adopt --init. The second answer of a question is the one that changes
// something: a dependency goes to the workspace, or the shared lint config
// replaces the project's.
func (m Model) answeredSplit(answers []int) (Model, Outcome) {
	var split adoptflow.Split
	for i, q := range m.splitQuestions {
		second := i < len(answers) && answers[i] == 2
		switch {
		case q.isLint() && second:
			split.ESLint = "shared"
		case q.isLint():
			split.ESLint = "keep"
		case second:
			split.Hoist = append(split.Hoist, q.name)
		}
	}
	m.split, m.splitAsked = split, true
	return m.afterMove(), Outcome{}
}

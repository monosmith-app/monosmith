package adoptwizard

import (
	"context"
	"errors"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Done reports the end of the job the shell ran.
func (m Model) Done(result any, err error) (Model, Outcome) {
	if err != nil {
		m.fw = m.fw.Fail(err)
		return m, Outcome{}
	}
	switch m.stage {
	case StageReading:
		flow, _ := result.(Flow)
		if flow == nil {
			m.fw = m.fw.Fail(errors.New(words.Say("adopt.error.read")))
			return m, Outcome{}
		}
		m.flow = flow
		m.found = flow.Found()
		return m.showFound(), Outcome{}
	case StageInstalling:
		flow, _ := result.(Flow)
		if flow == nil {
			m.fw = m.fw.Fail(errors.New(words.Say("adopt.error.again")))
			return m, Outcome{}
		}
		m.flow = flow
		m.found = flow.Found()
		return m.showFound(), Outcome{}
	case StageListing:
		l, _ := result.(listed)
		m.changes, m.changesDiff = l.changes, l.diff
		return m.showConservative(), Outcome{}
	case StagePreparing:
		p, _ := result.(prepared)
		m.defaults = p.defaults
		m.asked = nil
		for _, q := range p.questions {
			if len(q.Options) >= 2 {
				m.asked = append(m.asked, q)
			}
		}
		m.choices = m.defaults
		for _, q := range p.questions {
			if len(q.Options) == 1 {
				m.choices.Set(q.Key, q.Options[0].Value)
			}
		}
		if len(m.asked) == 0 {
			return m.planFull()
		}
		return m.askChoices(), Outcome{}
	case StagePlanning:
		m.preview, _ = result.(adoptflow.Preview)
		return m.reviewFull(), Outcome{}
	case StageConverting:
		m.summary, _ = result.(string)
		if m.full {
			return m.check()
		}
		return m.finish()
	case StageChecking:
		plan, _ := result.(removalplan.Plan)
		m.kept, m.notes = plan.Kept, plan.Notes
		if len(plan.Groups) == 0 {
			return m.finish()
		}
		m.rest = nil
		for _, g := range plan.Groups[1:] {
			m.rest = append(m.rest, g.ID)
		}
		return m.askGroup(plan.Groups[0]), Outcome{}
	case StageRemoving:
		m.applied = append(m.applied, m.group.ID)
		return m.check()
	case StageFinishing:
		done, _ := result.(finished)
		m.checked, m.findings = done.steps, done.findings
		return m.offerAgent(), Outcome{}
	case StageAgentInstalling:
		return m.done(), Outcome{}
	}
	return m, Outcome{}
}

// install installs the packages the repository lists and reads it again.
func (m Model) install() (Model, Outcome) {
	flow, open := m.flow, m.open
	job := Job{Name: words.Say("adopt.job.install"), Steps: 1, Run: func(ctx context.Context, out io.Writer) (any, error) {
		if err := flow.Install(ctx, out); err != nil {
			return nil, err
		}
		flow.Close()
		return open(io.Discard)
	}}
	style := adoptStyle
	style.Verb = words.Say("adopt.install.verb")
	m = m.running(StageInstalling, words.Say("adopt.install.headline"), job, withWorking(style, words.Say("adopt.install.working")))
	return m, Outcome{Job: &job}
}

// startConversion starts the conversion (or the adoption of a plain repository).
func (m Model) startConversion() (Model, Outcome) {
	flow, c := m.flow, m.choices
	if !m.found.Kind.Convertible() {
		fix := flow.AdoptFixChosen(m.commit, m.moveTo, m.initChoices, m.split)
		job := Job{Name: words.Say("adopt.job.adopt"), Steps: len(fix.Steps), Run: func(ctx context.Context, out io.Writer) (any, error) {
			return "", flow.ApplyFix(ctx, fix.Fix, out)
		}}
		style := adoptStyle
		style.NoRetry = true
		m = m.runFix(StageConverting, fix, job, style)
		return m, Outcome{Job: &job}
	}
	if !m.full {
		c = convertchoices.Choices{Level: convertchoices.LevelConservative, Lint: convertchoices.Keep, Test: convertchoices.Keep, Build: convertchoices.Smith, CI: convertchoices.Smith}
		m.choices = c
	}
	job := Job{Name: words.Say("adopt.job.convert"), Steps: 1, Run: func(ctx context.Context, out io.Writer) (any, error) {
		return flow.Convert(ctx, c, out)
	}}
	style := adoptStyle
	style.Verb, style.NoRetry = words.Say("adopt.convert.verb"), true
	m = m.running(StageConverting, words.Say("adopt.convert.headline"), job, style)
	return m, Outcome{Job: &job}
}

// check plans the next group from the files as they are.
func (m Model) check() (Model, Outcome) {
	flow, c := m.flow, m.choices
	declined := map[string]bool{}
	for k, v := range m.declined {
		declined[k] = v
	}
	job := Job{Name: words.Say("adopt.job.check"), Run: func(context.Context, io.Writer) (any, error) {
		return flow.NextGroup(c, declined)
	}}
	m = m.running(StageChecking, words.Say("adopt.check.headline"), job, withWorking(adoptStyle, words.Say("adopt.working.files")))
	return m, Outcome{Job: &job}
}

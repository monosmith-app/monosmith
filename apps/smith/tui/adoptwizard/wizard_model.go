package adoptwizard

import (
	"context"
	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"io"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// The levels, as the first question's options in order.
const (
	levelConservative = iota + 1
	levelFull
)

// Model walks a repository into a smith workspace. Keys are the fix wizard's
// (Up and Down or a digit choose, Enter confirms, Esc cancels, b goes back,
// y n s answer a group of removals); while a job runs, Ctrl+C or Esc stops it.
type Model struct {
	open Opener
	flow Flow

	stage Stage
	fw    fixwizard.Wizard
	sw    int
	sh    int
	// retry runs the job again after a failure.
	retry *Job

	found adoptflow.Found
	// full is true when the user chose the full level.
	full      bool
	defaults  convertchoices.Choices
	asked     []convertchoices.Question
	choices   convertchoices.Choices
	preview   adoptflow.Preview
	summary   string
	group     removalplan.Group
	declined  map[string]bool
	applied   []string
	declinedN []string
	// rest are the groups planned after the one being asked.
	rest  []string
	kept  []legacyinventory.Kept
	notes []string
	// checked are the closing checks that passed, with their times.
	checked []checkedStep
	// changes are the files the conversion writes, once they were worked out;
	// changesDiff shows them as diffs, the way the Doctor fix dialog does.
	changes     adoptflow.Changes
	changesDiff string
	// commit is true when the person ticked "Commit this change so it can be
	// undone" in the review of a plain repository's adoption.
	commit bool
	// commitNote is what committing the conversion said, once it ran.
	commitNote *string
	// moveTo is the answer to "where does the project at the root go?": a
	// folder, adoptmove.Leave, or "" when there is no such question or it is
	// not answered yet. moveName is the name as the person last edited it, and
	// moveNote tells why the last answer was not usable.
	moveTo, moveName, moveNote string
	// initQuestions are the decisions of adopt --init being asked; initChoices
	// the answers once given; initAsked says they were.
	initQuestions []adoptdecision.Question
	initChoices   adoptdecision.Choices
	initAsked     bool
}

// prepared is what Prepare returned.
type listed struct {
	changes adoptflow.Changes
	diff    string
}

type prepared struct {
	defaults  convertchoices.Choices
	questions []convertchoices.Question
}

// New starts the wizard on reading the repository: the Job it returns opens
// it, and the shell reports the result with Done.
func New(open Opener) (Model, Job) {
	m := Model{open: open, declined: map[string]bool{}}
	job := Job{Name: words.Say("adopt.job.reading"), Run: func(_ context.Context, out io.Writer) (any, error) {
		f, err := open(out)
		if err != nil {
			return nil, err
		}
		return f, nil
	}}
	return m.running(StageReading, words.Say("adopt.reading.headline"), job, withWorking(adoptStyle, words.Say("adopt.working.files"))), job
}

// Stage is where the walk is.
func (m Model) Stage() Stage { return m.stage }

// Found is what the repository is, once it was read.
func (m Model) Found() adoptflow.Found { return m.found }

// Running is true while a job runs.
func (m Model) Running() bool { return m.fw.Phase() == fixwizard.Applying }

// Dispose releases what the flow holds.
func (m Model) Dispose() {
	if m.flow != nil {
		m.flow.Close()
	}
}

// Place tells the wizard the screen size.
func (m Model) Place(w, h int) Model {
	m.sw, m.sh = w, h
	m.fw = m.fw.Place(w, h)
	return m
}

// Rect is where the dialog sits.
func (m Model) Rect() layout.Rect { return m.fw.Rect() }

// Lines draws the dialog.
func (m Model) Lines(th theme.Theme, tick int) []string { return m.fw.Lines(th, tick) }

// Hover highlights the button under the pointer.
func (m Model) Hover(x, y int) Model {
	m.fw = m.fw.Hover(x, y)
	return m
}

// Wheel scrolls the body.
func (m Model) Wheel(delta int) Model {
	m.fw = m.fw.Wheel(delta)
	return m
}

// Key handles a key name.
func (m Model) Key(name string) (Model, Outcome) {
	if name == "c" && m.togglesCommit() {
		return m.toggleCommit(), Outcome{}
	}
	fw, out := m.fw.Key(name)
	m.fw = fw
	return m.react(out)
}

// togglesCommit is true in the review of the changes, the step where the
// person chooses whether the result is committed.
func (m Model) togglesCommit() bool {
	return m.stage == StageReview && m.fw.Phase() == fixwizard.Reviewing
}

// toggleCommit flips the commit box and draws the review again.
func (m Model) toggleCommit() Model {
	m.commit = !m.commit
	if m.full {
		return m.reviewFull()
	}
	return m.showConservative()
}

// committed is true when the result is one commit: a plain repository's
// adoption asked to commit, or a conversion whose commit went through.
func (m Model) committed() bool {
	if !m.commit {
		return false
	}
	if !m.found.Kind.Convertible() {
		return true
	}
	return m.commitNote != nil && strings.HasPrefix(strings.TrimSpace(*m.commitNote), "Committed") // words:literal
}

// finished is the outcome that opens the workspace.
func (m Model) finished(build bool) Outcome {
	return Outcome{Close: true, Finished: words.Say("adopt.finished", "name", m.found.Name), Build: build, Committed: m.committed()}
}

// Click handles a left press.
func (m Model) Click(x, y int) (Model, Outcome) {
	if m.togglesCommit() && m.fw.Rect().Contains(x, y) && strings.Contains(m.fw.TextAt(y), words.Say("adopt.commit.click")) {
		return m.toggleCommit(), Outcome{}
	}
	fw, out := m.fw.Click(x, y)
	m.fw = fw
	return m.react(out)
}

// Progress records the step a job is on.
func (m Model) Progress(n int, name string) Model {
	m.fw = m.fw.Progress(n, m.retryTotal(), name)
	return m
}

func (m Model) retryTotal() int {
	if m.retry == nil {
		return 0
	}
	return m.retry.Steps
}

// react acts on what the fix wizard under the stage asked for.
func (m Model) react(out fixwizard.Outcome) (Model, Outcome) {
	switch {
	case out.Stop:
		return m, Outcome{Stop: true}
	case out.Copy != "":
		return m, Outcome{Copy: out.Copy}
	case out.Back:
		return m.goBack()
	case out.Done:
		return m.cancel()
	case out.Apply:
		return m.confirm()
	case out.Decline && m.stage == StageAgent:
		return m.done(), Outcome{}
	case out.Decline && m.stage == StageDone:
		return m, m.finished(true)
	case out.Decline:
		return m.decline(false)
	case out.Skip && m.stage == StageDone:
		return m.showReport(), Outcome{}
	case out.Skip:
		return m.decline(true)
	}
	if (m.stage == StageLevel || m.stage == StageMove || m.stage == StageInit || m.stage == StageChoices) && m.fw.Phase() == fixwizard.Reviewing {
		return m.answered()
	}
	return m, Outcome{}
}

// cancel closes the dialog; what was changed already stays, and is told.
func (m Model) cancel() (Model, Outcome) {
	switch m.stage {
	case StageDone:
		return m, m.finished(false)
	case StageReport, StageAgent:
		return m.done(), Outcome{}
	case StageGroup, StageChecking:
		return m, Outcome{Close: true, Note: words.Say("adopt.cancel.stopped")}
	case StageFinishing, StageRemoving, StageConverting:
		return m, Outcome{Close: true, Note: words.Say("adopt.cancel.closed")}
	}
	return m, Outcome{Close: true}
}

// goBack returns to the screen before the first question or the review.
func (m Model) goBack() (Model, Outcome) {
	switch m.stage {
	case StageLevel, StageMove:
		return m.showFound(), Outcome{}
	case StageInit:
		if m.hasRootProject() {
			return m.askMove(), Outcome{}
		}
		return m.showFound(), Outcome{}
	case StageChoices:
		return m.askLevel(), Outcome{}
	case StageReview:
		if m.full && len(m.asked) > 0 {
			return m.askChoices(), Outcome{}
		}
		if !m.found.Kind.Convertible() && m.initAsked {
			m.initAsked = false
			return m.askInit(), Outcome{}
		}
		if m.hasRootProject() {
			return m.askMove(), Outcome{}
		}
		if m.found.Full {
			return m.askLevel(), Outcome{}
		}
		return m.showFound(), Outcome{}
	}
	return m, Outcome{}
}

// confirm acts on Enter in a review, and on Retry after a failure.
func (m Model) confirm() (Model, Outcome) {
	switch m.stage {
	case StageFound:
		switch {
		case m.found.NeedsInstall:
			return m.install()
		case len(m.found.Blockers) > 0:
			return m, Outcome{Close: true}
		case m.found.Full:
			return m.askLevel(), Outcome{}
		}
		return m.reviewConservative()
	case StageReview:
		return m.startConversion()
	case StageGroup:
		g, flow := m.group, m.flow
		job := Job{Name: words.Say("adopt.job.removing", "group", g.ID), Steps: len(g.Fix.Steps), Run: func(ctx context.Context, out io.Writer) (any, error) {
			return nil, flow.ApplyFix(ctx, g.Fix.Fix, out)
		}}
		style := adoptStyle
		style.Verb = words.Say("adopt.removing.verb")
		return m.runFix(StageRemoving, g.Fix, job, style), Outcome{Job: &job}
	case StageAgent:
		fix, _ := m.flow.AgentFix()
		flow := m.flow
		job := Job{Name: words.Say("adopt.job.agent"), Steps: len(fix.Steps), Run: func(ctx context.Context, out io.Writer) (any, error) {
			return nil, flow.ApplyFix(ctx, fix.Fix, out)
		}}
		style := adoptStyle
		style.Verb = words.Say("adopt.agent.verb")
		return m.runFix(StageAgentInstalling, fix, job, style), Outcome{Job: &job}
	case StageDone:
		return m, m.finished(false)
	case StageReport:
		return m.done(), Outcome{}
	}
	// A failure: run the job again.
	if m.retry != nil {
		job := *m.retry
		return m, Outcome{Job: &job}
	}
	return m, Outcome{Close: true}
}

// decline keeps what the group would remove; skipRest also skips the groups after it.
func (m Model) decline(skipRest bool) (Model, Outcome) {
	if m.stage != StageGroup {
		return m, Outcome{}
	}
	m.declined[m.group.ID] = true
	m.declinedN = append(m.declinedN, m.group.ID)
	if skipRest {
		m.declinedN = append(m.declinedN, m.rest...)
		return m.finish()
	}
	return m.check()
}

// running shows a job that is going to run, with the style of the dialog.
func (m Model) running(stage Stage, headline string, job Job, style fixwizard.Style) Model {
	steps := make([]protocol.FixStep, job.Steps)
	p := doctorfix.Planned{Fix: protocol.Fix{Description: headline, Steps: steps}}
	m.stage = stage
	m.retry = &job
	m.fw = fixwizard.New(p, nil).Styled(style).Place(m.sw, m.sh).Start(headline, job.Steps)
	return m
}

// runFix shows a fix that is going to be applied.
func (m Model) runFix(stage Stage, p doctorfix.Planned, job Job, style fixwizard.Style) Model {
	m.stage = stage
	m.retry = &job
	m.fw = fixwizard.New(p, nil).Styled(style).Place(m.sw, m.sh).Start(p.Description, len(p.Steps))
	return m
}

var adoptStyle = fixwizard.Style{Name: words.Say("adopt.style.name"), Verb: words.Say("adopt.style.verb")}

// reviewOf makes a review screen of a text.
func (m Model) reviewOf(stage Stage, title, text string, style fixwizard.Style) Model {
	m.stage = stage
	m.retry = nil
	p := doctorfix.Planned{Fix: protocol.Fix{Description: title}}
	m.fw = fixwizard.New(p, func(doctorfix.Planned) (string, error) { return text, nil }).Styled(style).Place(m.sw, m.sh)
	return m
}

// showFound is the first screen.
func (m Model) showFound() Model {
	style := adoptStyle
	style.Review = words.Say("adopt.found.review")
	style.Apply = words.Say("adopt.found.continue")
	if len(m.found.Blockers) > 0 {
		style.Apply = words.Say("adopt.found.close")
	}
	if m.found.NeedsInstall {
		style.Apply = words.Say("adopt.found.install.apply")
	}
	return m.reviewOf(StageFound, "found", foundText(m.found), style)
}

var levelQuestion = words.Say("adopt.level.question")

// askLevel asks conservative or full.
func (m Model) askLevel() Model {
	m.stage = StageLevel
	m.retry = nil
	m.full = false
	desc := words.Say("adopt.level.desc", "source", m.found.Source, "name", strconv.Quote(m.found.Name), "projects", words.Count("adopt.project", len(m.found.Projects)))
	fix := protocol.Fix{Description: desc, Choices: []protocol.FixChoice{{
		Question: levelQuestion,
		Options: []protocol.FixChoiceOption{
			{Label: words.Say("adopt.level.conservative")},
			{Label: words.Say("adopt.level.full")},
		},
	}}}
	style := adoptStyle
	style.Back, style.Ask = true, words.Say("adopt.level.ask")
	m.fw = fixwizard.New(doctorfix.Planned{Fix: fix}, nil).Styled(style).Place(m.sw, m.sh)
	return m
}

// askChoices asks the four choices of the full level.
func (m Model) askChoices() Model {
	m.stage = StageChoices
	m.retry = nil
	m.full = true
	var qs []protocol.FixChoice
	for _, q := range m.asked {
		c := protocol.FixChoice{Question: q.Question}
		for _, o := range q.Options {
			c.Options = append(c.Options, protocol.FixChoiceOption{Label: o.Label})
		}
		qs = append(qs, c)
	}
	desc := words.Say("adopt.choices.desc")
	style := adoptStyle
	style.Back, style.Ask = true, words.Say("adopt.choices.ask")
	m.fw = fixwizard.New(doctorfix.Planned{Fix: protocol.Fix{Description: desc, Choices: qs}}, nil).Styled(style).Place(m.sw, m.sh)
	return m
}

// answered reads the answer of the level question or of the choices.
func (m Model) answered() (Model, Outcome) {
	answers := m.fw.Answers()
	if m.stage == StageMove {
		return m.answeredMove(answers)
	}
	if m.stage == StageInit {
		return m.answeredInit(answers)
	}
	if m.stage == StageLevel {
		if len(answers) == 1 && answers[0] == levelFull {
			m.full = true
			return m.prepare()
		}
		m.full = false
		return m.reviewConservative()
	}
	c := m.choices
	c.Level = convertchoices.LevelFull
	for i, q := range m.asked {
		if i < len(answers) && answers[i] >= 1 && answers[i] <= len(q.Options) {
			c.Set(q.Key, q.Options[answers[i]-1].Value)
		}
	}
	m.choices = c
	return m.planFull()
}

// prepare converts a scratch copy to learn the defaults and questions.
func (m Model) prepare() (Model, Outcome) {
	flow := m.flow
	job := Job{Name: words.Say("adopt.job.prepare"), Run: func(context.Context, io.Writer) (any, error) {
		def, qs, err := flow.Prepare()
		if err != nil {
			return nil, err
		}
		return prepared{defaults: def, questions: qs}, nil
	}}
	m = m.running(StagePreparing, words.Say("adopt.prepare.headline"), job, withWorking(adoptStyle, words.Say("adopt.prepare.working")))
	return m, Outcome{Job: &job}
}

func withWorking(s fixwizard.Style, working string) fixwizard.Style {
	s.Working = working
	return s
}

// planFull plans the removals for the choices.
func (m Model) planFull() (Model, Outcome) {
	flow, c := m.flow, m.choices
	job := Job{Name: words.Say("adopt.job.plan"), Run: func(context.Context, io.Writer) (any, error) {
		return flow.PlanFull(c, true)
	}}
	m = m.running(StagePlanning, words.Say("adopt.plan.headline"), job, withWorking(adoptStyle, words.Say("adopt.plan.working")))
	return m, Outcome{Job: &job}
}

// reviewFull shows the whole plan.
func (m Model) reviewFull() Model {
	style := adoptStyle
	style.Apply, style.Back, style.Review = words.Say("adopt.review.convert"), true, words.Say("adopt.review.change")
	text := reviewHeader + planSummaryText(m.preview.Changes, m.flow.Result()) + "\n" + m.preview.Text + commitBox(m.commit)
	return m.reviewOf(StageReview, words.Say("adopt.review.full"), text, style)
}

// reviewConservative works out the files a conservative conversion changes
// (on a scratch copy), then shows them; a plain repository shows its adoption.
func (m Model) reviewConservative() (Model, Outcome) {
	if m.asksWhereToMove() {
		return m.askMove(), Outcome{}
	}
	if !m.found.Kind.Convertible() {
		return m.afterMove(), Outcome{}
	}
	flow := m.flow
	job := Job{Name: words.Say("adopt.job.listing"), Run: func(context.Context, io.Writer) (any, error) {
		if _, _, err := flow.Prepare(); err != nil {
			return nil, err
		}
		c, err := flow.Changes()
		if err != nil {
			return nil, err
		}
		diff, err := flow.ShowChanges(c)
		if err != nil {
			diff = words.Say("adopt.review.writes.failed", "error", err.Error())
		}
		return listed{changes: c, diff: diff}, nil
	}}
	m = m.running(StageListing, words.Say("adopt.listing.headline"), job, withWorking(adoptStyle, words.Say("adopt.listing.working")))
	return m, Outcome{Job: &job}
}

// showConservative shows what the conservative level does.
func (m Model) showConservative() Model {
	text := ""
	if !m.found.Kind.Convertible() {
		fix := m.flow.AdoptFixChosen(m.commit, m.moveTo, m.initChoices)
		shown, err := m.flow.ShowFix(fix)
		if err != nil {
			shown = words.Say("adopt.show.failed", "description", fix.Description, "error", err.Error())
		}
		// The plan shows every file smith writes, and a diff of each file of yours it edits.
		shown += "\n" + m.flow.InitPlan(m.initChoices)
		text = conservativeText(m.found, shown, adoptflow.Changes{}, "", convert.Highlights{}, m.commit)
	} else {
		text = conservativeText(m.found, "", m.changes, m.changesDiff, m.flow.Result(), m.commit)
	}
	style := adoptStyle
	style.Apply, style.Back, style.Review = words.Say("adopt.review.go"), true, words.Say("adopt.review.change")
	m = m.reviewOf(StageReview, "conservative", text, style)
	m.full = false
	return m
}

// askGroup asks about one group of removals.
func (m Model) askGroup(g removalplan.Group) Model {
	m.group = g
	m.stage = StageGroup
	m.retry = nil
	flow := m.flow
	style := adoptStyle
	style.Apply, style.Decline, style.Skip = words.Say("adopt.group.apply"), words.Say("adopt.group.decline"), words.Say("adopt.group.skip")
	style.Review = words.Say("adopt.group.review")
	show := func(p doctorfix.Planned) (string, error) {
		text, err := flow.ShowFix(p)
		return groupText(g, text), err
	}
	m.fw = fixwizard.New(g.Fix, show).Styled(style).Place(m.sw, m.sh)
	return m
}

// finish installs and checks the workspace.
func (m Model) finish() (Model, Outcome) {
	flow := m.flow
	if !m.found.Kind.Convertible() {
		return m.offerAgent(), Outcome{}
	}
	fix := flow.FinishFix(m.full)
	note := new(string)
	m.commitNote = note
	commit := m.commit
	job := Job{Name: words.Say("adopt.job.finish"), Steps: len(fix.Steps), Run: func(ctx context.Context, out io.Writer) (any, error) {
		timer := newStepTimer(out)
		if err := flow.ApplyFix(ctx, fix.Fix, timer); err != nil {
			return nil, err
		}
		if commit {
			*note = flow.CommitConversion()
		}
		return timer.Steps(), nil
	}}
	style := adoptStyle
	style.Verb = words.Say("adopt.finish.verb")
	style.Working = words.Say("adopt.finish.working")
	m = m.runFix(StageFinishing, fix, job, style)
	return m, Outcome{Job: &job}
}

// offerAgent asks whether to add the MonoSmith agent, when the workspace lacks
// it and has not opted out; otherwise it goes straight to the result.
func (m Model) offerAgent() Model {
	fix, ok := m.flow.AgentFix()
	if !ok {
		return m.done()
	}
	text, err := m.flow.ShowFix(fix)
	if err != nil {
		text = words.Say("adopt.show.fix", "description", fix.Description)
	}
	style := adoptStyle
	style.Apply, style.Decline, style.Review = words.Say("adopt.agent.apply"), words.Say("adopt.agent.decline"), words.Say("adopt.agent.review")
	return m.reviewOf(StageAgent, "agent", words.Say("adopt.agent.text")+text, style)
}

// done tells what happened.
func (m Model) done() Model {
	style := adoptStyle
	style.Apply, style.Review = words.Say("adopt.done.apply"), words.Say("adopt.done.review")
	checked := checkedText(m.checked)
	if checked == "" {
		// Nothing proved that it works: offer to build everything on opening.
		style.Decline = words.Say("adopt.done.build")
	}
	if m.found.Kind.Convertible() {
		style.Skip = words.Say("adopt.done.report")
	}
	text := doneText(m.found, m.flow.Result(), checked, m.applied, m.declinedN, m.kept, m.notes, m.committed())
	if m.commit && m.commitNote != nil && !m.committed() && *m.commitNote != "" {
		text += "\n  " + strings.TrimSpace(*m.commitNote) + "\n"
	}
	return m.reviewOf(StageDone, "done", strings.TrimLeft(text, "\n"), style)
}

// showReport shows the report of the conversion, saved when it was applied.
func (m Model) showReport() Model {
	text, err := m.flow.Report()
	if err != nil || strings.TrimSpace(text) == "" {
		text = words.Say("adopt.report.missing", "file", convert.ReportFile)
	}
	style := adoptStyle
	style.Apply, style.Review = words.Say("adopt.report.back"), words.Say("adopt.report.review")
	return m.reviewOf(StageReport, "report", text, style)
}

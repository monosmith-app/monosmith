package adoptwizard

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
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
	job := Job{Name: "reading the repository", Run: func(_ context.Context, out io.Writer) (any, error) {
		f, err := open(out)
		if err != nil {
			return nil, err
		}
		return f, nil
	}}
	return m.running(StageReading, "Reading the repository…", job, fixwizard.Style{Name: "Bring in", Verb: "Bringing in…", Working: "reading the files"}), job
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
func (m Model) Rect() dialog.Rect { return m.fw.Rect() }

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
		m.commit = !m.commit
		return m.showConservative(), Outcome{}
	}
	fw, out := m.fw.Key(name)
	m.fw = fw
	return m.react(out)
}

// togglesCommit is true in the review of a plain repository's adoption, the
// only step of the wizard that can make a commit.
func (m Model) togglesCommit() bool {
	return m.stage == StageReview && !m.full && !m.found.Kind.Convertible() && m.fw.Phase() == fixwizard.Reviewing
}

// finished is the outcome that opens the workspace.
func (m Model) finished(build bool) Outcome {
	return Outcome{Close: true, Finished: "Workspace " + m.found.Name + " is ready.", Build: build, Committed: m.commit && !m.found.Kind.Convertible()}
}

// Click handles a left press.
func (m Model) Click(x, y int) (Model, Outcome) {
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
	if (m.stage == StageLevel || m.stage == StageChoices) && m.fw.Phase() == fixwizard.Reviewing {
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
		return m, Outcome{Close: true, Note: "Stopped. The conversion is in your files; git status shows it, and git restore --staged --worktree . && git clean -fd takes it back (installed packages stay installed)."}
	case StageFinishing, StageRemoving, StageConverting:
		return m, Outcome{Close: true, Note: "Closed. Whatever ran is in your files; git status shows it and git can take it back (installed packages stay installed)."}
	}
	return m, Outcome{Close: true}
}

// goBack returns to the screen before the first question or the review.
func (m Model) goBack() (Model, Outcome) {
	switch m.stage {
	case StageLevel:
		return m.showFound(), Outcome{}
	case StageChoices:
		return m.askLevel(), Outcome{}
	case StageReview:
		if m.full && len(m.asked) > 0 {
			return m.askChoices(), Outcome{}
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
		job := Job{Name: "removing " + g.ID, Steps: len(g.Fix.Steps), Run: func(ctx context.Context, out io.Writer) (any, error) {
			return nil, flow.ApplyFix(ctx, g.Fix.Fix, out)
		}}
		style := adoptStyle
		style.Verb = "Removing…"
		return m.runFix(StageRemoving, g.Fix, job, style), Outcome{Job: &job}
	case StageAgent:
		fix, _ := m.flow.AgentFix()
		flow := m.flow
		job := Job{Name: "installing the MonoSmith agent", Steps: len(fix.Steps), Run: func(ctx context.Context, out io.Writer) (any, error) {
			return nil, flow.ApplyFix(ctx, fix.Fix, out)
		}}
		style := adoptStyle
		style.Verb = "Installing the agent…"
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
			m.fw = m.fw.Fail(fmt.Errorf("could not read the repository"))
			return m, Outcome{}
		}
		m.flow = flow
		m.found = flow.Found()
		return m.showFound(), Outcome{}
	case StageInstalling:
		flow, _ := result.(Flow)
		if flow == nil {
			m.fw = m.fw.Fail(fmt.Errorf("could not read the repository again"))
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
		m.checked, _ = result.([]checkedStep)
		return m.offerAgent(), Outcome{}
	case StageAgentInstalling:
		return m.done(), Outcome{}
	}
	return m, Outcome{}
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

var adoptStyle = fixwizard.Style{Name: "Bring in", Verb: "Bringing in…"}

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
	style.Review = "what smith found"
	style.Apply = "Continue"
	if len(m.found.Blockers) > 0 {
		style.Apply = "Close"
	}
	if m.found.NeedsInstall {
		style.Apply = "Install the packages and continue"
	}
	return m.reviewOf(StageFound, "found", foundText(m.found), style)
}

const levelQuestion = "How far should smith go?"

// askLevel asks conservative or full.
func (m Model) askLevel() Model {
	m.stage = StageLevel
	m.retry = nil
	m.full = false
	desc := fmt.Sprintf("%s %q: %s", m.found.Source, m.found.Name, plural(len(m.found.Projects), "project"))
	fix := protocol.Fix{Description: desc, Choices: []protocol.FixChoice{{
		Question: levelQuestion,
		Options: []protocol.FixChoiceOption{
			{Label: "Conservative: keep what works, map what can be mapped and list the rest as things to check"},
			{Label: "Full: make smith the only tool, replacing the old lint, test, build and CI tooling and deleting what is left over, asking about each group first"},
		},
	}}}
	style := adoptStyle
	style.Back, style.Ask = true, "how far to go"
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
	desc := "Full level: smith becomes the only tool. The first option of each question is the one this repository's own tooling suggests."
	style := adoptStyle
	style.Back, style.Ask = true, "your choices"
	m.fw = fixwizard.New(doctorfix.Planned{Fix: protocol.Fix{Description: desc, Choices: qs}}, nil).Styled(style).Place(m.sw, m.sh)
	return m
}

// answered reads the answer of the level question or of the choices.
func (m Model) answered() (Model, Outcome) {
	answers := m.fw.Answers()
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
	job := Job{Name: "working out what the conversion leaves behind", Run: func(context.Context, io.Writer) (any, error) {
		def, qs, err := flow.Prepare()
		if err != nil {
			return nil, err
		}
		return prepared{defaults: def, questions: qs}, nil
	}}
	m = m.running(StagePreparing, "Working out what the conversion leaves behind…", job, withWorking(adoptStyle, "converting a scratch copy"))
	return m, Outcome{Job: &job}
}

func withWorking(s fixwizard.Style, working string) fixwizard.Style {
	s.Working = working
	return s
}

// planFull plans the removals for the choices.
func (m Model) planFull() (Model, Outcome) {
	flow, c := m.flow, m.choices
	job := Job{Name: "planning the removals", Run: func(context.Context, io.Writer) (any, error) {
		return flow.PlanFull(c, true)
	}}
	m = m.running(StagePlanning, "Planning what the choices remove…", job, withWorking(adoptStyle, "reading the converted files"))
	return m, Outcome{Job: &job}
}

// reviewFull shows the whole plan.
func (m Model) reviewFull() Model {
	style := adoptStyle
	style.Apply, style.Back, style.Review = "Convert", true, "what will change"
	text := reviewHeader + planSummaryText(m.preview.Changes, m.flow.Result()) + "\n" + m.preview.Text
	return m.reviewOf(StageReview, "full conversion", text, style)
}

// reviewConservative works out the files a conservative conversion changes
// (on a scratch copy), then shows them; a plain repository shows its adoption.
func (m Model) reviewConservative() (Model, Outcome) {
	if !m.found.Kind.Convertible() {
		return m.showConservative(), Outcome{}
	}
	flow := m.flow
	job := Job{Name: "working out which files change", Run: func(context.Context, io.Writer) (any, error) {
		if _, _, err := flow.Prepare(); err != nil {
			return nil, err
		}
		c, err := flow.Changes()
		if err != nil {
			return nil, err
		}
		diff, err := flow.ShowChanges(c)
		if err != nil {
			diff = `● The files smith writes
  could not show the changes: ` + err.Error() + `
`
		}
		return listed{changes: c, diff: diff}, nil
	}}
	m = m.running(StageListing, "Working out which files change…", job, withWorking(adoptStyle, "trying the conversion on a copy"))
	return m, Outcome{Job: &job}
}

// showConservative shows what the conservative level does.
func (m Model) showConservative() Model {
	text := ""
	if !m.found.Kind.Convertible() {
		fix := m.flow.AdoptFix(m.commit)
		shown, err := m.flow.ShowFix(fix)
		if err != nil {
			shown = "● " + fix.Description + "\n  could not show the changes: " + err.Error() + "\n"
		}
		text = conservativeText(m.found, shown, adoptflow.Changes{}, "", convert.Highlights{}, m.commit)
	} else {
		text = conservativeText(m.found, "", m.changes, m.changesDiff, m.flow.Result(), false)
	}
	style := adoptStyle
	style.Apply, style.Back, style.Review = "Go ahead", true, "what will change"
	m = m.reviewOf(StageReview, "conservative", text, style)
	m.full = false
	return m
}

// install installs the packages the repository lists and reads it again.
func (m Model) install() (Model, Outcome) {
	flow, open := m.flow, m.open
	job := Job{Name: "installing the packages", Steps: 1, Run: func(ctx context.Context, out io.Writer) (any, error) {
		if err := flow.Install(ctx, out); err != nil {
			return nil, err
		}
		flow.Close()
		return open(io.Discard)
	}}
	style := adoptStyle
	style.Verb = "Installing…"
	m = m.running(StageInstalling, "Installing the packages", job, withWorking(style, "starting the package manager"))
	return m, Outcome{Job: &job}
}

// startConversion starts the conversion (or the adoption of a plain repository).
func (m Model) startConversion() (Model, Outcome) {
	flow, c := m.flow, m.choices
	if !m.found.Kind.Convertible() {
		fix := flow.AdoptFix(m.commit)
		job := Job{Name: "adopting the repository", Steps: len(fix.Steps), Run: func(ctx context.Context, out io.Writer) (any, error) {
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
	job := Job{Name: "converting the workspace", Steps: 1, Run: func(ctx context.Context, out io.Writer) (any, error) {
		return flow.Convert(ctx, c, out)
	}}
	style := adoptStyle
	style.Verb, style.NoRetry = "Converting…", true
	m = m.running(StageConverting, "Converting the workspace to smith", job, style)
	return m, Outcome{Job: &job}
}

// check plans the next group from the files as they are.
func (m Model) check() (Model, Outcome) {
	flow, c := m.flow, m.choices
	declined := map[string]bool{}
	for k, v := range m.declined {
		declined[k] = v
	}
	job := Job{Name: "checking what the old tooling leaves", Run: func(context.Context, io.Writer) (any, error) {
		return flow.NextGroup(c, declined)
	}}
	m = m.running(StageChecking, "Checking what is left of the old tooling…", job, withWorking(adoptStyle, "reading the files"))
	return m, Outcome{Job: &job}
}

// askGroup asks about one group of removals.
func (m Model) askGroup(g removalplan.Group) Model {
	m.group = g
	m.stage = StageGroup
	m.retry = nil
	flow := m.flow
	style := adoptStyle
	style.Apply, style.Decline, style.Skip = "Yes, delete", "No, keep", "Skip rest"
	style.Review = "delete old tooling?"
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
	job := Job{Name: "installing and checking", Steps: len(fix.Steps), Run: func(ctx context.Context, out io.Writer) (any, error) {
		timer := newStepTimer(out)
		if err := flow.ApplyFix(ctx, fix.Fix, timer); err != nil {
			return nil, err
		}
		return timer.Steps(), nil
	}}
	style := adoptStyle
	style.Verb = "Installing and checking…"
	style.Working = "starting"
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
		text = "● " + fix.Description + "\n"
	}
	style := adoptStyle
	style.Apply, style.Decline, style.Review = "Add the agent", "Not now", "the MonoSmith agent"
	return m.reviewOf(StageAgent, "agent", "The MonoSmith agent lets Claude Code and VS Code work in this workspace through smith.\nIt reads on its own and asks before it changes anything.\n\n"+text, style)
}

// done tells what happened.
func (m Model) done() Model {
	style := adoptStyle
	style.Apply, style.Review = "Open the workspace", "the result"
	checked := checkedText(m.checked)
	if checked == "" {
		// Nothing proved that it works: offer to build everything on opening.
		style.Decline = "Open it and build everything"
	}
	if m.found.Kind.Convertible() {
		style.Skip = "Read the report"
	}
	text := doneText(m.found, m.flow.Result(), checked, m.applied, m.declinedN, m.kept, m.notes, m.commit && !m.found.Kind.Convertible())
	return m.reviewOf(StageDone, "done", strings.TrimLeft(text, "\n"), style)
}

// showReport shows the report of the conversion, saved when it was applied.
func (m Model) showReport() Model {
	text, err := m.flow.Report()
	if err != nil || strings.TrimSpace(text) == "" {
		text = "The report could not be read: " + convert.ReportFile + " is missing or empty.\n"
	}
	style := adoptStyle
	style.Apply, style.Review = "Back to the result", "the report"
	return m.reviewOf(StageReport, "report", text, style)
}

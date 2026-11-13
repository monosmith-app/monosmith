package tui

import (
	"context"
	"errors"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/depsrepair"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// depsRepairWalk is the dependency repair while its dialog is open: the offer
// it was read from, once it is read. It uses the fix wizard and its run (s.fix,
// s.fixing); this state says the walk is a repair, not a doctor fix.
type depsRepairWalk struct {
	// loading is true while the plugins are asked what they can repair.
	loading bool
	offer   depsrepair.Offer
}

// depsRepairStyle words the wizard for the repair. The kinds are a choice
// smith never makes: no option is marked as the suggested one.
func depsRepairStyle(noRetry bool) fixwizard.Style {
	return fixwizard.Style{
		Name:         words.Say("tui.deps_repair.name"),
		Ask:          words.Say("tui.deps_repair.ask"),
		Review:       words.Say("tui.deps_repair.review"),
		Verb:         words.Say("tui.deps_repair.verb"),
		Working:      words.Say("tui.deps_repair.working"),
		Apply:        words.Say("tui.deps_repair.apply"),
		NoRetry:      noRetry,
		NoSuggestion: true,
	}
}

// openDepsRepair opens the repair dialog at once, with a spinner, and asks the
// plugins in the background what each can repair.
func (s *Shell) openDepsRepair() tea.Cmd {
	if s.running() || s.fixing != nil {
		s.say(words.Say("tui.fix.a_run_is_in_progress_ctrl"))
		return nil
	}
	load := s.sources.RepairOffer
	if load == nil || s.sources.RepairDeps == nil {
		s.say(words.Say("tui.deps_repair.not_here"))
		return nil
	}
	reading := words.Say("tui.deps_repair.reading")
	w := fixwizard.New(doctorfix.Planned{Fix: protocol.Fix{Description: reading}}, nil).Styled(depsRepairStyle(true)).Within(s.dialogMargins()).Place(s.w, s.h).Start(reading, 0)
	s.fix = &w
	_, cancel := context.WithCancel(context.Background())
	run := &fixRun{ch: make(chan tea.Msg, 32), cancel: cancel}
	s.fixing, s.depsRepair = run, &depsRepairWalk{loading: true}
	go func() {
		defer cancel()
		offer, err := load()
		run.ch <- fixDoneMsg{run: run, err: err, value: offer}
	}()
	return waitFix(run)
}

// depsRepairWizard is the walk once the offer is read: one question with the
// four kinds, lightest first. A kind no tool has is shown and disabled, with
// the reason. Each option's steps are the steps of the repair, so the review
// and the progress list them.
func depsRepairWizard(offer depsrepair.Offer, sw, sh int) fixwizard.Wizard {
	var options []protocol.FixChoiceOption
	for _, kind := range protocol.RepairKinds() {
		o := protocol.FixChoiceOption{Label: depsrepair.KindLabel(kind) + ": " + depsrepair.KindExplanation(kind)}
		if !offer.Has(kind) {
			o.Disabled = depsrepair.NoToolText()
		}
		for _, st := range depsrepair.Steps(offer, kind) {
			o.Fix.Steps = append(o.Fix.Steps, protocol.FixStep{Description: st.Name})
		}
		options = append(options, o)
	}
	fix := protocol.Fix{
		Description: words.Say("tui.deps_repair.description"),
		Choices:     []protocol.FixChoice{{Question: words.Say("tui.deps_repair.question"), Options: options}},
	}
	show := func(p doctorfix.Planned) (string, error) {
		return depsrepair.PlanText(offer, kindOfSteps(offer, p.Steps)), nil
	}
	return fixwizard.New(doctorfix.Planned{Fix: fix}, show).Styled(depsRepairStyle(false)).Place(sw, sh)
}

// kindOfSteps is the kind whose steps are steps: the review only has the
// resolved fix, not the answer.
func kindOfSteps(offer depsrepair.Offer, steps []protocol.FixStep) protocol.RepairKind {
	for _, kind := range protocol.RepairKinds() {
		mine := depsrepair.Steps(offer, kind)
		if len(mine) != len(steps) || len(mine) == 0 {
			continue
		}
		if slices.EqualFunc(mine, steps, func(a depsrepair.Step, b protocol.FixStep) bool { return a.Name == b.Description }) {
			return kind
		}
	}
	return ""
}

// depsRepairKind is the kind the person picked in the wizard.
func (s Shell) depsRepairKind() (protocol.RepairKind, bool) {
	if s.fix == nil {
		return "", false
	}
	answers := s.fix.Answers()
	kinds := protocol.RepairKinds()
	if len(answers) == 0 || answers[0] < 1 || answers[0] > len(kinds) {
		return "", false
	}
	return kinds[answers[0]-1], true
}

// applyDepsRepair runs the picked repair in the background; each step reports
// through the wizard as a doctor fix does.
func (s *Shell) applyDepsRepair() tea.Cmd {
	kind, ok := s.depsRepairKind()
	if !ok {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	run := &fixRun{ch: make(chan tea.Msg, 32), cancel: cancel}
	s.fixing = run
	repair := s.sources.RepairDeps
	go func() {
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				run.ch <- fixDoneMsg{run: run, err: errors.New(words.Say("tui.deps_repair.crashed"))}
			}
		}()
		run.ch <- fixDoneMsg{run: run, err: repair(ctx, kind, &stepWriter{run: run})}
	}()
	return waitFix(run)
}

// depsRepairDone ends a step of the walk: the offer is read (the question
// opens), or the repair ended (a toast, and the views read again).
func (s *Shell) depsRepairDone(msg fixDoneMsg) tea.Cmd {
	s.fixing = nil
	walk := s.depsRepair
	if walk.loading {
		walk.loading = false
		if msg.err != nil {
			w := s.fix.Fail(msg.err)
			s.fix = &w
			return nil
		}
		offer, _ := msg.value.(depsrepair.Offer)
		if len(offer.Tools) == 0 {
			w := s.fix.Fail(errors.New(words.Say("tui.deps_repair.nothing")))
			s.fix = &w
			return nil
		}
		walk.offer = offer
		w := depsRepairWizard(offer, s.w, s.h).Within(s.dialogMargins()).Place(s.w, s.h)
		s.fix = &w
		return nil
	}
	refresh := s.refreshAfter([]string{"deps", "repair"})
	if msg.err != nil {
		w := s.fix.Fail(msg.err)
		s.fix = &w
		return refresh
	}
	kind, _ := s.depsRepairKind()
	s.fix, s.depsRepair = nil, nil
	s.notify(words.Say("tui.deps_repair.done", "kind", depsrepair.KindLabel(kind)), toast.Success)
	return refresh
}

package cachesetup

import (
	"maps"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
)

// entry is one step the flow took: what it was (op, see StepChoose and the
// others), what identifies it (key) and what came back.
type entry struct {
	op, key string
	// value, yes and commit are the answer of a question.
	value       string
	yes, commit bool
	// stdout, stderr and note are the result of a program or of the write.
	stdout, stderr, note string
	err                  error
	// azure is what the setup knew of the pipeline after a secret screen.
	azure azureState
}

// history is every step of one run of the setup, in order. The flow is a
// straight line of calls, so going back means: keep the steps before the
// question, start the flow again, answer those steps from the history without
// asking or running anything, and ask the question live.
type history struct {
	entries []entry
	// earlier keeps the question answers that going back dropped, by position,
	// so the question comes again with its old answer chosen.
	earlier map[int]entry
}

// backResult says what asking to go back did.
type backResult int

const (
	// wentBack: the history now ends before the earlier question.
	wentBack backResult = iota
	// refusedBack: an action ran since the earlier question.
	refusedBack
	// firstQuestion: there is no earlier question.
	firstQuestion
)

// goBack drops the steps from the earlier question on, unless an action ran
// after it: an action is a point of no return, so the history stays as it is.
func (h *history) goBack() backResult {
	prev := -1
	ranAction := false
	for i := len(h.entries) - 1; i >= 0; i-- {
		if isQuestion[h.entries[i].op] {
			prev = i
			break
		}
		ranAction = true
	}
	if prev < 0 && !ranAction {
		return firstQuestion
	}
	if prev < 0 || ranAction {
		return refusedBack
	}
	if h.earlier == nil {
		h.earlier = map[int]entry{}
	}
	for i := prev; i < len(h.entries); i++ {
		h.earlier[i] = h.entries[i]
	}
	h.entries = h.entries[:prev]
	return wentBack
}

// stepper is the Dialogue the flow talks to. It records each step; while the
// history has steps ahead (after going back) it answers from them and tells
// the person nothing, so no program runs twice and no question repeats. It
// is the only place that knows the history.
type stepper struct {
	inner Dialogue
	run   Runner
	write func(remote string, commit bool) (string, error)
	h     *history
	// pos is the next step to replay; it equals len(h.entries) when steps are live.
	pos int
	// azure is the pipeline state of this pass of the flow, saved after each secret.
	azure *azureState
	// inSecret is true while a secret screen is open: what it runs is part of that step.
	inSecret bool
}

// replay returns the recorded step at pos when it is the step the flow asks
// for. When the flow took another road than the history, the history ends here.
func (s *stepper) replay(op, key string) (entry, bool) {
	if s.pos >= len(s.h.entries) {
		return entry{}, false
	}
	if e := s.h.entries[s.pos]; e.op == op && e.key == key {
		s.pos++
		return e, true
	}
	s.h.entries = s.h.entries[:s.pos]
	return entry{}, false
}

func (s *stepper) record(e entry) {
	s.h.entries = append(s.h.entries, e)
	s.pos++
}

// allowBack tells a dialogue that offers a way back whether a question exists before this one.
func (s *stepper) allowBack() {
	b, ok := s.inner.(BackAware)
	if !ok {
		return
	}
	b.AllowBack(slices.ContainsFunc(s.h.entries[:s.pos], func(e entry) bool { return isQuestion[e.op] }))
}

// earlierAnswer is the answer given at this position before the person went
// back, when the same question comes again.
func (s *stepper) earlierAnswer(op, key string) (entry, bool) {
	e, ok := s.h.earlier[s.pos]
	return e, ok && e.op == op && e.key == key
}

// Say tells the person, except while the history replays.
func (s *stepper) Say(text string) {
	if s.pos < len(s.h.entries) {
		return
	}
	s.inner.Say(text)
}

// Choose asks, with the earlier answer chosen when the question comes again.
func (s *stepper) Choose(prompt string, options []Option, def string) string {
	if e, ok := s.replay(StepChoose, prompt); ok {
		return e.value
	}
	if e, ok := s.earlierAnswer(StepChoose, prompt); ok && slices.ContainsFunc(options, func(o Option) bool { return o.Value == e.value }) {
		def = e.value
	}
	s.allowBack()
	v := s.inner.Choose(prompt, options, def)
	s.record(entry{op: StepChoose, key: prompt, value: v})
	return v
}

// Text asks, with the earlier line as the default when the question comes again.
func (s *stepper) Text(prompt, def string) string {
	if e, ok := s.replay(StepText, prompt); ok {
		return e.value
	}
	if e, ok := s.earlierAnswer(StepText, prompt); ok {
		def = e.value
	}
	s.allowBack()
	v := s.inner.Text(prompt, def)
	s.record(entry{op: StepText, key: prompt, value: v})
	return v
}

// Confirm asks, with the earlier answer marked when the question comes again.
func (s *stepper) Confirm(c Confirmation) (bool, bool) {
	key := c.Prompt + "\x00" + c.Shown
	if e, ok := s.replay(StepConfirm, key); ok {
		return e.yes, e.commit
	}
	if e, ok := s.earlierAnswer(StepConfirm, key); ok {
		c.Earlier = &ConfirmAnswer{Yes: e.yes, Commit: e.commit}
	}
	s.allowBack()
	yes, commit := s.inner.Confirm(c)
	s.record(entry{op: StepConfirm, key: key, yes: yes, commit: commit})
	return yes, commit
}

// Working shows nothing while the history replays.
func (s *stepper) Working(text string) func() {
	if s.pos < len(s.h.entries) {
		return func() {}
	}
	return s.inner.Working(text)
}

// Show presents a page; a page that was read comes again when the person goes back to it.
func (s *stepper) Show(title string, lines []string) {
	if _, ok := s.replay(StepShow, title); ok {
		return
	}
	s.allowBack()
	s.inner.Show(title, lines)
	s.record(entry{op: StepShow, key: title})
}

// Secret opens the secret screen once. While the history replays it returns
// the recorded result and makes no secret again, so nothing secret is kept or
// replayed: the secret lived in the screen only.
func (s *stepper) Secret(noun, variable string, produce func() (string, error), actions []cachekey.KeyAction) error {
	if e, ok := s.replay(StepSecret, noun); ok {
		*s.azure = e.azure.copy()
		return e.err
	}
	s.inSecret = true
	err := s.inner.Secret(noun, variable, produce, actions)
	s.inSecret = false
	s.record(entry{op: StepSecret, key: noun, err: err, azure: s.azure.copy()})
	return err
}

// runProgram runs a program once and keeps what it printed for a replay. A
// program started by a secret screen is part of that screen's step.
func (s *stepper) runProgram(stdin string, argv ...string) (string, string, error) {
	if s.inSecret {
		return s.run(stdin, argv...)
	}
	key := CommandLine(argv)
	if e, ok := s.replay(StepRun, key); ok {
		return e.stdout, e.stderr, e.err
	}
	stdout, stderr, err := s.run(stdin, argv...)
	s.record(entry{op: StepRun, key: key, stdout: stdout, stderr: stderr, err: err})
	return stdout, stderr, err
}

// writeRemote writes the setting once; a replay returns the first result.
func (s *stepper) writeRemote(remote string, commit bool) (string, error) {
	if e, ok := s.replay(StepWrite, remote); ok {
		return e.note, e.err
	}
	note, err := s.write(remote, commit)
	s.record(entry{op: StepWrite, key: remote, note: note, err: err})
	return note, err
}

// copy is the state with its own map and slice.
func (a azureState) copy() azureState {
	a.variables = maps.Clone(a.variables)
	a.set = slices.Clone(a.set)
	return a
}

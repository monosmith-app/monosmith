package doctorview

import (
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The steps of a fix shown in the detail of Health, in their order.
const (
	// StepReview shows what is wrong and what the fix changes.
	StepReview = 1
	// StepApply runs the fix.
	StepApply = 2
	// StepCheck checks again and shows the new state of the check.
	StepCheck = 3
)

// Steps is the line of the three steps of a fix, "1 Review ── 2 Apply ── 3
// Check again": the current step in the accent and bold, the steps done in the
// success colour and the steps to come muted.
//
// Parameters: th is the theme; current is the step the fix is on (StepReview,
// StepApply or StepCheck).
// Returns: the line, not cut to a width.
func Steps(th theme.Theme, current int) string {
	labels := []string{words.Say("healthfix.step.review"), words.Say("healthfix.step.apply"), words.Say("healthfix.step.check")}
	parts := make([]string, len(labels))
	for i, label := range labels {
		text := strconv.Itoa(i+1) + " " + label
		switch n := i + 1; {
		case n == current:
			parts[i] = th.Bold(th.Accent(text))
		case n < current:
			parts[i] = th.Success(text)
		default:
			parts[i] = th.Muted(text)
		}
	}
	return strings.Join(parts, th.Border(" ── ", false))
}

// StateLine is the state of a check in words with its mark: "▲ Warning",
// "× Failed" or "✓ Passed", and its project after a dot when it names one.
//
// Parameters: th is the theme; f is the check.
// Returns: the line, not cut to a width.
func StateLine(th theme.Theme, f doctor.Finding) string {
	key := "healthfix.state.failed"
	switch {
	case f.OK:
		key = "healthfix.state.passed"
	case f.Warning:
		key = "healthfix.state.warning"
	}
	line := Mark(th, f) + " " + th.Bold(words.Say(key))
	if f.Project != "" {
		line += th.Muted(" · " + f.Project)
	}
	return line
}

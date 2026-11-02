package guide

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// askedAxes are the axes of a language's offers that a choice needs and that
// have no default: the trigger and the runtime of a function app. smith never
// picks them (ADR 0072), so the form asks each in its own question once a choice
// that needs it is ticked. They are listed in a stable order.
func askedAxes(offers []ChoiceOffer) []string {
	defaulted := map[string]bool{}
	for _, o := range offers {
		defaulted[o.Axis] = defaulted[o.Axis] || o.Default
	}
	var axes []string
	for _, o := range offers {
		for _, axis := range o.Needs {
			if !defaulted[axis] && !slices.Contains(axes, axis) && slices.ContainsFunc(offers, func(x ChoiceOffer) bool { return x.Axis == axis }) {
				axes = append(axes, axis)
			}
		}
	}
	sort.Strings(axes)
	return axes
}

// needsAsked reports whether the offer needs one of the asked axes.
func needsAsked(o ChoiceOffer, asked []string) bool {
	return slices.ContainsFunc(o.Needs, func(axis string) bool { return slices.Contains(asked, axis) })
}

// axisKey is the answer key of the question for one asked axis of a language.
func axisKey(lang, axis string) string { return withKeyPrefix + lang + ":" + axis }

// axisSteps are the questions for the asked axes of a language: one choice each,
// with the suggested one first so that enter takes it, and left out unless a
// ticked choice needs the axis. An option says which choices it does not go with.
func axisSteps(lang string, offers []ChoiceOffer) []Step {
	var steps []Step
	for _, axis := range askedAxes(offers) {
		var onAxis []ChoiceOffer
		for _, o := range offers {
			if o.Axis == axis {
				onAxis = append(onAxis, o)
			}
		}
		sort.SliceStable(onAxis, func(i, j int) bool { return onAxis[i].Suggested && !onAxis[j].Suggested })
		step := Step{
			Key:    axisKey(lang, axis),
			Prompt: fmt.Sprintf("Which %s? smith does not choose it for you (the first is suggested)", axis),
			Skip: func(a map[string][]string) bool {
				for _, picked := range a[withKeyPrefix+lang] {
					if i := slices.IndexFunc(offers, func(o ChoiceOffer) bool { return o.Name == picked }); i >= 0 && slices.Contains(offers[i].Needs, axis) {
						return false
					}
				}
				return true
			},
		}
		for _, o := range onAxis {
			hint := o.Description
			if o.Suggested {
				hint += " (suggested)"
			}
			if len(o.Conflicts) > 0 {
				hint += " (not with " + strings.Join(o.Conflicts, ", ") + ")"
			}
			step.Options = append(step.Options, Option{Label: o.Name, Value: o.Name, Hint: hint})
		}
		steps = append(steps, step)
	}
	return steps
}

package guide

import (
	"slices"
	"sort"
	"strings"
)

// SwapProject is a project of the workspace as the Swap form lists it: its
// name, and the language that decides which choices it can swap.
type SwapProject struct {
	Name, Language string
}

// Keys of the answers the Swap form collects. The axis and the choice are
// asked once per language (and axis), so their keys carry that.
const (
	swapAxisPrefix   = "axis:"
	swapChoicePrefix = "choice:"
)

// SwapArgv turns the answers of the Swap form into `smith swap`. The plain
// form asks the axis and the choice under the keys "axis" and "choice"; the
// form with pick lists asks them once per language, under prefixed keys.
//
// answers are what the person answered. It returns the arguments of smith.
func SwapArgv(answers map[string][]string) []string {
	axis, choice := first(answers, "axis"), first(answers, "choice")
	for key, v := range answers {
		if len(v) == 0 {
			continue
		}
		switch {
		case strings.HasPrefix(key, swapAxisPrefix):
			axis = v[0]
		case strings.HasPrefix(key, swapChoicePrefix):
			choice = v[0]
		}
	}
	return []string{"swap", first(answers, "project"), axis, choice}
}

// SwapProjectAction is the "Swap a choice" form with pick lists in place of
// free text: the project, then the decisions its language has, then the
// answers to that decision. A project can only swap to a choice of its own
// language.
//
// base is the plain form, offers the usable choices and projects the
// workspace's projects; a project whose language has no choice is not listed. A
// form with nothing to pick from is returned as it is.
func SwapProjectAction(base Action, offers []ChoiceOffer, projects []SwapProject) Action {
	if len(projects) == 0 || len(offers) == 0 {
		return base
	}
	offered := offerLanguages(offers)
	languageOf := map[string]string{}
	pick := Step{Key: "project", Prompt: "Which project?"}
	for _, p := range projects {
		if !slices.Contains(offered, p.Language) {
			continue // no plugin offers a choice for its language
		}
		languageOf[p.Name] = p.Language
		pick.Options = append(pick.Options, Option{Label: p.Name, Value: p.Name, Hint: p.Language})
	}
	if len(pick.Options) == 0 {
		return base
	}
	pick.Filter = len(pick.Options) > 10
	steps := []Step{pick}
	for _, lang := range offered {
		steps = append(steps, swapLanguageSteps(lang, offersOf(offers, lang), languageOf)...)
	}
	out := base
	out.Steps = steps
	return out
}

// swapLanguageSteps are the questions for the projects of one language: the
// decision, then one question for the answers of each decision. They are left
// out for a project of another language.
func swapLanguageSteps(lang string, offers []ChoiceOffer, languageOf map[string]string) []Step {
	var axes []string
	for _, o := range offers {
		if !slices.Contains(axes, o.Axis) {
			axes = append(axes, o.Axis)
		}
	}
	sort.Strings(axes)
	axisKey := swapAxisPrefix + lang
	steps := []Step{{
		Key: axisKey, Prompt: "Which decision?",
		Skip: func(a map[string][]string) bool { return languageOf[first(a, "project")] != lang },
	}}
	for _, axis := range axes {
		steps[0].Options = append(steps[0].Options, Option{Label: axis, Value: axis})
		step := Step{
			Key: swapChoicePrefix + lang + ":" + axis, Prompt: "Swap " + axis + " to which answer?",
			Skip: func(a map[string][]string) bool {
				return languageOf[first(a, "project")] != lang || first(a, axisKey) != axis
			},
		}
		for _, o := range offers {
			if o.Axis != axis {
				continue
			}
			hint := o.Description
			if o.Default {
				hint += " (default)"
			}
			step.Options = append(step.Options, Option{Label: o.Name, Value: o.Name, Hint: hint})
		}
		steps = append(steps, step)
	}
	return steps
}

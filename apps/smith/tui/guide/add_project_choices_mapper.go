package guide

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// ChoiceOffer is one choice of a project as `smith choices --json` lists it:
// a framework, a test runner or a style that a plugin publishes.
type ChoiceOffer struct {
	Axis        string   `json:"axis"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Default     bool     `json:"default"`
	Requires    []string `json:"requires"`
	Conflicts   []string `json:"conflicts"`
	// Problem says why the choice cannot be used here; empty when it can.
	Problem string `json:"problem"`
}

// ParseChoices reads the output of `smith choices --json`.
//
// text is that output. It returns the offers that can be used, and an error
// when the text is not the JSON list smith prints.
func ParseChoices(text string) ([]ChoiceOffer, error) {
	var all []ChoiceOffer
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &all); err != nil {
		return nil, fmt.Errorf("smith choices did not print a list of choices: %w", err)
	}
	var usable []ChoiceOffer
	for _, o := range all {
		if o.Problem == "" {
			usable = append(usable, o)
		}
	}
	return usable, nil
}

// withKeyPrefix starts the answer key of the choices asked for one language.
const withKeyPrefix = "with:"

// siblingsMin is how many projects a workspace needs before the form asks
// whether the new project should match them: smith follows siblings only from
// three projects of one kind and language on.
const siblingsMin = 3

// AddProjectAction is the "Add a project" form with the questions the
// workspace's plugins add: after the name, the choices (a framework, a test
// runner, a style) of the language picked, and, when the workspace already
// has projects to follow, whether the new one matches them or takes the
// stack's own defaults (`--fresh`).
//
// base is the plain form, offers the usable choices and existing the names of
// the projects the workspace has: a name that one of them has is refused as
// the person types it. The questions are only added when there is something
// to ask.
func AddProjectAction(base Action, offers []ChoiceOffer, existing []string) Action {
	steps := slices.Clone(base.Steps)
	for i := range steps {
		if steps[i].Key == "name" && steps[i].Check != nil {
			check := steps[i].Check
			steps[i].Check = func(a map[string][]string, value string) string {
				if slices.Contains(existing, value) {
					return fmt.Sprintf("%s exists already: pick another name", value)
				}
				return check(a, value)
			}
		}
	}
	base.Steps = slices.Clone(steps)
	for _, lang := range offerLanguages(offers) {
		steps = append(steps, choicesStep(lang, offersOf(offers, lang)))
	}
	if len(existing) >= siblingsMin {
		steps = append(steps, Step{Key: "follow", Prompt: "Should it do what the other projects here do?", Options: []Option{
			{Label: "Yes, match the other projects", Value: "follow", Hint: "when most projects of this kind share a test runner or a build tool, so does this one"},
			{Label: "No, use the stack's own defaults", Value: "fresh", Hint: "smith add --fresh"},
		}})
	}
	if len(steps) == len(base.Steps) {
		return base
	}
	out := base
	out.Steps = steps
	out.Argv = func(a map[string][]string) []string {
		argv := base.Argv(a)
		var picked []string
		for _, lang := range offerLanguages(offers) {
			picked = append(picked, a[withKeyPrefix+lang]...)
		}
		if len(picked) > 0 {
			argv = append(argv, "--with", strings.Join(picked, ","))
		}
		if first(a, "follow") == "fresh" {
			argv = append(argv, "--fresh")
		}
		return argv
	}
	return out
}

// RemoveProjectAction is the "Remove a project" form with a list of the
// projects to pick from, in place of a free-text name that can be mistyped.
// names are the workspace's projects; a form with no project is returned as
// it is.
func RemoveProjectAction(base Action, names []string) Action {
	if len(names) == 0 {
		return base
	}
	pick := Step{Key: "name", Prompt: "Which project?", Filter: len(names) > 10}
	for _, n := range names {
		pick.Options = append(pick.Options, Option{Label: n, Value: n})
	}
	out := base
	out.Steps = slices.Clone(base.Steps)
	for i := range out.Steps {
		if out.Steps[i].Key == "name" {
			out.Steps[i] = pick
		}
	}
	return out
}

// choicesStep asks for the choices of one language, and is left out for any
// other language.
func choicesStep(lang string, offers []ChoiceOffer) Step {
	step := Step{
		Key:      withKeyPrefix + lang,
		Prompt:   "Build it from choices? (space ticks; enter with nothing ticked uses the stack's defaults)",
		Multi:    true,
		Optional: true,
		Skip:     func(a map[string][]string) bool { return first(a, "lang") != lang },
	}
	for _, o := range offers {
		hint := o.Description
		if len(o.Requires) > 0 {
			hint += " (needs " + strings.Join(o.Requires, ", ") + ")"
		}
		if o.Default {
			hint += " (default)"
		}
		step.Options = append(step.Options, Option{Label: o.Axis + ": " + o.Name, Value: o.Name, Hint: hint})
	}
	return step
}

// offerLanguages are the languages that have choices, in a stable order.
func offerLanguages(offers []ChoiceOffer) []string {
	var langs []string
	for _, o := range offers {
		if !slices.Contains(langs, o.Language) {
			langs = append(langs, o.Language)
		}
	}
	sort.Strings(langs)
	return langs
}

// offersOf are the offers of one language, by axis and then name.
func offersOf(offers []ChoiceOffer, lang string) []ChoiceOffer {
	var out []ChoiceOffer
	for _, o := range offers {
		if o.Language == lang {
			out = append(out, o)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Axis != out[j].Axis {
			return out[i].Axis < out[j].Axis
		}
		return out[i].Name < out[j].Name
	})
	return out
}

package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The answers of the outcome-first questions.
const (
	outcomeKey = "outcome"
	pickKey    = "pick"
	// everyLanguage is the outcome that keeps the old path: kind first, then a
	// language from the full list.
	everyLanguage = "every"
	// pickNone is the answer "I decide in the next question".
	pickNone = "none"
)

// withOutcomes puts "What are you building?" in front of the kind and language
// questions of the Add form. The outcomes come from the catalogue the offers
// and the languages make (compose.Outcomes), so the wizard and the composer
// agree. An outcome answers the kind; the language is asked only when the
// outcome leaves more than one; a website, a function app or a desktop app asks
// which framework, host or shell it uses, and the answer starts ticked in the
// choices question after the name. The old path stays as the last answer.
//
// steps are the plain steps (kind, language, name); offers are the usable
// choices. The steps come back unchanged when there is no kind or language
// question.
func withOutcomes(steps []Step, offers []ChoiceOffer) []Step {
	ki := slices.IndexFunc(steps, func(s Step) bool { return s.Key == "kind" })
	li := slices.IndexFunc(steps, func(s Step) bool { return s.Key == "lang" })
	if ki < 0 || li < 0 {
		return steps
	}
	kind, lang := steps[ki], steps[li]
	outcomes := compose.Outcomes(catalogueItems(lang.Options, offers))
	byID := map[string]compose.Outcome{}
	ask := Step{
		Key: outcomeKey, Prompt: words.Say("guide.add_project_outcomes.what_are_you_building"), Filter: true,
		Note: words.Say("guide.add_project_outcomes.not_sure_pick_the_closest_one"),
		Try:  words.Say("guide.add_project_outcomes.react_azure_function_electron_tauri_api"),
	}
	for _, o := range outcomes {
		byID[o.ID] = o
		ask.Options = append(ask.Options, Option{Label: o.Title, Value: o.ID, Hint: outcomeHint(o), Also: o.Words})
	}
	ask.Options = append(ask.Options, Option{Label: words.Say("guide.add_project_outcomes.i_know_the_language_show_every"), Value: everyLanguage,
		Hint: words.Say("guide.add_project_outcomes.the_long_way_pick_app_library"), Also: []string{"language", "other", "all"}})

	kind.Auto = func(a map[string][]string) string { return byID[first(a, outcomeKey)].Kind }
	lang.Skip = func(a map[string][]string) bool { return first(a, outcomeKey) != everyLanguage }
	out := []Step{ask, kind, lang}
	for _, o := range outcomes {
		out = append(out, outcomeLanguageStep(o, lang.Options))
	}
	for _, o := range outcomes {
		for _, l := range o.Languages {
			if step, ok := outcomePickStep(o, l); ok {
				out = append(out, step)
			}
		}
	}
	for i, s := range steps {
		if i != ki && i != li {
			out = append(out, s)
		}
	}
	return out
}

// outcomeHint is the second part of a list line: what people usually pick, then
// what the outcome means.
func outcomeHint(o compose.Outcome) string {
	typical := strings.Join(o.Typical, ", ")
	if len(o.Typical) > 5 {
		typical = strings.Join(o.Typical[:5], ", ") + "…"
	}
	if typical == "" {
		return o.Consequence
	}
	return typical + ". " + o.Consequence
}

// outcomeLanguageStep asks the language of an outcome, among the languages that
// fit it; it answers itself when only one fits.
func outcomeLanguageStep(o compose.Outcome, all []Option) Step {
	var options []Option
	for _, opt := range all {
		if slices.Contains(o.Languages, opt.Value) {
			options = append(options, opt)
		}
	}
	step := Step{Key: "lang", Prompt: words.Say("guide.add_project_outcomes.which_language_is_it_written_in"), Options: options, Filter: len(options) > 10,
		Skip: func(a map[string][]string) bool { return first(a, outcomeKey) != o.ID }}
	if len(options) == 1 {
		only := options[0].Value
		step.Auto = func(map[string][]string) string { return only }
	}
	return step
}

// outcomePickStep asks which framework, desktop shell or cloud host an outcome
// uses, for one language; ok is false for an outcome with nothing to pick. A
// website and a desktop app may decide later; a function app may not, because
// without a host it is not a function app.
func outcomePickStep(o compose.Outcome, language string) (Step, bool) {
	var options []Option
	for _, it := range o.Items {
		if it.Language == language {
			options = append(options, Option{Label: it.Title, Value: it.Name, Hint: it.Description})
		}
	}
	if len(options) == 0 {
		return Step{}, false
	}
	prompt := map[string]string{"framework": words.Say("guide.add_project_outcomes.which_framework"), "shell": words.Say("guide.add_project_outcomes.which_desktop_shell"), "host": words.Say("guide.add_project_outcomes.which_cloud")}[o.Layer]
	if prompt == "" {
		prompt = words.Say("guide.add_project_outcomes.which_one")
	}
	if o.Layer != "host" {
		// First, so a first-timer's Enter ticks nothing, and Down reads the list in order.
		options = append([]Option{{Label: words.Say("guide.add_project_outcomes.decide_in_the_next_question"), Value: pickNone, Hint: words.Say("guide.add_project_outcomes.you_can_still_tick_a_framework")}}, options...)
	}
	return Step{
		Key: pickKey, Prompt: prompt, Options: options,
		Note: words.Say("guide.add_project_outcomes.pick_note", "what", WhatIs(o.Layer)),
		Skip: func(a map[string][]string) bool { return first(a, outcomeKey) != o.ID || first(a, "lang") != language },
		Preselect: func(typed string, opts []Option) string {
			for _, opt := range opts {
				if typed != "" && opt.Value != pickNone && MatchesFilter(Option{Label: opt.Label, Value: opt.Value}, typed) {
					return opt.Value
				}
			}
			// Nothing was typed: the cursor stays on the first row, "Decide in the
			// next question", so Enter ticks nothing.
			return ""
		},
	}, true
}

// catalogueItems are the items of the catalogue the Add form knows: the
// languages of its language question and the choices that are on offer.
func catalogueItems(languages []Option, offers []ChoiceOffer) []compose.Item {
	var items []compose.Item
	for _, l := range languages {
		items = append(items, compose.Item{ID: compose.KindLanguage + "/" + l.Value, Kind: compose.KindLanguage, Name: l.Value, Title: l.Label, Layer: "language", Language: l.Value})
	}
	for _, o := range offers {
		title := o.Title
		if title == "" {
			title = o.Name
		}
		items = append(items, compose.Item{
			ID: o.Language + "/" + o.Name, Kind: compose.KindChoice, Name: o.Name, Title: title, Description: o.Description,
			Aliases: o.Aliases, Layer: compose.LayerOf(o.Layer, o.Axis), Axis: o.Axis, Language: o.Language, Needs: o.Needs,
		})
	}
	return items
}

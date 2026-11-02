package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/compose"
)

// ChoicesInWords says the choices of a new project in words, and marks the
// ones the catalogue gives by default, so a person who pressed Enter through
// the questions can see what they got: "React, ESLint (default), Vitest
// (default); no desktop shell, no style".
//
// Parameters: offers are the choices smith lists; lang is the language of the
// project; names are the choices picked (the values of --with).
// Returns: the sentence without a final stop; "" when nothing was picked.
func ChoicesInWords(offers []ChoiceOffer, lang string, names []string) string {
	own := offersOf(offers, lang)
	var chosen []string
	taken := map[string]bool{}
	for _, name := range names {
		i := slices.IndexFunc(own, func(o ChoiceOffer) bool { return o.Name == name })
		if i < 0 {
			chosen = append(chosen, name)
			continue
		}
		o := own[i]
		taken[o.Axis] = true
		title := o.Title
		if title == "" {
			title = o.Name
		}
		if o.Default {
			title += " (default)"
		}
		chosen = append(chosen, title)
	}
	if len(chosen) == 0 {
		return ""
	}
	asked := askedAxes(own)
	defaulted := map[string]bool{}
	many := map[string]bool{}
	var axes []string
	for _, o := range own {
		defaulted[o.Axis] = defaulted[o.Axis] || o.Default
		many[o.Axis] = many[o.Axis] || o.Many
		if !slices.Contains(axes, o.Axis) {
			axes = append(axes, o.Axis)
		}
	}
	text := strings.Join(chosen, ", ")
	var without []string
	for _, axis := range axes {
		if taken[axis] || defaulted[axis] || many[axis] || slices.Contains(asked, axis) {
			continue
		}
		without = append(without, "no "+strings.ToLower(compose.TitleOfLayer(axis)))
	}
	if len(without) > 0 {
		text += "; " + strings.Join(without, ", ")
	}
	return text
}

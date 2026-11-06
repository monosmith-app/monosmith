package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// glossary holds the catalog key of the words the forms use before a person has
// read anything, each in one sentence. They are the first sentences of docs/reference/glossary.md
// (a test keeps the terms there); the layers (framework, shell, host) come from
// the composer's catalogue, so the two say the same.
var glossary = map[string]string{
	"stack":  "guide.what_is.stack",
	"slice":  "guide.what_is.slice",
	"plugin": "guide.what_is.plugin",
	"recipe": "guide.what_is.recipe",
}

// WhatIs is the one line that says what a word of smith means, to show where
// the word first appears.
//
// Parameters: term is stack, slice, plugin, recipe, or a layer of the
// composer (framework, shell, host, test, lint, style).
// Returns: the sentence, or "" for a word that has none.
func WhatIs(term string) string {
	if line, ok := glossary[term]; ok {
		return words.Say(line)
	}
	if about := compose.AboutLayer(term); about != "" {
		return term + " = " + about + "."
	}
	return ""
}

// choiceWords explains the words of the choices question: the stack, and each
// layer that the list shows (framework, test, lint, style, shell, host).
func choiceWords(offers []ChoiceOffer) string {
	var layers []string
	for _, o := range offers {
		if l := compose.LayerOf(o.Layer, o.Axis); !slices.Contains(layers, l) && WhatIs(l) != "" {
			layers = append(layers, l)
		}
	}
	lines := []string{WhatIs("stack")}
	for _, l := range layers {
		lines = append(lines, WhatIs(l))
	}
	return strings.Join(lines, "\n")
}

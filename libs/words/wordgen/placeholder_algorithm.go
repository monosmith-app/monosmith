package main

import (
	"regexp"
	"strconv"
)

// placeholder matches a {name} the catalogue fills by name.
var placeholder = regexp.MustCompile(`\{([a-z][A-Za-z0-9]*)\}`)

// Placeholders are the names in text, in order of first appearance, each once.
// The order is the argument order of a positional target: the English text of a
// key sets it, and every translation of the key uses the same positions.
func Placeholders(text string) []string {
	var names []string
	seen := map[string]bool{}
	for _, m := range placeholder.FindAllStringSubmatch(text, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			names = append(names, m[1])
		}
	}
	return names
}

// Positional turns each {name} of text into {n}, where n is the position of name
// in order (the names of the English text). A name that is not in order stays as
// written; Resolve has already refused such a text.
func Positional(text string, order []string) string {
	return placeholder.ReplaceAllStringFunc(text, func(whole string) string {
		name := placeholder.FindStringSubmatch(whole)[1]
		for i, n := range order {
			if n == name {
				return "{" + strconv.Itoa(i) + "}"
			}
		}
		return whole
	})
}

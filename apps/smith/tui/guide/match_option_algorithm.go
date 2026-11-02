package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
)

// MatchesFilter reports whether an option is on offer for what was typed: every
// word typed is found in its label, value, hint or one of its Also names, in
// any case ("c sharp" and "dotnet" both find C#, "react" finds Node).
func MatchesFilter(o Option, typed string) bool {
	text := strings.ToLower(o.Label + " " + o.Value + " " + o.Hint + " " + strings.Join(o.Also, " "))
	for _, word := range strings.Fields(strings.ToLower(typed)) {
		if !strings.Contains(text, word) {
			return false
		}
	}
	return true
}

// MatchedVia lists the Also names that made an option match what was typed,
// when its label, value and hint alone do not match it. It is empty for an
// option that matches by itself, so the list shows why only where it is not
// plain.
func MatchedVia(o Option, typed string) []string {
	own := o
	own.Also = nil
	if MatchesFilter(own, typed) {
		return nil
	}
	var via []string
	for _, word := range strings.Fields(strings.ToLower(typed)) {
		for _, name := range o.Also {
			if strings.Contains(strings.ToLower(name), word) && !slices.Contains(via, name) {
				via = append(via, name)
			}
		}
	}
	return via
}

// Suggestions are the closest names the options know (labels and Also names)
// to what was typed, nearest first; none when nothing is close.
func Suggestions(options []Option, typed string) []string {
	var names []string
	for _, o := range options {
		names = append(names, strings.ToLower(o.Label))
		for _, a := range o.Also {
			if !slices.Contains(names, strings.ToLower(a)) {
				names = append(names, strings.ToLower(a))
			}
		}
	}
	return nearmatch.Closest(strings.TrimSpace(typed), names)
}

// NothingMatches is the line shown when what was typed matches no option: it
// says so and suggests the closest names the options know (labels and Also
// names), or a few names to try when none is close.
func NothingMatches(options []Option, typed string) string {
	line := "nothing matches \"" + typed + "\"."
	if near := Suggestions(options, typed); len(near) > 0 {
		return line + " Did you mean " + strings.Join(near, " or ") + "?"
	}
	var some []string
	for _, o := range options {
		some = append(some, o.Also...)
		if len(some) >= 3 {
			break
		}
	}
	if len(some) >= 3 {
		return line + " Try a language or a framework, such as " + strings.Join(some[:3], ", ") + "."
	}
	return line + " Try a language name."
}

package guide

import "strings"

// MatchesFilter reports whether an option is on offer for what was typed: every
// word typed is found in its label, value or hint, in any case ("c sharp" and
// "dotnet" both find C#).
func MatchesFilter(o Option, typed string) bool {
	text := strings.ToLower(o.Label + " " + o.Value + " " + o.Hint)
	for _, word := range strings.Fields(strings.ToLower(typed)) {
		if !strings.Contains(text, word) {
			return false
		}
	}
	return true
}

package projectlist

import "strings"

// changedWord is the filter word that keeps the projects the current changes touch.
const changedWord = "changed" // words:literal

// Matches reports whether a project passes a filter. The filter is words
// separated by spaces, and the project must pass every word: "kind:app" keeps
// the kinds that start with "app", "lang:go" the languages that start with
// "go", "changed" the projects the current changes touch, and any other word
// keeps the projects whose name, kind or language holds it. Case does not
// count. An empty filter keeps every project.
//
// Parameters: p is the project; filter is the text the person typed.
// Returns: true when the project is listed.
func Matches(p Project, filter string) bool {
	for _, word := range strings.Fields(strings.ToLower(filter)) {
		if !matchesWord(p, word) {
			return false
		}
	}
	return true
}

// matchesWord is one word of a filter against one project.
func matchesWord(p Project, word string) bool {
	name, kind, lang := strings.ToLower(p.Name), strings.ToLower(p.Kind), strings.ToLower(p.Language)
	if v, ok := strings.CutPrefix(word, "kind:"); ok {
		return strings.HasPrefix(kind, v)
	}
	if v, ok := strings.CutPrefix(word, "lang:"); ok {
		return strings.HasPrefix(lang, v)
	}
	if word == changedWord && p.Affected {
		return true
	}
	return strings.Contains(name, word) || strings.Contains(kind, word) || strings.Contains(lang, word)
}

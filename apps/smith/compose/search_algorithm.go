package compose

import (
	"sort"
	"strings"
)

// Hit is an item that matched a search, with why.
type Hit struct {
	Item
	// Score ranks the hit; a higher score is a closer match.
	Score int `json:"score"`
	// MatchedBy names where the words were found: "name", "title", "alias",
	// "description" or "language".
	MatchedBy string `json:"matchedBy"`
}

// Search finds the items that every word of the query matches, ranked: a name
// or title match first, then an alias, then the description. Matching is
// case-insensitive and by prefix of a word, so "desk" finds "desktop". An empty
// query matches every item, in catalogue order.
//
// Parameters: items are the catalogue's items; query is the words typed.
// Returns: the hits, best first; equal scores keep the catalogue order.
func Search(items []Item, query string) []Hit {
	words := strings.Fields(strings.ToLower(query))
	var hits []Hit
	for _, it := range items {
		points, by, ok := score(it, words)
		if ok {
			hits = append(hits, Hit{Item: it, Score: points, MatchedBy: by})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	return hits
}

// score rates an item against all words; every word must match somewhere.
func score(it Item, words []string) (int, string, bool) {
	if len(words) == 0 {
		return 0, "", true
	}
	total, best, bestScore := 0, "", -1
	for _, w := range words {
		s, by := scoreWord(it, w)
		if s == 0 {
			return 0, "", false
		}
		total += s
		if s > bestScore {
			best, bestScore = by, s
		}
	}
	return total, best, true
}

// scoreWord rates one word against the fields of an item, and names the field.
func scoreWord(it Item, w string) (int, string) {
	switch {
	case strings.ToLower(it.Name) == w || strings.ToLower(it.Title) == w:
		return 100, "name"
	case hasPrefixWord(it.Name, w) || hasPrefixWord(it.Title, w):
		return 80, "title"
	}
	for _, a := range it.Aliases {
		if hasPrefixWord(a, w) {
			return 50, "alias"
		}
	}
	if it.Language != "" && it.Kind != KindLanguage && hasPrefixWord(it.Language, w) {
		return 30, "language"
	}
	if hasPrefixWord(it.Layer, w) {
		return 25, "layer"
	}
	if hasPrefixWord(it.Description, w) {
		return 10, "description"
	}
	return 0, ""
}

// hasPrefixWord reports whether a word of the text starts with w. Words split
// at anything that is not a letter or a digit, so "Node.js 24" has node, js and 24.
func hasPrefixWord(text, w string) bool {
	text = strings.ToLower(text)
	if strings.HasPrefix(text, w) {
		return true
	}
	for _, field := range strings.FieldsFunc(text, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		if strings.HasPrefix(field, w) {
			return true
		}
	}
	return false
}

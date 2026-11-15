package palette

import (
	"sort"
	"strings"
)

// closestLimit is how many commands the palette offers when nothing matches.
const closestLimit = 3

// closest returns up to closestLimit of the usable entries that sound most like
// the query, for the "Did you mean" list. It compares each word of the query
// with the words of the title, the keywords and the hint, and tolerates typos.
// An entry whose words are all too far from the query is left out.
func closest(items []Entry, query string) []Entry {
	type scored struct {
		entry Entry
		score int
	}
	qwords := strings.Fields(strings.ToLower(query))
	var found []scored
	for _, e := range items {
		if e.unavailable() {
			continue
		}
		vocab := strings.Fields(strings.ToLower(e.Title + " " + e.Keywords + " " + e.Section))
		total, ok := 0, true
		for _, q := range qwords {
			best := len(q)
			for _, v := range vocab {
				best = min(best, wordDistance(q, strings.Trim(v, ",.…:()")))
			}
			// More than a third of the word wrong is no match.
			if best*3 > len(q) {
				ok = false
				break
			}
			total += best
		}
		if ok {
			found = append(found, scored{e, total})
		}
	}
	sort.SliceStable(found, func(i, j int) bool { return found[i].score < found[j].score })
	var out []Entry
	for _, f := range found[:min(len(found), closestLimit)] {
		out = append(out, f.entry)
	}
	return out
}

// wordDistance is the edit distance between a typed word and the start of a
// word of a command, as long as the typed word. A typed word that starts the
// command word costs nothing.
func wordDistance(typed, word string) int {
	if strings.HasPrefix(word, typed) {
		return 0
	}
	if len(word) > len(typed) {
		word = word[:len(typed)]
	}
	prev := make([]int, len(word)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(typed); i++ {
		cur := make([]int, len(word)+1)
		cur[0] = i
		for j := 1; j <= len(word); j++ {
			cost := 1
			if typed[i-1] == word[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(word)]
}

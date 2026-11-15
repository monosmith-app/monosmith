package dependencies

import (
	"strings"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// UpdateSummary counts the updates and splits the count by level, in the order
// of Levels: "3 updates: 1 patch, 2 minor". A level that has no update is left
// out, and a level smith does not know comes after the known ones.
//
// Parameters: updates are the updates the summary tells.
// Returns: the sentence, or "" when there is no update.
func UpdateSummary(updates []Update) string {
	if len(updates) == 0 {
		return ""
	}
	counts := map[string]int{}
	order := append([]string(nil), Levels...)
	for _, u := range updates {
		if levelIndex(u.Level) == len(Levels) && counts[u.Level] == 0 {
			order = append(order, u.Level)
		}
		counts[u.Level]++
	}
	var parts []string
	for _, level := range order {
		if n := counts[level]; n > 0 {
			parts = append(parts, words.Say("dependencies.up.level", "count", n, "level", level))
		}
	}
	return words.Count("dependencies.up.summary", len(updates), "levels", strings.Join(parts, words.Say("dependencies.up.level_separator")))
}

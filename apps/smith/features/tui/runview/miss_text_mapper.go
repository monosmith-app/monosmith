package runview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/runevents"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// missWhat are the catalog keys of the kinds of input that changed.
var missWhat = map[string]string{
	"command":    "runview.miss.what.command",
	"env":        "runview.miss.what.env",
	"upstream":   "runview.miss.what.upstream",
	"dependency": "runview.miss.what.dependency",
	"tool":       "runview.miss.what.tool",
	"toolchain":  "runview.miss.what.toolchain",
	"serve":      "runview.miss.what.serve",
	"platform":   "runview.miss.what.platform",
}

// MissText says in words why a task ran instead of being restored from the
// cache: "2 files changed", "the command changed", "no entry yet".
//
// Parameters: m is the reason of the event stream, nil when smith gave none.
// Returns: the words, "" for no reason.
func MissText(m *runevents.Miss) string {
	if m == nil {
		return ""
	}
	switch m.Reason {
	case runevents.MissNoEntry:
		return words.Say("runview.miss.no_entry")
	case runevents.MissKey:
		return words.Say("runview.miss.key")
	case runevents.MissOff:
		return words.Say("runview.miss.off")
	case runevents.MissInputs:
		var parts []string
		if m.Files > 0 {
			parts = append(parts, words.Count("runview.miss.files", m.Files))
		}
		for _, what := range m.Changed {
			if key, ok := missWhat[what]; ok {
				parts = append(parts, words.Say(key))
			}
		}
		if len(parts) == 0 {
			return words.Say("runview.miss.inputs")
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

package convertchoices

import (
	"encoding/json"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

// questionJSON is one question of smith convert as `smith convert --questions
// --json` prints it, in the shape of `smith new --questions --json`, so an
// editor reads both the same way.
type questionJSON struct {
	Key     string            `json:"key"`
	Prompt  string            `json:"prompt"`
	Default string            `json:"default"`
	Flag    string            `json:"flag"`
	When    map[string]string `json:"when,omitempty"`
	Options []optionJSON      `json:"options"`
}

// optionJSON is one answer to a question; the first of a list is the default.
type optionJSON struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// QuestionsJSON renders the questions of smith convert as indented JSON: the
// level first, then the four choices that apply to the full level (their
// `when` says so). It looks at no folder: the answers are the ones a workspace
// that uses everything has, and each default is the one smith suggests when it
// knows nothing about the workspace. The plan of `smith convert --level full
// --dry-run` shows what the workspace's own tooling suggests.
func QuestionsJSON() ([]byte, error) {
	def := Defaults(legacyinventory.Usage{})
	out := []questionJSON{{
		Key:     "level",
		Prompt:  "How far should the conversion go?",
		Default: string(LevelConservative),
		Flag:    "level",
		Options: []optionJSON{
			{string(LevelConservative), "conservative: keep the build and lint setup that works, and list what needs manual work"},
			{string(LevelFull), "full: make smith the only tool; the old tooling is deleted after the plan is shown and you say yes"},
		},
	}}
	for _, q := range Questions(legacyinventory.Usage{ESLint: true}, def) {
		item := questionJSON{
			Key:     q.Key,
			Prompt:  q.Question,
			Flag:    strings.TrimPrefix(q.Flag, "--"),
			When:    map[string]string{"level": string(LevelFull)},
			Options: make([]optionJSON, 0, len(q.Options)),
		}
		for _, o := range q.Options {
			item.Options = append(item.Options, optionJSON(o))
		}
		if len(item.Options) > 0 {
			item.Default = item.Options[0].Value
		}
		out = append(out, item)
	}
	return json.MarshalIndent(out, "", "  ")
}

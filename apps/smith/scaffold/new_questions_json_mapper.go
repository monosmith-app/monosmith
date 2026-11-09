package scaffold

import "encoding/json"

// newQuestionJSON is one question of smith new as `smith new --questions --json`
// prints it. The questions come from NewQuestions, the data the CLI and the
// TUI form ask from, so an editor draws the same questions without a list of
// its own.
type newQuestionJSON struct {
	Key      string            `json:"key"`
	Prompt   string            `json:"prompt"`
	Default  string            `json:"default"`
	Flag     string            `json:"flag,omitempty"`
	Many     bool              `json:"many,omitempty"`
	Advanced bool              `json:"advanced,omitempty"`
	AtEnd    bool              `json:"atEnd,omitempty"`
	When     map[string]string `json:"when,omitempty"`
	Options  []newOptionJSON   `json:"options,omitempty"`
}

// newOptionJSON is one answer to a question; the first of a list is the default.
type newOptionJSON struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Hint  string `json:"hint,omitempty"`
}

// NewQuestionsJSON renders the questions of smith new as indented JSON: for
// each, its key, prompt, default, flag, the answers it offers (label, value,
// hint) and the answers that make it apply. A question with no answers is free
// text and an empty answer takes its default.
func NewQuestionsJSON() ([]byte, error) {
	questions := NewQuestions()
	out := make([]newQuestionJSON, 0, len(questions))
	for _, q := range questions {
		item := newQuestionJSON{Key: q.Key, Prompt: q.Prompt, Default: q.Default, Flag: q.Flag, Many: q.Many, Advanced: q.Advanced, AtEnd: q.AtEnd, When: q.When}
		for _, o := range q.Options {
			item.Options = append(item.Options, newOptionJSON{Value: o.Value, Label: o.Label, Hint: o.Hint})
		}
		out = append(out, item)
	}
	return json.MarshalIndent(out, "", "  ")
}

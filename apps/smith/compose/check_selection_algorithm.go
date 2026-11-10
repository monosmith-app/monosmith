package compose

import (
	"errors"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
)

// Option is one answer to a Question.
type Option struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	// Suggested marks the answer a screen pre-selects. Nothing is chosen
	// unless the person confirms it (ADR 0072).
	Suggested bool `json:"suggested,omitempty"`
}

// Question is a decision the picks leave open: a pick needs an answer on an
// axis and no default fills it.
type Question struct {
	// Choice is the pick that asks.
	Choice string `json:"choice"`
	// About says what that pick writes, in plain words.
	About string `json:"about,omitempty"`
	// Axis is what is asked: "trigger", "runtime".
	Axis    string   `json:"axis"`
	Options []Option `json:"options"`
}

// Check is what the picks add up to.
type Check struct {
	// Language is the language of the picks; empty when no pick fixes one.
	Language string `json:"language,omitempty"`
	// Picks are the names asked for.
	Picks []string `json:"picks"`
	// Complete is true when the picks make a project: nothing conflicts,
	// nothing is open and one choice writes the project.
	Complete bool `json:"complete"`
	// Conflict is set when the picks cannot go together. It names the items
	// and the fix, in one sentence.
	Conflict string `json:"conflict,omitempty"`
	// Missing is set when the picks are right so far but not finished: the next
	// thing to add. It is not an error.
	Missing string `json:"missing,omitempty"`
	// Question is set when a pick needs an answer that smith does not choose.
	Question *Question `json:"question,omitempty"`
	// Selection is the picks with the defaults added, in the order they apply;
	// set when Complete.
	Selection []string `json:"selection,omitempty"`
	// Added are the names the solver added because an axis was needed and had a
	// default; set when Complete.
	Added []string `json:"added,omitempty"`
}

// CheckPicks checks the picks against the offers with the solver of smith add,
// so the composer and `smith add --with` always agree.
//
// Parameters: offers are all the choices; language is the language asked for
// (--lang), or empty to take the one the picks share; picks are choice names.
// Returns: the check. A conflict, a missing piece and an open question are
// fields of the result, not errors.
func CheckPicks(offers []choiceselect.Offer, language string, picks []string) Check {
	out := Check{Picks: picks, Language: language}
	if out.Language == "" {
		out.Language = sharedLanguage(offers, picks)
	}
	if len(picks) == 0 {
		out.Missing = "pick at least the choice that writes the project, for example react"
		return out
	}
	sel, err := choiceselect.SolveIn(offers, out.Language, picks)
	var needs *choiceselect.NeedsAxisError
	var incomplete *choiceselect.IncompleteError
	switch {
	case err == nil:
		out.Complete = true
		out.Selection = sel.Names()
		out.Added = sel.Added
		if len(sel.Choices) > 0 {
			out.Language = sel.Choices[0].Language
		}
	case errors.As(err, &needs):
		out.Question = questionOf(needs)
	case errors.As(err, &incomplete):
		out.Missing = incomplete.Message
	default:
		out.Conflict = err.Error()
	}
	return out
}

// sharedLanguage is the one language all the named choices are offered for;
// empty when there are none, or they belong to several.
func sharedLanguage(offers []choiceselect.Offer, picks []string) string {
	var languages []string
	for _, name := range picks {
		for _, o := range offers {
			if o.Name == name && !slices.Contains(languages, o.Language) {
				languages = append(languages, o.Language)
			}
		}
	}
	if len(languages) == 1 {
		return languages[0]
	}
	return ""
}

// questionOf turns the solver's open axis into a question with its options.
func questionOf(e *choiceselect.NeedsAxisError) *Question {
	q := &Question{Choice: e.Choice, About: e.About, Axis: e.Axis}
	for _, o := range e.Options {
		q.Options = append(q.Options, Option{Name: o.Name, Title: titleOr(o.Title, o.Name), Description: o.Description, Suggested: o.Suggested})
	}
	return q
}

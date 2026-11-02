package choiceselect

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// OfLanguage keeps the offers of one language. A choice name is unique per
// language (ADR 0072), so a caller that knows the language of a project filters
// first and then reads names as before.
//
// Parameters: offers are all the choices; language is the language plugin name.
// Returns: the offers of that language, in the order given.
func OfLanguage(offers []Offer, language string) []Offer {
	var out []Offer
	for _, o := range offers {
		if o.Language == language {
			out = append(out, o)
		}
	}
	return out
}

// scopeToLanguage narrows the offers to the language of the selection. With a
// language, only its offers stay. With none, the offers stay as they are, but a
// name that several languages offer is refused: the person must say which one.
//
// It returns an error that names the languages and the flag to add.
func scopeToLanguage(offers []Offer, language string, names []string) ([]Offer, error) {
	if language != "" {
		return OfLanguage(offers, language), nil
	}
	for _, name := range names {
		name = strings.TrimSpace(name)
		var languages []string
		for _, o := range offers {
			if o.Name == name && !slices.Contains(languages, o.Language) {
				languages = append(languages, o.Language)
			}
		}
		if len(languages) > 1 {
			sort.Strings(languages)
			return nil, fmt.Errorf("%s is offered for %s: add --lang with one of them", name, strings.Join(languages, ", "))
		}
	}
	return offers, nil
}

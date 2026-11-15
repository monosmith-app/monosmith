package choiceselect

import "fmt"

// ProblemKey is the key of a choice in the map that Problems returns: its
// language and name, because a name is unique per language only (ADR 0072).
func ProblemKey(o Offer) string { return o.Language + "/" + o.Name }

// Problems says, for each choice that cannot be used with the loaded plugins,
// why: its language plugin is not loaded, a choice it requires is offered by
// no plugin, or an axis it needs has no choice to fill it. A usable choice has
// no entry. languageLoaded says whether a language plugin is enabled. The
// keys are ProblemKey of the offer.
func Problems(offers []Offer, languageLoaded func(language string) bool) map[string]string {
	out := map[string]string{}
	for _, o := range offers {
		key := ProblemKey(o)
		if !languageLoaded(o.Language) {
			out[key] = fmt.Sprintf("the %s plugin is not enabled: enable it, or ignore the choice", o.Language)
			continue
		}
		same := OfLanguage(offers, o.Language)
		index, err := Index(same)
		if err != nil {
			index = map[string]Offer{}
		}
		for _, need := range o.Requires {
			if _, ok := index[need]; !ok {
				out[key] = fmt.Sprintf("it requires %s, which no enabled plugin offers", need)
				break
			}
		}
		if _, bad := out[key]; bad {
			continue
		}
		for _, axis := range o.Needs {
			if len(onAxis(same, axis, o.Language)) == 0 {
				out[key] = fmt.Sprintf("it needs a choice on the %s axis, and no enabled plugin offers one", axis)
				break
			}
		}
	}
	return out
}

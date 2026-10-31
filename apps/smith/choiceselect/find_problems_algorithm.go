package choiceselect

import "fmt"

// Problems says, for each choice that cannot be used with the loaded plugins,
// why: its language plugin is not loaded, a choice it requires is offered by
// no plugin, or an axis it needs has no choice to fill it. A usable choice has
// no entry. languageLoaded says whether a language plugin is enabled.
func Problems(offers []Offer, languageLoaded func(language string) bool) map[string]string {
	index, err := Index(offers)
	if err != nil {
		index = map[string]Offer{}
	}
	out := map[string]string{}
	for _, o := range offers {
		if !languageLoaded(o.Language) {
			out[o.Name] = fmt.Sprintf("the %s plugin is not enabled: enable it, or ignore the choice", o.Language)
			continue
		}
		for _, need := range o.Requires {
			if _, ok := index[need]; !ok {
				out[o.Name] = fmt.Sprintf("it requires %s, which no enabled plugin offers", need)
				break
			}
		}
		if _, bad := out[o.Name]; bad {
			continue
		}
		for _, axis := range o.Needs {
			if len(onAxis(offers, axis, o.Language)) == 0 {
				out[o.Name] = fmt.Sprintf("it needs a choice on the %s axis, and no enabled plugin offers one", axis)
				break
			}
		}
	}
	return out
}

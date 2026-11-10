package compose

import "github.com/monosmith-app/monosmith/apps/smith/choiceselect"

// CheckItems is CheckPicks for the items a screen holds. It first compares the
// languages of the items, because two items may share a name and differ in
// language (azure-functions exists for node, python and dotnet).
//
// Parameters: offers are all the choices; language is the language item the
// person ticked, or empty; picks are the choice items ticked, in order.
// Returns: the check; a mix of languages is a conflict that names both items.
func CheckItems(offers []choiceselect.Offer, language string, picks []Item) Check {
	names := make([]string, len(picks))
	for i, p := range picks {
		names[i] = p.Name
	}
	if language == "" && len(picks) > 0 {
		language = picks[0].Language
	}
	for _, p := range picks {
		if p.Language != "" && language != "" && p.Language != language {
			return Check{Picks: names, Language: language, Conflict: p.Name + " is for " + p.Language + " projects and " + picks[0].Name + " for " + language + ": keep items of one language, or remove one of them"}
		}
	}
	return CheckPicks(offers, language, names)
}

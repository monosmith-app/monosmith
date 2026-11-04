package guide

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// projectKinds are the kinds the choices step is made for, one list each. A
// step with no kind lists every choice: it serves a form that never asked the kind.
var projectKinds = []string{"app", "lib", "package"}

// kindOfAnswers is the kind a choices step is for: the kind answered, or "" when
// the form holds none it knows.
func kindOfAnswers(a map[string][]string) string {
	if k := first(a, "kind"); slices.Contains(projectKinds, k) {
		return k
	}
	return ""
}

// fitsKind reports whether the choice can be used for a project of the kind; a
// choice that names no kinds fits all, and so does an empty kind.
func fitsKind(o ChoiceOffer, kind string) bool {
	return kind == "" || len(o.Kinds) == 0 || slices.Contains(o.Kinds, kind)
}

// fitting keeps the offers that fit the kind.
func fitting(offers []ChoiceOffer, kind string) []ChoiceOffer {
	var out []ChoiceOffer
	for _, o := range offers {
		if fitsKind(o, kind) {
			out = append(out, o)
		}
	}
	return out
}

// solverOffers are the offers as the solver of smith add reads them, so the form
// and `smith add --with` agree.
func solverOffers(offers []ChoiceOffer) []choiceselect.Offer {
	out := make([]choiceselect.Offer, len(offers))
	for i, o := range offers {
		var roles []protocol.Role
		for _, r := range o.Roles {
			roles = append(roles, protocol.Role(r))
		}
		out[i] = choiceselect.Offer{Plugin: o.Plugin, Choice: protocol.Choice{
			Axis: o.Axis, Name: o.Name, Title: o.Title, Description: o.Description, Aliases: o.Aliases, Layer: o.Layer,
			Language: o.Language, Roles: roles, Needs: o.Needs, Default: o.Default, Suggested: o.Suggested, Many: o.Many,
			Requires: o.Requires, Conflicts: o.Conflicts,
		}}
	}
	return out
}

// layerNeeds maps an outcome that needs a choice on a layer (a website needs a
// framework, a function app a cloud host, a desktop app a shell) to that layer.
func layerNeeds(outcomes []compose.Outcome) map[string]string {
	out := map[string]string{}
	for _, o := range outcomes {
		if o.Layer != "" {
			out[o.ID] = o.Layer
		}
	}
	return out
}

// layerWords say what a layer holds, for the sentence that asks for one.
var layerWords = map[string]string{"framework": "a framework", "shell": "a desktop shell", "host": "a cloud host"}

// picksCheck is the check of the choices step: it runs the solver of smith add
// on the ticked set, offline, and says in plain words what is wrong or missing,
// so the form does not fail at its end. Nothing ticked is fine: the stack's own
// defaults are used.
//
// Parameters: lang is the language of the step; offers are its offers; needs is
// layerNeeds. Returns: the check a Step runs on enter, with the ticked names
// joined by commas; it returns "" when the set is fine or still open to a question.
func picksCheck(lang string, offers []ChoiceOffer, needs map[string]string) func(map[string][]string, string) string {
	solver := solverOffers(offers)
	// A list that says no roles is one the solver cannot judge: smith add checks it.
	knowsRoles := slices.ContainsFunc(offers, func(o ChoiceOffer) bool { return len(o.Roles) > 0 })
	return func(a map[string][]string, value string) string {
		if strings.TrimSpace(value) == "" {
			return ""
		}
		picks := strings.Split(value, ",")
		kind := kindOfAnswers(a)
		for _, name := range picks {
			if i := slices.IndexFunc(offers, func(o ChoiceOffer) bool { return o.Name == name }); i >= 0 && !fitsKind(offers[i], kind) {
				var kinds []string
				for _, k := range offers[i].Kinds {
					kinds = append(kinds, k+"s")
				}
				return fmt.Sprintf("%s makes only %s, and this is a %s: untick it, or go back and pick another kind", name, strings.Join(kinds, " and "), kind)
			}
		}
		if !knowsRoles {
			return ""
		}
		check := compose.CheckPicks(solver, lang, picks)
		switch {
		case check.Conflict != "":
			return check.Conflict
		case check.Missing != "":
			return plainWords(check.Missing)
		case check.Question != nil:
			return ""
		}
		layer := needs[first(a, outcomeKey)]
		if layer == "" {
			return ""
		}
		var have []string
		for _, o := range offers {
			if compose.LayerOf(o.Layer, o.Axis) == layer && fitsKind(o, kind) {
				have = append(have, o.Name)
				if slices.Contains(picks, o.Name) {
					return ""
				}
			}
		}
		return fmt.Sprintf("this kind of project needs %s: tick one of %s, or untick everything", layerWords[layer], strings.Join(have, ", "))
	}
}

// plainWords turns a sentence of smith add (which talks of --with) into one for
// the form, where nothing is typed as a flag.
func plainWords(s string) string {
	return strings.NewReplacer(" to --with", "", "give no --with to start", "untick everything to start").Replace(s)
}

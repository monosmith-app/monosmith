package choiceselect

import (
	"errors"
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// FitKinds says, for each choice, the project kinds it can be used for. A form
// lists only the choices that fit the kind the person picked first, so a package
// is not offered a framework that cannot write one.
//
// A choice that applies on the stack's own project (AppliesOnBase) fits every
// kind and has no entry. A choice that writes the project fits the kinds its
// plugin scaffolds. Any other choice (a style, a library, a desktop shell) fits
// the kinds of the choices that it can join: those that write the project and
// solve with it.
//
// Parameters: offers are all the choices; scaffoldKinds gives the kinds the
// named plugin scaffolds.
// Returns: the kinds by ProblemKey of the offer, in the order app, lib, package.
func FitKinds(offers []Offer, scaffoldKinds func(plugin string) []protocol.ProjectKind) map[string][]protocol.ProjectKind {
	out := map[string][]protocol.ProjectKind{}
	for _, o := range offers {
		if AppliesOnBase(o) {
			continue
		}
		var kinds []protocol.ProjectKind
		add := func(more []protocol.ProjectKind) {
			for _, k := range more {
				if !slices.Contains(kinds, k) {
					kinds = append(kinds, k)
				}
			}
		}
		if o.Plays(protocol.RoleScaffold) {
			add(scaffoldKinds(o.Plugin))
		} else {
			for _, host := range offers {
				if host.Language != o.Language || !host.Plays(protocol.RoleScaffold) || host.Name == o.Name {
					continue
				}
				var need *NeedsAxisError
				if _, err := solve(OfLanguage(offers, o.Language), []string{host.Name, o.Name}, false); err == nil || errors.As(err, &need) {
					add(scaffoldKinds(host.Plugin))
				}
			}
		}
		slices.SortFunc(kinds, func(a, b protocol.ProjectKind) int {
			order := []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage}
			return slices.Index(order, a) - slices.Index(order, b)
		})
		out[ProblemKey(o)] = kinds
	}
	return out
}

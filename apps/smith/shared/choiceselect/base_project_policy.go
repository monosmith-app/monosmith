package choiceselect

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// shapingLayers are the layers whose choices change what the project is: a
// framework, a desktop shell or a cloud host. A choice on one of them is not
// added to a project the stack wrote by itself.
var shapingLayers = []string{"framework", "shell", "host"}

// AppliesOnBase reports whether the choice can be added to the project that the
// stack's own scaffold writes, with no choice that writes the project: a lint
// tool or a test runner. Such a choice only joins the toolchain, needs nothing
// else, and does not shape the project.
//
// Parameters: o is the choice.
// Returns: true when a selection of such choices alone is complete.
func AppliesOnBase(o Offer) bool {
	layer := o.Layer
	if layer == "" {
		layer = o.Axis
	}
	return o.Plays(protocol.RoleToolchain) && !o.Plays(protocol.RoleScaffold) &&
		len(o.Needs) == 0 && len(o.Requires) == 0 && !slices.Contains(shapingLayers, layer)
}

// onBase reports whether every chosen choice applies on the stack's own project.
func onBase(chosen []Offer) bool {
	return len(chosen) > 0 && !slices.ContainsFunc(chosen, func(o Offer) bool { return !AppliesOnBase(o) })
}

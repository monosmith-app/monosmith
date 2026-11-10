package swapchoice

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ChoicesOf are the choices a stack stands for. A stack made with smith add
// --with lists them. A stack that a plugin publishes by name (node-jest) lists
// none, so they are read from its plugins: a plugin that offers exactly one
// choice (jest, vitest, eslint) contributes it. A plugin that offers many (the
// Vite plugin's frameworks and styles) says nothing about which one the stack
// has, so it contributes none.
//
// Parameters: offers are the choices the loaded plugins offer; st is the stack.
// Returns: the choice names, in the order of the stack's plugins.
func ChoicesOf(offers []choiceselect.Offer, st protocol.Stack) []string {
	if len(st.Choices) > 0 {
		return st.Choices
	}
	var out []string
	for _, plugin := range st.Plugins() {
		var own []string
		for _, o := range offers {
			if o.Plugin == plugin {
				own = append(own, o.Name)
			}
		}
		if len(own) == 1 && !slices.Contains(out, own[0]) {
			out = append(out, own[0])
		}
	}
	return out
}

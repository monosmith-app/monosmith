package choiceselect

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Offer is a choice with the plugin that owns it.
type Offer struct {
	protocol.Choice
	// Plugin is the name of the plugin that published the choice.
	Plugin string `json:"plugin"`
}

// Selection is a checked set of choices, ready to become a stack.
type Selection struct {
	// Choices are in the order they apply: the choice that writes the project
	// first, and a choice after those it requires.
	Choices []Offer
	// Added are the names of the choices the solver added because an axis was
	// needed and had a default.
	Added []string
}

// Names are the names of the choices, in order.
func (s Selection) Names() []string {
	out := make([]string, len(s.Choices))
	for i, c := range s.Choices {
		out[i] = c.Name
	}
	return out
}

// Stack is the selection as a stack called name: the language of the
// choices, the plugin that writes the project as the scaffold, and the
// plugins of the choices that play the toolchain role, in order.
func (s Selection) Stack(name, description string) protocol.Stack {
	st := protocol.Stack{Name: name, Description: description, Choices: s.Names()}
	for _, c := range s.Choices {
		st.Language = c.Language
		if c.Plays(protocol.RoleScaffold) {
			st.Scaffold = c.Plugin
		}
		if c.Plays(protocol.RoleToolchain) && !slices.Contains(st.Toolchain, c.Plugin) {
			st.Toolchain = append(st.Toolchain, c.Plugin)
		}
	}
	return st
}

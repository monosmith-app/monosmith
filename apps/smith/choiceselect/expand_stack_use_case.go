package choiceselect

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ExpandStack turns a stack that is a saved selection into a full stack. The
// choices are solved against the offers: the stack gets the language, the
// scaffold and the toolchain they imply, the default choices the selection
// left out, and then the stack's own extra toolchain, structure and deploy
// plugins. A stack with no choices is returned as it is.
//
// It returns an error when the selection is not valid (the message of Solve),
// or when the stack names a language or a scaffold that its choices do not
// imply.
func ExpandStack(st protocol.Stack, offers []Offer) (protocol.Stack, error) {
	if len(st.Choices) == 0 {
		return st, nil
	}
	sel, err := Solve(offers, st.Choices)
	if err != nil && st.Language != "" {
		// A name that several languages offer (ADR 0072) is read in the language
		// the stack names.
		if scoped, scopedErr := SolveIn(offers, st.Language, st.Choices); scopedErr == nil {
			sel, err = scoped, nil
		}
	}
	if err != nil {
		return st, err
	}
	derived := sel.Stack(st.Name, st.Description)
	if st.Language != "" && st.Language != derived.Language {
		return st, fmt.Errorf("its choices are for %s projects, but the stack names the language %s", derived.Language, st.Language)
	}
	if st.Scaffold != "" && st.Scaffold != derived.Scaffold {
		return st, fmt.Errorf("its choices are written by %s, but the stack names the scaffold %s", derived.Scaffold, st.Scaffold)
	}
	for _, extra := range st.Toolchain {
		if !slices.Contains(derived.Toolchain, extra) {
			derived.Toolchain = append(derived.Toolchain, extra)
		}
	}
	derived.Structure, derived.Deploy = st.Structure, st.Deploy
	return derived, nil
}

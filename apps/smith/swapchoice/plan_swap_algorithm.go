package swapchoice

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
)

// Swap is a change of one choice of a project, checked by the choice solver.
type Swap struct {
	// Axis is the question the choice answers ("test", "lint", "style").
	Axis string
	// From is the choice the project has on the axis; To replaces it.
	From, To string
	// Choices are the choices of the project after the swap, in the order
	// they apply.
	Choices []string
	// Added and Removed are the choices that join and leave the project:
	// the swapped one, and any default the new choice needs or a choice
	// that only fitted the old one.
	Added, Removed []string
}

// PlanSwap checks that a project can have another choice on one axis and says
// what changes. It reads and writes nothing.
//
// Parameters: offers are the choices the loaded plugins offer; current are the
// choices of the project's stack; axis is the axis to change; to is the choice
// that replaces the project's choice on the axis.
// Returns: the swap, or an error in one sentence that says what is wrong and
// what to do: the project has no choice on the axis, the new choice is unknown,
// is on another axis, is the choice the project has, or does not fit the rest.
func PlanSwap(offers []choiceselect.Offer, current []string, axis, to string) (Swap, error) {
	index, err := choiceselect.Index(offers)
	if err != nil {
		return Swap{}, err
	}
	var from string
	var kept []string
	for _, name := range current {
		if o, ok := index[name]; ok && o.Axis == axis {
			from = name
			continue
		}
		kept = append(kept, name)
	}
	if from == "" {
		return Swap{}, fmt.Errorf("the project has no choice on the %s axis, so there is nothing to swap. Its choices are: %s", axis, strings.Join(current, ", "))
	}
	target, ok := index[to]
	if !ok {
		return Swap{}, fmt.Errorf("no choice is named %q on the %s axis. Choices there: %s (smith choices lists them)", to, axis, strings.Join(onAxis(offers, axis), ", "))
	}
	if target.Axis != axis {
		return Swap{}, fmt.Errorf("%s is a choice on the %s axis, not on the %s axis. Choices on %s: %s", to, target.Axis, axis, axis, strings.Join(onAxis(offers, axis), ", "))
	}
	if to == from {
		return Swap{}, fmt.Errorf("the project already uses %s on the %s axis", to, axis)
	}
	sel, err := choiceselect.SolveForExisting(offers, append(kept, to))
	if err != nil {
		return Swap{}, fmt.Errorf("%s does not fit the rest of the project: %w", to, err)
	}
	s := Swap{Axis: axis, From: from, To: to, Choices: sel.Names()}
	for _, name := range s.Choices {
		if !slices.Contains(current, name) {
			s.Added = append(s.Added, name)
		}
	}
	for _, name := range current {
		if !slices.Contains(s.Choices, name) {
			s.Removed = append(s.Removed, name)
		}
	}
	return s, nil
}

// onAxis lists the names of the choices on an axis, in the order offered.
func onAxis(offers []choiceselect.Offer, axis string) []string {
	var out []string
	for _, o := range offers {
		if o.Axis == axis {
			out = append(out, o.Name)
		}
	}
	return out
}

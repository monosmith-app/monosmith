package choiceselect

import (
	"fmt"
	"strings"
)

// NeedsAxisError says a choice needs an answer on an axis and no choice of the
// axis is the default, so smith must ask: it never picks for the person
// (ADR 0072).
type NeedsAxisError struct {
	// Choice is the choice that needs the axis.
	Choice string
	// About is the description of that choice: what it writes.
	About string
	// Axis is the axis that has no choice yet.
	Axis string
	// Options are the choices of the axis that fit the selection, by name.
	Options []Offer
}

// Suggested is the option an asker pre-selects; false when none is marked.
func (e *NeedsAxisError) Suggested() (Offer, bool) {
	for _, o := range e.Options {
		if o.Suggested {
			return o, true
		}
	}
	return Offer{}, false
}

// Error names the choice, the axis, the options and the suggested one, and
// tells the person to add the answer to --with.
func (e *NeedsAxisError) Error() string {
	names := make([]string, len(e.Options))
	for i, o := range e.Options {
		names[i] = o.Name
		if o.Suggested {
			names[i] += " (suggested)"
		}
	}
	return fmt.Sprintf("%s needs a choice on the %s axis, and none is the default: add one of %s to --with", e.Choice, e.Axis, strings.Join(names, ", "))
}

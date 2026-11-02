package composer

import "strings"

// PlanRequest asks for the plan of a complete selection. The model cannot read
// the files an add writes, so its host answers with WithPlan.
type PlanRequest struct {
	// Kind is app, lib or package; Name is the project name.
	Kind, Name string
	// Language is the language of the selection; Picks are the names ticked.
	Language string
	Picks    []string
}

// Argv is the smith command line that reads the plan (without the leading
// "smith"): the same text the screens of other media ask for.
func (r PlanRequest) Argv() []string {
	return []string{"compose", "--json", "--kind", r.Kind, "--name", r.Name, "--lang", r.Language, "--with", strings.Join(r.Picks, ",")}
}

// Outcome is what a key did beyond changing the screen.
type Outcome struct {
	// Done is true when the screen should close.
	Done bool
	// Plan asks the host for the plan of the selection.
	Plan *PlanRequest
	// Run is the smith command line to run now (without the leading "smith").
	Run []string
	// Copy is text for the clipboard.
	Copy string
	// Note is a short message for the host to show.
	Note string
}

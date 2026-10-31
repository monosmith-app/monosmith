package explain

import (
	"fmt"
	"io"
)

// printFacts writes facts for a person: one line per fact with its value,
// then the scope and the source it came from.
func printFacts(w io.Writer, project string, facts []FactLine) {
	if len(facts) == 0 {
		fmt.Fprintf(w, "%s has no shared facts: no plugin published one and smith.toml sets none.\n", project)
		return
	}
	width := 0
	for _, f := range facts {
		width = max(width, len(f.Key))
	}
	noun := "shared facts"
	if len(facts) == 1 {
		noun = "shared fact"
	}
	fmt.Fprintf(w, "%s has %d %s:\n", project, len(facts), noun)
	for _, f := range facts {
		fmt.Fprintf(w, "  %-*s  %s\n", width, f.Key, valueText(f.Value))
		fmt.Fprintf(w, "  %-*s  (%s, from %s)\n", width, "", f.Scope, f.Source)
	}
}

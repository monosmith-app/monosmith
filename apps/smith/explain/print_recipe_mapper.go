package explain

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/recipedrift"
)

// printRecipe lists what a recipe decided for a project, each value with the
// recipe file and line it came from.
func printRecipe(w io.Writer, e recipedrift.Explained) {
	version := ""
	if e.Version != "" {
		version = " " + e.Version
	}
	fmt.Fprintf(w, "recipe %s%s — %s\n", e.Recipe, version, e.Source)
	width := 0
	for _, v := range e.Values {
		width = max(width, len(v.Name))
	}
	for _, v := range e.Values {
		fmt.Fprintf(w, "  %-*s  %s\n", width, v.Name, v.Value)
		if v.Origin != nil {
			fmt.Fprintf(w, "  %-*s    from %s\n", width, "", v.Origin)
		}
	}
	fmt.Fprintln(w)
}

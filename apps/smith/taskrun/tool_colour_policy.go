package taskrun

import (
	"io"
	"os"

	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// ToolColour decides whether the task output written to out reaches a person
// who sees colour (termcolour.ToolWantedFor, with this process's environment).
//
// Parameters: out is where the run's text goes.
// Returns: the value for Request.Colour.
func ToolColour(out io.Writer) bool {
	return termcolour.ToolWantedFor(os.Getenv, out, term.IsTerminal)
}

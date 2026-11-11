package taskrun

import (
	"io"
	"os"

	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// ToolColour decides whether the task output written to out reaches a person
// who sees colour: termcolour.ToolWanted, with out as the terminal to check.
// A writer that is not a file (a buffer, a pipe smith reads) is no terminal,
// so only a CI that renders colour, FORCE_COLOR or SMITH_TOOL_COLOR turn it on.
//
// Parameters: out is where the run's text goes.
// Returns: the value for Request.Colour.
func ToolColour(out io.Writer) bool {
	f, ok := out.(*os.File)
	return termcolour.ToolWanted(os.Getenv, ok && term.IsTerminal(int(f.Fd())))
}

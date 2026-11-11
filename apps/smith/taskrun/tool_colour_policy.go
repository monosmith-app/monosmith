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

// runColour is the colour decision of `smith run`. --json is read by a
// program, so never. --json-events keeps every line plain in "line" and adds
// "styled" only when the host handed tool colour on explicitly (the TUI shows
// it). Text output follows ToolColour.
//
// Parameters: getenv reads the environment; asJSON and jsonEvents are the
// output flags; out is where the text goes; isTerminal checks a descriptor.
// Returns: the value for Request.Colour.
func runColour(getenv func(string) string, asJSON, jsonEvents bool, out io.Writer, isTerminal func(int) bool) bool {
	switch {
	case asJSON:
		return false
	case jsonEvents:
		return getenv(termcolour.ToolColourEnv) == "1" && termcolour.ToolWanted(getenv, false)
	}
	return termcolour.ToolWantedFor(getenv, out, isTerminal)
}

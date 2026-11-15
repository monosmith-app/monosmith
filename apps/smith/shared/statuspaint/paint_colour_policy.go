package statuspaint

import (
	"io"
	"os"

	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// Wanted is the colour decision for a command's own lines: the same
// termcolour.Wanted every table of smith follows (a terminal, or a CI that
// renders colour; never NO_COLOR or TERM=dumb). It ignores the tool colour a host
// hands on, because a host that reads the lines (the interface) must get them plain.
//
// Parameters: getenv reads the environment; out is where the text goes;
// isTerminal checks a descriptor.
// Returns: true when the lines may be painted.
func Wanted(getenv func(string) string, out io.Writer, isTerminal func(int) bool) bool {
	f, ok := out.(*os.File)
	return termcolour.Wanted(getenv, ok && f != nil && isTerminal(int(f.Fd())))
}

// WantedHere is Wanted for this process: its environment and term.IsTerminal.
func WantedHere(out io.Writer) bool { return Wanted(os.Getenv, out, term.IsTerminal) }

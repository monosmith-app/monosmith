package lintreport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/taskrun"
)

// selfPath is the smith that is running.
var selfPath = os.Executable

// Check runs `smith run lint -p <name>` with the smith that is running,
// without --fix, and says what failed. Its output is not shown: the items
// name the counts, the rules and the command that shows the rest.
//
// Parameters: root is the workspace root; name is the project; out gets one
// progress line.
// Returns: the items left (Summarize), or nil when the lint passes or smith
// cannot find its own executable.
func Check(root, name string, out io.Writer) []Item {
	self, err := selfPath()
	if err != nil {
		return nil
	}
	fmt.Fprintf(out, "… smith run lint -p %s\n", name)
	var captured bytes.Buffer
	code, err := taskrun.RunProcess(context.Background(), taskrun.Spec{Argv: []string{self, "run", "lint", "-p", name}, Dir: root, Env: taskrun.Environment(nil), Out: &captured})
	if err == nil && code == 0 {
		lines := strings.Split(strings.TrimRight(captured.String(), "\r\n"), "\n")
		fmt.Fprintf(out, "  %s\n", strings.TrimSpace(lines[len(lines)-1]))
		return nil
	}
	return Summarize(name, captured.String())
}

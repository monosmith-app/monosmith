package adoptmove

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// AskWorkspace asks where one development dependency goes and reads the answer
// from in: 1 for the workspace manifest, 2 for the project's (an empty line
// takes the project, which changes the least). It gives up after three bad
// answers. Nothing is changed.
//
// Parameters: name is the package; in and out are the terminal.
// Returns: true when the person chose the workspace.
func AskWorkspace(name string, in io.Reader, out io.Writer) (bool, error) {
	r := bufio.NewReader(in)
	fmt.Fprintf(out, "\n%s is a development dependency, and smith does not know it as a tool. Where should it go?\n"+
		"  1) the workspace: the root package.json, for every project\n"+
		"  2) the project: it stays in the project's package.json (default)\n", name)
	for try := 0; try < maxTries; try++ {
		fmt.Fprint(out, "Choose 1-2 [2]: ")
		line, err := r.ReadString('\n')
		answer := strings.TrimSpace(strings.TrimPrefix(line, string(rune(0xFEFF))))
		if err != nil && answer == "" {
			return false, problem.New("no answer was given", "--hoist "+name+" sends it to the workspace without a question")
		}
		switch answer {
		case "1":
			return true, nil
		case "", "2":
			return false, nil
		}
		fmt.Fprintf(out, "  %q is not a choice; type 1 or 2.\n", answer)
	}
	return false, problem.New("no valid answer after three tries; nothing was changed", "--hoist "+name+" sends it to the workspace without a question")
}

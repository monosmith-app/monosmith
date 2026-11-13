package convert

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

// AskOutputChoices asks, for each build task whose declared outputs differ
// from what its build writes, which outputs to cache: the detected ones
// (suggested, what the plan holds) or the ones the old tool declared. An
// empty answer takes the suggestion, as --yes does.
//
// Parameters: in reads the answers; out shows the questions; plan is the
// conversion plan, changed in place when the declared outputs are kept;
// given holds the "<project>:<task>" keys --outputs answered, not asked again.
// Returns: an error when an answer is not one of the options.
func AskOutputChoices(in *bufio.Reader, out io.Writer, plan *convertplan.Plan, given map[string]bool) error {
	for i, c := range plan.OutputChoices {
		if given[c.Project+":"+c.Task] {
			continue
		}
		fmt.Fprintf(out, "\n%s:%s writes %s, but %s declared %s as its outputs. Cache:\n", c.Project, c.Task, strings.Join(c.Detected, ", "), plan.ToolName(), strings.Join(c.Declared, ", "))
		fmt.Fprintf(out, "  1) %s, what the build writes (suggested)\n", strings.Join(c.Detected, ", "))
		fmt.Fprintf(out, "  2) %s, as %s declared them\n", strings.Join(c.Declared, ", "), plan.ToolName())
		fmt.Fprint(out, "Choose [1]: ")
		line, _ := in.ReadString('\n')
		switch strings.TrimSpace(line) {
		case "", "1":
		case "2":
			plan.KeepDeclaredOutputs(i)
		default:
			return fmt.Errorf("%q is not one of the options 1 to 2 for the outputs of %s:%s", strings.TrimSpace(line), c.Project, c.Task)
		}
	}
	return nil
}

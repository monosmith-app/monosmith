package doctorfix

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/smithfix"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// planTimeout bounds the plan a step prints for the review.
const planTimeout = time.Minute

// selfPath is the smith that is running; tests replace it.
var selfPath = os.Executable

// readPlan runs a plan command line and returns what it printed; tests
// replace it so no test starts a process.
var readPlan = func(ctx context.Context, dir string, argv []string) (string, error) {
	var out bytes.Buffer
	code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: argv, Dir: dir, Env: taskrun.Environment(nil), Out: &out})
	if err == nil && code != 0 {
		err = fmt.Errorf("exited with code %d", code)
	}
	return out.String(), err
}

// resolveProgram makes a step that runs smith run the smith that is running,
// not the one the PATH finds; other steps are returned as they are.
func resolveProgram(argv []string) []string {
	if !smithfix.IsStep(argv) {
		return argv
	}
	self, err := selfPath()
	if err != nil {
		return argv
	}
	return append([]string{self}, argv[1:]...)
}

// writePlan prints, under a step that runs smith, what the command would do
// (its plan flag), so the review shows the change and not only the command. It
// writes nothing for a step without a plan. root is the folder the plan runs in:
// a copy that already holds the fix's files, or the workspace when the copy could
// not be made (copyErr says why, and the plan then shows the files as they are now).
func writePlan(w io.Writer, root string, copyErr error, step protocol.FixStep) {
	plan, ok := smithfix.PlanArgv(step.Argv)
	if !ok {
		return
	}
	dir, err := inside(root, orRoot(step.Dir))
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), planTimeout)
	defer cancel()
	text, err := readPlan(ctx, dir, resolveProgram(plan))
	fmt.Fprintf(w, "          what it would do ($ %s):\n", displayCommand(plan))
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if line != "" {
			fmt.Fprintf(w, "            %s\n", line)
		}
	}
	if err != nil {
		fmt.Fprintf(w, "            (smith could not print the plan: %v)\n", err)
	}
	if copyErr != nil {
		fmt.Fprintf(w, "            (this plan does not include the files above: %v)\n", copyErr)
	}
}

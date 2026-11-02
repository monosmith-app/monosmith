package operationplan

import "io"

// Gate shows what a command will touch and decides whether it goes on: the
// plan is printed, a terminal is asked, --yes passes, and without a terminal
// and without --yes it refuses and names both flags. A command that writes
// files calls it before its first write. A plan with no line passes unasked.
//
// Parameters: out gets the plan and the question; in gives the answer; title is
// one sentence for the plan; heading names the group of lines; lines are the
// paths the command touches; yes and dryRun are the --yes and --dry-run flags.
// Returns: whether to write, and ErrDeclined or the refusal as the error.
func Gate(out io.Writer, in io.Reader, title, heading string, lines []Line, yes, dryRun bool) (bool, error) {
	plan := Plan{Title: title}
	for _, l := range lines {
		plan.Add(heading, l)
	}
	return Present(out, in, plan, Options{Yes: yes, DryRun: dryRun, Ask: true, Env: DetectEnvironment()})
}

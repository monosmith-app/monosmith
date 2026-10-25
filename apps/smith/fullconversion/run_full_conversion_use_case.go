package fullconversion

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
)

// RunOptions say how the removals are asked and applied.
type RunOptions struct {
	// Root is the converted workspace.
	Root string
	Out  io.Writer
	// In is where answers are read from (one reader for the whole run, so no
	// answer is lost to a read-ahead); used only when Interactive.
	In          *bufio.Reader
	Interactive bool
	// Yes applies every group without asking.
	Yes     bool
	Choices convertchoices.Choices
	// SkipFinish leaves out the install and the checks.
	SkipFinish bool
	// Verify builds, tests, lints and runs doctor after the install.
	Verify bool
	// Smith is the command that runs smith for the checks.
	Smith string
	// Files lists the workspace's files as they are now.
	Files func() ([]string, error)
}

// RunResult is what the run did.
type RunResult struct {
	// Applied and Declined are the groups' IDs.
	Applied  []string
	Declined []string
	// Finished is true when the install and the checks ran and passed.
	Finished bool
}

// Run asks each group of removals in turn and applies the ones agreed. After
// every answer the next group is planned again from the files as they are,
// so a group declined earlier keeps what it would have removed and the
// groups after it see that. It ends with the install and the checks.
func Run(ctx context.Context, o RunOptions) (RunResult, error) {
	var res RunResult
	declined := map[string]bool{}
	var last removalplan.Plan
	for {
		files, err := o.Files()
		if err != nil {
			return res, err
		}
		last = removalplan.Build(legacyinventory.NewTree(o.Root, files), o.Choices, declined)
		if len(last.Groups) == 0 {
			break
		}
		g := last.Groups[0]
		fmt.Fprintln(o.Out)
		walked, err := doctorfix.Walk(ctx, []doctorfix.Planned{g.Fix}, walkOptions(o))
		if err != nil {
			return res, err
		}
		if len(walked.Applied) > 0 {
			res.Applied = append(res.Applied, g.ID)
		} else {
			declined[g.ID] = true
			res.Declined = append(res.Declined, g.ID)
		}
	}
	for _, k := range last.Kept {
		fmt.Fprintf(o.Out, "kept %s: %s\n", k.What, k.Why)
	}
	for _, n := range last.Notes {
		fmt.Fprintf(o.Out, "! %s\n", n)
	}
	if o.SkipFinish {
		fmt.Fprintln(o.Out)
		fmt.Fprintln(o.Out, "Run npm install to refresh package-lock.json for the packages taken out and put in.")
		return res, nil
	}
	fmt.Fprintln(o.Out)
	walked, err := doctorfix.Walk(ctx, []doctorfix.Planned{removalplan.Finish(o.Smith, o.Verify)}, walkOptions(o))
	if err != nil {
		return res, err
	}
	res.Finished = len(walked.Applied) > 0
	return res, nil
}

func walkOptions(o RunOptions) doctorfix.Options {
	opts := doctorfix.Options{
		Root: o.Root, Out: o.Out, Interactive: o.Interactive,
		Yes: o.Yes, Explicit: true,
		Command: func(doctorfix.Planned, []int) string {
			return "smith convert " + convertchoices.Flags(o.Choices) + " --yes"
		},
	}
	if o.In != nil {
		opts.In = o.In
	}
	return opts
}

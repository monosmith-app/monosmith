package projectimport

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/convert"
	"github.com/monosmith-app/monosmith/apps/smith/features/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/features/projecttooling"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
)

// SettleTooling asks the tooling question of a Node project before anything
// changes: keep its own build and lint tools, or use smith's, with smith's
// recommended when a tool is legacy (projecttooling.Assess and Decide, the
// question of smith adopt). --tooling answers it; with --yes the project
// keeps its tools; in a dry run nothing is asked. A project of another
// language keeps its tools and is not asked. With smith's tooling and no
// --eslint-setup, the split puts the project's own ESLint setup aside.
//
// Parameters: src is the fetched source; p the plan, which gets Tooling and
// Legacy; o the options, whose ESLint the answer may settle.
// Returns: an error when the source's package.json is not JSON, a --tooling
// value is unknown, or nobody can answer.
func SettleTooling(src Source, p *Plan, o *Options) error {
	given, err := projecttooling.ParseAnswers(o.Tooling)
	if err != nil {
		return problem.New(err.Error(), "--tooling smith|keep")
	}
	a, err := projecttooling.Assess(src.Dir, ".")
	if err != nil || !a.Node {
		return err
	}
	a.Name = p.Name
	in, _ := o.In.(*bufio.Reader)
	answers, err := projecttooling.Decide(projecttooling.DecideRequest{
		Projects: []projecttooling.Assessment{a}, Given: given,
		Interactive: operationplan.ShouldAsk(o.Env, false), Yes: o.Yes, Preview: o.DryRun,
		In: in, Out: o.Out,
	})
	if err != nil {
		return err
	}
	p.Tooling, p.Legacy = string(answers[p.Name]), a.Legacy
	if p.Tooling == string(projecttooling.Smith) && o.ESLint == "" {
		o.ESLint = ESLintShared
	}
	return nil
}

// ApplyTooling takes the tooling answer for the imported project, once it is
// placed, split and joined (projecttooling.Apply, the step smith adopt takes
// too). It installs the whole workspace once, because smith's tools go into
// the root package.json; the build, the tests and doctor are the import's
// own check.
//
// Parameters: root is the workspace; p the plan; o the options.
// Returns: what the step could not do, as items of the import's list.
func ApplyTooling(root string, p Plan, o Options) []Left {
	if p.Tooling != string(projecttooling.Smith) {
		if p.Tooling == string(projecttooling.Keep) {
			fmt.Fprintf(o.Out, "  tooling: %s keeps its own tools, as you chose\n", p.Name)
		}
		return nil
	}
	in, _ := o.In.(*bufio.Reader)
	res, err := projecttooling.Apply(context.Background(), projecttooling.Input{Root: root, Dir: p.Folder, Answer: projecttooling.Smith}, projecttooling.RunOptions{
		Out: o.Out, In: in, Interactive: operationplan.ShouldAsk(o.Env, o.Yes), Yes: o.Yes, Smith: convert.SmithCommand(),
		// The install covers the whole workspace: smith's tools go into the
		// root package.json, which smith install -p does not install.
		// The build and the tests are the import's own check.
		Verify: false,
	})
	var left []Left
	for _, l := range res.Left {
		problemText, remedy, _ := strings.Cut(l, ". ")
		left = append(left, Left{Problem: problemText, Remedy: remedy})
	}
	if err != nil {
		left = append(left, Left{Problem: "the change of tools of " + p.Name + " stopped: " + firstLine(err.Error()),
			Remedy: "fix the cause, or take the change back with git restore and git clean, then run smith doctor"})
	}
	return left
}

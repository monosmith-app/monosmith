package adoption

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/convert"
	"github.com/monosmith-app/monosmith/apps/smith/features/lintreport"
	"github.com/monosmith-app/monosmith/apps/smith/features/projecttooling"
	"github.com/monosmith-app/monosmith/apps/smith/features/removalplan"
)

// ToolingRun is what adopt --init settled about each project's tools, and
// who answers the removal questions.
type ToolingRun struct {
	// Answers are by package name (SettledInit.Tooling); a project without an
	// answer keeps its own tools.
	Answers map[string]projecttooling.Answer
	// In reads the answers; Interactive says a person answers; Yes removes
	// every group of old tools without asking.
	In          *bufio.Reader
	Interactive bool
	Yes         bool
}

// ApplyTooling gives each Node project whose answer is smith's tooling that
// tooling (projecttooling.Apply): smith's build and ESLint config first, then
// the old tools go, each group shown and asked. The install and the checks
// run once, after the last project.
//
// Parameters: r is the report of the adopted repository (after the move);
// t the answers; out gets the steps.
// Returns: the steps left to do, in words (a project whose change stopped is
// one of them); an error for a project that cannot be read.
func ApplyTooling(r Report, t ToolingRun, out io.Writer) ([]string, error) {
	var dirs []string
	for _, p := range r.Projects {
		if p.Language != "node" {
			continue
		}
		a, err := projecttooling.Assess(r.Root, p.Root)
		if err != nil {
			return nil, err
		}
		if t.Answers[a.Name] == projecttooling.Smith {
			dirs = append(dirs, p.Root)
		}
	}
	var left []string
	names := map[string]string{}
	for _, p := range r.Projects {
		names[p.Root] = p.Name
	}
	smith := convert.SmithCommand()
	for i, dir := range dirs {
		res, err := projecttooling.Apply(context.Background(), projecttooling.Input{Root: r.Root, Dir: dir, Answer: projecttooling.Smith}, projecttooling.RunOptions{
			Out: out, In: t.In, Interactive: t.Interactive, Yes: t.Yes,
			Smith: smith, Verify: true, SkipFinish: i < len(dirs)-1, WholeRoot: true,
		})
		left = append(left, res.Left...)
		if err != nil {
			// The change stays in the work tree, where git shows it and
			// takes it back; the rest of the adoption still finishes.
			left = append(left, fmt.Sprintf("the change of tools of %s stopped: %s. Fix the cause, or take the change back with git restore and git clean, then run smith doctor", dir, firstLine(err.Error())))
			continue
		}
		for _, f := range res.Findings {
			// What doctor finds is listed by DescribeOutcome.
			switch f.Step.Description {
			case removalplan.StepDoctor:
				continue
			case removalplan.StepLint:
				// lintReport below names what the lint finds, per project.
				continue
			}
			left = append(left, "the check \""+f.Step.Description+"\" failed after the change of tools: run "+strings.Join(append([]string{"smith"}, f.Step.Argv[1:]...), " ")+" to see why")
		}
	}
	for _, dir := range dirs {
		left = append(left, lintReport(r.Root, names[dir], out)...)
	}
	return left, nil
}

// lintReport runs the lint of one project without --fix, so the list names
// what smith's rules find in the code it took in: the count, what --fix
// repairs and the rules found most. The lint of the finish runs for every
// project and stops behind a failed test, so it cannot say this.
func lintReport(root, name string, out io.Writer) []string {
	if name == "" {
		return nil
	}
	var left []string
	for _, item := range lintreport.Check(root, name, out) {
		left = append(left, item.Problem+". "+strings.ReplaceAll(item.Remedy, "\n", ", "))
	}
	return left
}

// firstLine is the first line of a message.
func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
